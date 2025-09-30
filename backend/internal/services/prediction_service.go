package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/sainaif/holy-home/internal/config"
	"github.com/sainaif/holy-home/internal/models"
	"github.com/sainaif/holy-home/internal/utils"
)

type PredictionService struct {
	db     *mongo.Database
	cfg    *config.Config
	client *http.Client
}

func NewPredictionService(db *mongo.Database, cfg *config.Config) *PredictionService {
	return &PredictionService{
		db:  db,
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.ML.Timeout,
		},
	}
}

type MLForecastRequest struct {
	Target            string    `json:"target"`
	HistoricalDates   []string  `json:"historical_dates"`
	HistoricalValues  []float64 `json:"historical_values"`
	HorizonMonths     int       `json:"horizon_months"`
	ConfidenceLevel   float64   `json:"confidence_level"`
	CostPerUnit       *float64  `json:"cost_per_unit,omitempty"`
}

type MLModelInfo struct {
	Name       string                 `json:"name"`
	Version    string                 `json:"version"`
	Parameters map[string]interface{} `json:"parameters"`
	FitStats   map[string]interface{} `json:"fit_stats"`
}

type MLConfidenceInterval struct {
	Lower []float64 `json:"lower"`
	Upper []float64 `json:"upper"`
}

type MLForecastResponse struct {
	Target             string               `json:"target"`
	Model              MLModelInfo          `json:"model"`
	PredictedDates     []string             `json:"predicted_dates"`
	PredictedValues    []float64            `json:"predicted_values"`
	ConfidenceInterval MLConfidenceInterval `json:"confidence_interval"`
	PredictedCosts     []float64            `json:"predicted_costs,omitempty"`
	CreatedAt          string               `json:"created_at"`
}

type RecomputeRequest struct {
	Target  string `json:"target"` // electricity, gas, shared_budget
	Horizon int    `json:"horizon"`
}

// RecomputePrediction generates a new prediction for a target
func (s *PredictionService) RecomputePrediction(ctx context.Context, req RecomputeRequest) (*models.Prediction, error) {
	// Validate target
	validTargets := map[string]bool{"electricity": true, "gas": true, "shared_budget": true}
	if !validTargets[req.Target] {
		return nil, errors.New("invalid target, must be electricity, gas, or shared_budget")
	}

	horizon := req.Horizon
	if horizon == 0 {
		horizon = 3 // Default horizon
	}

	// Get historical data based on target
	var historicalDates []string
	var historicalValues []float64
	var costPerUnit *float64
	var createdFrom string

	switch req.Target {
	case "electricity", "gas":
		dates, values, cost, err := s.getUtilityHistory(ctx, req.Target)
		if err != nil {
			return nil, err
		}
		historicalDates = dates
		historicalValues = values
		costPerUnit = &cost
		createdFrom = "bills"

	case "shared_budget":
		dates, values, err := s.getSharedBudgetHistory(ctx)
		if err != nil {
			return nil, err
		}
		historicalDates = dates
		historicalValues = values
		createdFrom = "bills"
	}

	if len(historicalValues) < 3 {
		return nil, errors.New("insufficient historical data for forecasting (need at least 3 data points)")
	}

	// Call ML service
	mlRequest := MLForecastRequest{
		Target:            req.Target,
		HistoricalDates:   historicalDates,
		HistoricalValues:  historicalValues,
		HorizonMonths:     horizon,
		ConfidenceLevel:   0.95,
		CostPerUnit:       costPerUnit,
	}

	mlResponse, err := s.callMLService(ctx, mlRequest)
	if err != nil {
		return nil, fmt.Errorf("ML service error: %w", err)
	}

	// Parse predicted dates
	periodStart, _ := time.Parse(time.RFC3339, mlResponse.PredictedDates[0])
	periodEnd, _ := time.Parse(time.RFC3339, mlResponse.PredictedDates[len(mlResponse.PredictedDates)-1])

	// Calculate total predicted units and cost
	totalUnits := 0.0
	for _, v := range mlResponse.PredictedValues {
		totalUnits += v
	}

	totalCost := 0.0
	if len(mlResponse.PredictedCosts) > 0 {
		for _, c := range mlResponse.PredictedCosts {
			totalCost += c
		}
	}

	// Store prediction in database
	unitsDec, _ := utils.DecimalFromFloat(totalUnits)
	costDec, _ := utils.DecimalFromFloat(totalCost)

	prediction := models.Prediction{
		ID:                 primitive.NewObjectID(),
		Target:             req.Target,
		PeriodStart:        periodStart,
		PeriodEnd:          periodEnd,
		HorizonMonths:      horizon,
		PredictedUnits:     unitsDec,
		PredictedAmountPLN: costDec,
		Model: models.ModelInfo{
			Name:    mlResponse.Model.Name,
			Version: mlResponse.Model.Version,
		},
		CreatedFrom: createdFrom,
		CreatedAt:   time.Now(),
	}

	_, err = s.db.Collection("predictions").InsertOne(ctx, prediction)
	if err != nil {
		return nil, fmt.Errorf("failed to store prediction: %w", err)
	}

	return &prediction, nil
}

