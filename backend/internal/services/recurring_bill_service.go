package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sainaif/holy-home/internal/config"
	"github.com/sainaif/holy-home/internal/models"
	"github.com/sainaif/holy-home/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RecurringBillService struct {
	db  *mongo.Database
	cfg *config.Config
}

func NewRecurringBillService(db *mongo.Database, cfg *config.Config) *RecurringBillService {
	return &RecurringBillService{
		db:  db,
		cfg: cfg,
	}
}

// CreateTemplate creates a new recurring bill template
func (s *RecurringBillService) CreateTemplate(ctx context.Context, template *models.RecurringBillTemplate) error {
	// Validate allocations
	if err := s.validateAllocations(template.Allocations); err != nil {
		return err
	}

	// Set timestamps
	now := time.Now()
	template.CreatedAt = now
	template.UpdatedAt = now
	template.IsActive = true

	// Calculate next due date based on day of month
	template.NextDueDate = s.calculateNextDueDate(now, template.DayOfMonth, template.Frequency)

	result, err := s.db.Collection("recurring_bill_templates").InsertOne(ctx, template)
	if err != nil {
		return err
	}

	template.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetTemplate retrieves a recurring bill template by ID
func (s *RecurringBillService) GetTemplate(ctx context.Context, id primitive.ObjectID) (*models.RecurringBillTemplate, error) {
	var template models.RecurringBillTemplate
	err := s.db.Collection("recurring_bill_templates").FindOne(ctx, bson.M{"_id": id}).Decode(&template)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("template not found")
		}
		return nil, err
	}
	return &template, nil
}

// ListTemplates retrieves all recurring bill templates
func (s *RecurringBillService) ListTemplates(ctx context.Context) ([]models.RecurringBillTemplate, error) {
	cursor, err := s.db.Collection("recurring_bill_templates").Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var templates []models.RecurringBillTemplate
	if err := cursor.All(ctx, &templates); err != nil {
		return nil, err
	}

	return templates, nil
}

// UpdateTemplate updates an existing template
func (s *RecurringBillService) UpdateTemplate(ctx context.Context, id primitive.ObjectID, updates map[string]interface{}) error {
	// Validate allocations if being updated
	if allocations, ok := updates["allocations"].([]models.RecurringBillAllocation); ok {
		if err := s.validateAllocations(allocations); err != nil {
			return err
		}
	}

	updates["updated_at"] = time.Now()

	result, err := s.db.Collection("recurring_bill_templates").UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("template not found")
	}

	return nil
}

// DeleteTemplate deletes a template (soft delete by setting IsActive to false)
func (s *RecurringBillService) DeleteTemplate(ctx context.Context, id primitive.ObjectID) error {
	result, err := s.db.Collection("recurring_bill_templates").UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"is_active": false, "updated_at": time.Now()}},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("template not found")
	}

	return nil
}

// GenerateBillsFromTemplates generates bills from all active templates that are due
func (s *RecurringBillService) GenerateBillsFromTemplates(ctx context.Context) error {
	now := time.Now()

	// Find all active templates where next_due_date <= now
	cursor, err := s.db.Collection("recurring_bill_templates").Find(ctx, bson.M{
		"is_active": true,
		"next_due_date": bson.M{"$lte": now},
	})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var templates []models.RecurringBillTemplate
	if err := cursor.All(ctx, &templates); err != nil {
		return err
	}

	// Generate bills for each template
	for _, template := range templates {
		if err := s.generateBillFromTemplate(ctx, &template); err != nil {
			fmt.Printf("Error generating bill from template %s: %v\n", template.ID.Hex(), err)
			continue
		}
	}

	return nil
}

