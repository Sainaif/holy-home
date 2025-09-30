package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/sainaif/holy-home/internal/config"
	"github.com/sainaif/holy-home/internal/models"
	"github.com/sainaif/holy-home/internal/utils"
)

type AuthService struct {
	db  *mongo.Database
	cfg *config.Config
}

func NewAuthService(db *mongo.Database, cfg *config.Config) *AuthService {
	return &AuthService{
		db:  db,
		cfg: cfg,
	}
}

type LoginRequest struct {
	Email    string  `json:"email"`
	Password string  `json:"password"`
	TOTP     *string `json:"totp,omitempty"`
}

type TokenResponse struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

// Login authenticates a user and returns JWT tokens
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*TokenResponse, error) {
	// Find user by email
	var user models.User
	err := s.db.Collection("users").FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("invalid credentials")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	if !user.IsActive {
		return nil, errors.New("user account is disabled")
	}

	// Verify password
	valid, err := utils.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("password verification error: %w", err)
	}
	if !valid {
		return nil, errors.New("invalid credentials")
	}

	// Check 2FA if enabled
	if s.cfg.Auth.TwoFAEnabled && user.TOTPSecret != nil {
		if req.TOTP == nil || !utils.ValidateTOTP(*req.TOTP, *user.TOTPSecret) {
			return nil, errors.New("invalid 2FA code")
		}
	}

	// Generate tokens
	accessToken, err := utils.GenerateAccessToken(
		user.ID,
		user.Email,
		user.Role,
		s.cfg.JWT.Secret,
		s.cfg.JWT.AccessTTL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken(
		user.ID,
		s.cfg.JWT.RefreshSecret,
		s.cfg.JWT.RefreshTTL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenResponse{
		Access:  accessToken,
		Refresh: refreshToken,
	}, nil
}

// RefreshTokens generates new tokens from a valid refresh token
func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	// Validate refresh token
	userID, err := utils.ValidateRefreshToken(refreshToken, s.cfg.JWT.RefreshSecret)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Find user
	var user models.User
	err = s.db.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	if !user.IsActive {
		return nil, errors.New("user account is disabled")
	}

	// Generate new tokens
	accessToken, err := utils.GenerateAccessToken(
		user.ID,
		user.Email,
		user.Role,
		s.cfg.JWT.Secret,
		s.cfg.JWT.AccessTTL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := utils.GenerateRefreshToken(
		user.ID,
		s.cfg.JWT.RefreshSecret,
		s.cfg.JWT.RefreshTTL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenResponse{
		Access:  accessToken,
		Refresh: newRefreshToken,
	}, nil
}

// Enable2FA generates a new TOTP secret for the user
func (s *AuthService) Enable2FA(ctx context.Context, userID primitive.ObjectID) (string, string, error) {
	// Get user
	var user models.User
	err := s.db.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return "", "", fmt.Errorf("user not found: %w", err)
	}

	// Generate TOTP secret
	secret, err := utils.GenerateTOTPSecret()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	// Update user with TOTP secret
	_, err = s.db.Collection("users").UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"totp_secret": secret}},
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to save TOTP secret: %w", err)
	}

	// Generate provisioning URL
	otpauthURL := utils.GenerateTOTPURL(secret, user.Email, s.cfg.App.Name)

	return secret, otpauthURL, nil
}

// Disable2FA removes the TOTP secret for the user
func (s *AuthService) Disable2FA(ctx context.Context, userID primitive.ObjectID) error {
	_, err := s.db.Collection("users").UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$unset": bson.M{"totp_secret": ""}},
	)
	if err != nil {
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}
	return nil
}

// BootstrapAdmin creates the admin user if it doesn't exist
func (s *AuthService) BootstrapAdmin(ctx context.Context) error {
	// Check if admin already exists
	count, err := s.db.Collection("users").CountDocuments(ctx, bson.M{"role": "ADMIN"})
	if err != nil {
		return fmt.Errorf("failed to check for existing admin: %w", err)
	}

	if count > 0 {
		// Admin already exists
		return nil
	}

	// Hash password if it's plain text (for development)
	passwordHash := s.cfg.Admin.PasswordHash
	if passwordHash != "" && len(passwordHash) < 50 {
		// Looks like plain text password, hash it
		hashed, err := utils.HashPassword(passwordHash)
		if err != nil {
			return fmt.Errorf("failed to hash admin password: %w", err)
		}
		passwordHash = hashed
	}

	// Create admin user from config
	admin := models.User{
		ID:           primitive.NewObjectID(),
		Email:        s.cfg.Admin.Email,
		PasswordHash: passwordHash,
		Role:         "ADMIN",
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	_, err = s.db.Collection("users").InsertOne(ctx, admin)
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	return nil
}