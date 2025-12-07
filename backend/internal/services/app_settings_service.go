package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/sainaif/holy-home/internal/models"
)

type AppSettingsService struct {
	db *mongo.Database
}

func NewAppSettingsService(db *mongo.Database) *AppSettingsService {
	return &AppSettingsService{db: db}
}

// GetSettings retrieves app settings (creates default if not exists)
func (s *AppSettingsService) GetSettings(ctx context.Context) (*models.AppSettings, error) {
	var settings models.AppSettings
	err := s.db.Collection("app_settings").FindOne(ctx, bson.M{}).Decode(&settings)

	if err == mongo.ErrNoDocuments {
		// Create default settings
		settings = models.AppSettings{
			ID:                primitive.NewObjectID(),
			AppName:           "Holy Home",
			DefaultLanguage:   "en",
			DisableAutoDetect: false,
			UpdatedAt:         time.Now(),
		}

		_, err = s.db.Collection("app_settings").InsertOne(ctx, settings)
		if err != nil {
			return nil, fmt.Errorf("failed to create default app settings: %w", err)
		}

		return &settings, nil
	}

	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &settings, nil
}

// SupportedLanguages defines the list of supported locale codes
var SupportedLanguages = []string{"en", "pl"}

// IsLanguageSupported checks if a language code is supported
func IsLanguageSupported(lang string) bool {
	for _, l := range SupportedLanguages {
		if l == lang {
			return true
		}
	}
	return false
}

// UpdateSettingsInput holds the input for updating app settings
type UpdateSettingsInput struct {
	AppName           *string `json:"appName"`
	DefaultLanguage   *string `json:"defaultLanguage"`
	DisableAutoDetect *bool   `json:"disableAutoDetect"`
}

// UpdateSettings updates app settings (ADMIN only - enforced at handler)
func (s *AppSettingsService) UpdateSettings(ctx context.Context, input UpdateSettingsInput) error {
	updates := bson.M{
		"updated_at": time.Now(),
	}

	if input.AppName != nil {
		if *input.AppName == "" {
			return errors.New("app name cannot be empty")
		}
		updates["app_name"] = *input.AppName
	}

	if input.DefaultLanguage != nil {
		if !IsLanguageSupported(*input.DefaultLanguage) {
			return fmt.Errorf("unsupported language: %s", *input.DefaultLanguage)
		}
		updates["default_language"] = *input.DefaultLanguage
	}

	if input.DisableAutoDetect != nil {
		updates["disable_auto_detect"] = *input.DisableAutoDetect
	}

	update := bson.M{"$set": updates}

	result, err := s.db.Collection("app_settings").UpdateOne(ctx, bson.M{}, update)
	if err != nil {
		return fmt.Errorf("failed to update app settings: %w", err)
	}

	if result.MatchedCount == 0 {
		return errors.New("app settings not found")
	}

	return nil
}
