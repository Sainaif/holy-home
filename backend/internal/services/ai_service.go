package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/sainaif/holy-home/internal/models"
	openai "github.com/sashabaranov/go-openai"
)

type AIService struct {
	client *openai.Client
}

func NewAIService() *AIService {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil
	}

	client := openai.NewClient(apiKey)
	return &AIService{
		client: client,
	}
}

// ParseBillFromText uses GPT-4 to extract structured data from OCR text
func (s *AIService) ParseBillFromText(ctx context.Context, ocrText string) (*models.AIExtraction, error) {
	if s.client == nil {
		return nil, fmt.Errorf("OpenAI client not initialized - check OPENAI_API_KEY environment variable")
	}

	systemPrompt := `You are a bill parsing assistant. Extract information from Polish utility bills.
Return a JSON object with the following fields:
- type: one of "electricity", "gas", "water", "internet", "inne" (other)
- totalAmount: total amount in PLN (number)
- units: total units consumed if available (number, optional)
- periodStart: billing period start date (ISO 8601 format)
- periodEnd: billing period end date (ISO 8601 format)
- deadline: payment deadline if available (ISO 8601 format, optional)
- notes: any additional notes (string, optional)

If you cannot extract a field with confidence, use null. For dates, use ISO 8601 format (YYYY-MM-DD).`

	userPrompt := fmt.Sprintf("Extract billing information from this OCR text:\n\n%s", ocrText)

	resp, err := s.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
			Temperature: 0.1,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	content := resp.Choices[0].Message.Content

	// Parse JSON response
	var rawResult map[string]interface{}
	if err := json.Unmarshal([]byte(content), &rawResult); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	// Convert to AIExtraction
	extraction := &models.AIExtraction{}

	// Type
	if t, ok := rawResult["type"].(string); ok {
		extraction.Type = t
	}

	// Total Amount
	if amount, ok := rawResult["totalAmount"].(float64); ok {
		extraction.TotalAmount = amount
	}

	// Units (optional)
	if units, ok := rawResult["units"].(float64); ok {
		extraction.Units = &units
	}

	// Period Start
	if periodStart, ok := rawResult["periodStart"].(string); ok {
		if t, err := time.Parse("2006-01-02", periodStart); err == nil {
			extraction.PeriodStart = t
		}
	}

	// Period End
	if periodEnd, ok := rawResult["periodEnd"].(string); ok {
		if t, err := time.Parse("2006-01-02", periodEnd); err == nil {
			extraction.PeriodEnd = t
		}
	}

	// Deadline (optional)
	if deadline, ok := rawResult["deadline"].(string); ok && deadline != "" {
		if t, err := time.Parse("2006-01-02", deadline); err == nil {
			extraction.Deadline = &t
		}
	}

	// Notes (optional)
	if notes, ok := rawResult["notes"].(string); ok && notes != "" {
		extraction.Notes = &notes
	}

	return extraction, nil
}

// CalculateConfidence calculates a confidence score for the AI extraction
func (s *AIService) CalculateConfidence(extraction *models.AIExtraction, ocrText string) float64 {
	score := 1.0

	// Check if AI found key data
	if extraction.TotalAmount == 0 {
		score -= 0.3
	}
	if extraction.PeriodStart.IsZero() || extraction.PeriodEnd.IsZero() {
		score -= 0.2
	}
	if extraction.Type == "inne" || extraction.Type == "" {
		score -= 0.15
	}

	// Check OCR quality
	if len(ocrText) < 50 {
		score -= 0.2
	}

	// Check for common keywords in OCR text
	keywords := []string{"faktura", "kwota", "PLN", "zł", "okres", "termin"}
	foundKeywords := 0
	lowerText := strings.ToLower(ocrText)
	for _, kw := range keywords {
		if strings.Contains(lowerText, kw) {
			foundKeywords++
		}
	}
	if foundKeywords < 2 {
		score -= 0.15
	}

	return math.Max(0, score)
}
