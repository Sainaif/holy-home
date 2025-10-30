package services

import (
	"context"
	"fmt"
	"time"

	"github.com/sainaif/holy-home/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type BillScanService struct {
	db         *mongo.Database
	ocrService *OCRService
	aiService  *AIService
}

func NewBillScanService(db *mongo.Database, ocrService *OCRService, aiService *AIService) *BillScanService {
	return &BillScanService{
		db:         db,
		ocrService: ocrService,
		aiService:  aiService,
	}
}

// CreateScan creates a new bill scan record in the database
func (s *BillScanService) CreateScan(ctx context.Context, userID primitive.ObjectID, imagePath string) (*models.BillScan, error) {
	scan := &models.BillScan{
		ID:         primitive.NewObjectID(),
		UploadedBy: userID,
		ImagePath:  imagePath,
		Status:     "processing",
		CreatedAt:  time.Now(),
	}

	_, err := s.db.Collection("bill_scans").InsertOne(ctx, scan)
	if err != nil {
		return nil, fmt.Errorf("failed to create scan record: %w", err)
	}

	return scan, nil
}

// ProcessScan performs OCR and AI extraction on a bill scan
func (s *BillScanService) ProcessScan(ctx context.Context, scanID primitive.ObjectID) error {
	// Get the scan record
	var scan models.BillScan
	err := s.db.Collection("bill_scans").FindOne(ctx, bson.M{"_id": scanID}).Decode(&scan)
	if err != nil {
		return fmt.Errorf("failed to find scan: %w", err)
	}

	// Perform OCR
	ocrText, err := s.ocrService.ExtractTextFromImage(scan.ImagePath)
	if err != nil {
		// Update scan status to failed
		errMsg := err.Error()
		_, updateErr := s.db.Collection("bill_scans").UpdateOne(
			ctx,
			bson.M{"_id": scanID},
			bson.M{
				"$set": bson.M{
					"status":        "failed",
					"error_message": errMsg,
					"processed_at":  time.Now(),
				},
			},
		)
		if updateErr != nil {
			return fmt.Errorf("OCR failed and failed to update status: %v (original error: %w)", updateErr, err)
		}
		return fmt.Errorf("OCR failed: %w", err)
	}

	// Perform AI extraction if AI service is available
	var aiExtraction *models.AIExtraction
	var confidence float64

	if s.aiService != nil {
		aiExtraction, err = s.aiService.ParseBillFromText(ctx, ocrText)
		if err != nil {
			// Log error but don't fail - we still have OCR text
			errMsg := fmt.Sprintf("AI extraction failed: %v", err)
			_, _ = s.db.Collection("bill_scans").UpdateOne(
				ctx,
				bson.M{"_id": scanID},
				bson.M{
					"$set": bson.M{
						"ocr_text":      ocrText,
						"error_message": errMsg,
						"confidence":    0.3,
						"status":        "completed",
						"processed_at":  time.Now(),
					},
				},
			)
			return fmt.Errorf("AI extraction failed: %w", err)
		}

		// Calculate confidence
		confidence = s.aiService.CalculateConfidence(aiExtraction, ocrText)
	} else {
		// No AI service - set low confidence
		confidence = 0.3
	}

	// Update scan with results
	update := bson.M{
		"$set": bson.M{
			"ocr_text":      ocrText,
			"ai_extraction": aiExtraction,
			"confidence":    confidence,
			"status":        "completed",
			"processed_at":  time.Now(),
		},
	}

	_, err = s.db.Collection("bill_scans").UpdateOne(ctx, bson.M{"_id": scanID}, update)
	if err != nil {
		return fmt.Errorf("failed to update scan with results: %w", err)
	}

	return nil
}

// GetScan retrieves a bill scan by ID
func (s *BillScanService) GetScan(ctx context.Context, scanID primitive.ObjectID) (*models.BillScan, error) {
	var scan models.BillScan
	err := s.db.Collection("bill_scans").FindOne(ctx, bson.M{"_id": scanID}).Decode(&scan)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("scan not found")
		}
		return nil, fmt.Errorf("failed to get scan: %w", err)
	}
	return &scan, nil
}

// GetUserScans retrieves all scans uploaded by a user
func (s *BillScanService) GetUserScans(ctx context.Context, userID primitive.ObjectID) ([]models.BillScan, error) {
	cursor, err := s.db.Collection("bill_scans").Find(ctx, bson.M{"uploaded_by": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to query scans: %w", err)
	}
	defer cursor.Close(ctx)

	var scans []models.BillScan
	if err := cursor.All(ctx, &scans); err != nil {
		return nil, fmt.Errorf("failed to decode scans: %w", err)
	}

	return scans, nil
}