// generateBillFromTemplate generates a single bill from a template
func (s *RecurringBillService) generateBillFromTemplate(ctx context.Context, template *models.RecurringBillTemplate) error {
	now := time.Now()

	// Calculate period based on frequency
	periodStart, periodEnd := s.calculatePeriod(template.NextDueDate, template.Frequency)

	// Create the bill
	allocationType := "simple"
	bill := &models.Bill{
		Type:                "inne",
		CustomType:          &template.CustomType,
		AllocationType:      &allocationType,
		PeriodStart:         periodStart,
		PeriodEnd:           periodEnd,
		PaymentDeadline:     &template.NextDueDate,
		TotalAmountPLN:      template.Amount,
		Notes:               template.Notes,
		Status:              "draft", // Start as draft so it's modifiable
		RecurringTemplateID: &template.ID,
		CreatedAt:           now,
	}

	// Insert the bill
	result, err := s.db.Collection("bills").InsertOne(ctx, bill)
	if err != nil {
		return err
	}
	bill.ID = result.InsertedID.(primitive.ObjectID)

	// Create allocations based on template
	for _, allocTemplate := range template.Allocations {
		var allocatedAmount primitive.Decimal128

		switch allocTemplate.AllocationType {
		case "fixed":
			allocatedAmount = *allocTemplate.FixedAmount
		case "percentage":
			amountFloat, _ := utils.DecimalToFloat(template.Amount)
			percentage := *allocTemplate.Percentage / 100.0
			allocatedFloat := amountFloat * percentage
			roundedAmount := utils.RoundPLN(allocatedFloat)
			allocatedAmount, _ = utils.DecimalFromFloat(roundedAmount)
		case "fraction":
			amountFloat, _ := utils.DecimalToFloat(template.Amount)
			fraction := float64(*allocTemplate.FractionNum) / float64(*allocTemplate.FractionDenom)
			allocatedFloat := amountFloat * fraction
			roundedAmount := utils.RoundPLN(allocatedFloat)
			allocatedAmount, _ = utils.DecimalFromFloat(roundedAmount)
		}

		allocation := bson.M{
			"bill_id":       bill.ID,
			"subject_type":  allocTemplate.SubjectType,
			"subject_id":    allocTemplate.SubjectID,
			"allocated_pln": allocatedAmount,
			"created_at":    now,
		}

		if _, err := s.db.Collection("allocations").InsertOne(ctx, allocation); err != nil {
			return fmt.Errorf("failed to create allocation: %w", err)
		}
	}

	// Update template's next due date and last generated timestamp
	nextDueDate := s.calculateNextDueDate(template.NextDueDate, template.DayOfMonth, template.Frequency)
	_, err = s.db.Collection("recurring_bill_templates").UpdateOne(
		ctx,
		bson.M{"_id": template.ID},
		bson.M{"$set": bson.M{
			"next_due_date":      nextDueDate,
			"last_generated_at":  now,
			"updated_at":         now,
		}},
	)

	return err
}

// calculateNextDueDate calculates the next due date based on frequency
func (s *RecurringBillService) calculateNextDueDate(from time.Time, dayOfMonth int, frequency string) time.Time {
	var next time.Time

	switch frequency {
	case "monthly":
		next = from.AddDate(0, 1, 0)
	case "quarterly":
		next = from.AddDate(0, 3, 0)
	case "yearly":
		next = from.AddDate(1, 0, 0)
	default:
		next = from.AddDate(0, 1, 0)
	}

	// Adjust to the specified day of month
	year, month, _ := next.Date()
	next = time.Date(year, month, dayOfMonth, 0, 0, 0, 0, next.Location())

	return next
}

// calculatePeriod calculates the billing period for a given due date
func (s *RecurringBillService) calculatePeriod(dueDate time.Time, frequency string) (time.Time, time.Time) {
	var periodStart time.Time

	switch frequency {
	case "monthly":
		periodStart = dueDate.AddDate(0, -1, 0)
	case "quarterly":
		periodStart = dueDate.AddDate(0, -3, 0)
	case "yearly":
		periodStart = dueDate.AddDate(-1, 0, 0)
	default:
		periodStart = dueDate.AddDate(0, -1, 0)
	}

	periodEnd := dueDate

	return periodStart, periodEnd
}

// validateAllocations validates that allocations are properly configured
func (s *RecurringBillService) validateAllocations(allocations []models.RecurringBillAllocation) error {
	if len(allocations) == 0 {
		return errors.New("at least one allocation is required")
	}

	for i, alloc := range allocations {
		switch alloc.AllocationType {
		case "percentage":
			if alloc.Percentage == nil {
				return fmt.Errorf("allocation %d: percentage is required for percentage type", i+1)
			}
			if *alloc.Percentage <= 0 || *alloc.Percentage > 100 {
				return fmt.Errorf("allocation %d: percentage must be between 0 and 100", i+1)
			}
		case "fraction":
			if alloc.FractionNum == nil || alloc.FractionDenom == nil {
				return fmt.Errorf("allocation %d: fraction numerator and denominator are required for fraction type", i+1)
			}
			if *alloc.FractionNum <= 0 || *alloc.FractionDenom <= 0 {
				return fmt.Errorf("allocation %d: fraction values must be positive", i+1)
			}
			if *alloc.FractionNum > *alloc.FractionDenom {
				return fmt.Errorf("allocation %d: fraction numerator cannot be greater than denominator", i+1)
			}
		case "fixed":
			if alloc.FixedAmount == nil {
				return fmt.Errorf("allocation %d: fixed amount is required for fixed type", i+1)
			}
		default:
			return fmt.Errorf("allocation %d: invalid allocation type '%s'", i+1, alloc.AllocationType)
		}
	}

	// Validate that allocations sum to 100% (for percentage/fraction types)
	totalFraction := 0.0
	hasNonFixed := false

	for _, alloc := range allocations {
		switch alloc.AllocationType {
		case "percentage":
			totalFraction += *alloc.Percentage / 100.0
			hasNonFixed = true
		case "fraction":
			totalFraction += float64(*alloc.FractionNum) / float64(*alloc.FractionDenom)
			hasNonFixed = true
		}
	}

	if hasNonFixed && (totalFraction < 0.999 || totalFraction > 1.001) {
		return fmt.Errorf("allocations must sum to 100%% (currently %.2f%%)", totalFraction*100)
	}

	return nil
}