// GetPredictions retrieves predictions with optional filters
func (s *PredictionService) GetPredictions(ctx context.Context, target *string, from *time.Time, to *time.Time) ([]models.Prediction, error) {
	filter := bson.M{}

	if target != nil {
		filter["target"] = *target
	}

	if from != nil || to != nil {
		dateFilter := bson.M{}
		if from != nil {
			dateFilter["$gte"] = *from
		}
		if to != nil {
			dateFilter["$lte"] = *to
		}
		filter["period_start"] = dateFilter
	}

	cursor, err := s.db.Collection("predictions").Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer cursor.Close(ctx)

	var predictions []models.Prediction
	if err := cursor.All(ctx, &predictions); err != nil {
		return nil, fmt.Errorf("failed to decode predictions: %w", err)
	}

	return predictions, nil
}

// Helper functions

func (s *PredictionService) getUtilityHistory(ctx context.Context, utilityType string) ([]string, []float64, float64, error) {
	// Get bills of the specified type, sorted by period_start
	cursor, err := s.db.Collection("bills").Find(
		ctx,
		bson.M{"type": utilityType, "status": bson.M{"$in": []string{"posted", "closed"}}},
		options.Find().SetSort(bson.M{"period_start": 1}),
	)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("database error: %w", err)
	}
	defer cursor.Close(ctx)

	var bills []models.Bill
	if err := cursor.All(ctx, &bills); err != nil {
		return nil, nil, 0, fmt.Errorf("failed to decode bills: %w", err)
	}

	if len(bills) == 0 {
		return nil, nil, 0, errors.New("no historical bills found")
	}

	dates := make([]string, len(bills))
	values := make([]float64, len(bills))
	var totalAmount, totalUnits float64

	for i, bill := range bills {
		dates[i] = bill.PeriodStart.Format(time.RFC3339)
		units, _ := utils.DecimalToFloat(bill.TotalUnits)
		values[i] = units

		amount, _ := utils.DecimalToFloat(bill.TotalAmountPLN)
		totalAmount += amount
		totalUnits += units
	}

	// Calculate average cost per unit
	costPerUnit := 0.0
	if totalUnits > 0 {
		costPerUnit = totalAmount / totalUnits
	}

	return dates, values, costPerUnit, nil
}

func (s *PredictionService) getSharedBudgetHistory(ctx context.Context) ([]string, []float64, error) {
	cursor, err := s.db.Collection("bills").Find(
		ctx,
		bson.M{"type": "shared", "status": bson.M{"$in": []string{"posted", "closed"}}},
		options.Find().SetSort(bson.M{"period_start": 1}),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("database error: %w", err)
	}
	defer cursor.Close(ctx)

	var bills []models.Bill
	if err := cursor.All(ctx, &bills); err != nil {
		return nil, nil, fmt.Errorf("failed to decode bills: %w", err)
	}

	if len(bills) == 0 {
		return nil, nil, errors.New("no historical shared budget bills found")
	}

	dates := make([]string, len(bills))
	values := make([]float64, len(bills))

	for i, bill := range bills {
		dates[i] = bill.PeriodStart.Format(time.RFC3339)
		amount, _ := utils.DecimalToFloat(bill.TotalAmountPLN)
		values[i] = amount
	}

	return dates, values, nil
}

func (s *PredictionService) callMLService(ctx context.Context, req MLForecastRequest) (*MLForecastResponse, error) {
	// Serialize request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/forecast", s.cfg.ML.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML service returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var mlResp MLForecastResponse
	if err := json.Unmarshal(respBody, &mlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &mlResp, nil
}