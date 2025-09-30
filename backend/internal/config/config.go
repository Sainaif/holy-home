package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	App       AppConfig
	JWT       JWTConfig
	Admin     AdminConfig
	Auth      AuthConfig
	Mongo     MongoConfig
	ML        MLConfig
	Logging   LogConfig
}

type AppConfig struct {
	Name    string
	Env     string
	Host    string
	Port    string
	BaseURL string
}

type JWTConfig struct {
	AccessTTL       time.Duration
	RefreshTTL      time.Duration
	Secret          string
	RefreshSecret   string
}

type AdminConfig struct {
	Email        string
	PasswordHash string
}

type AuthConfig struct {
	TwoFAEnabled bool
}

type MongoConfig struct {
	URI      string
	Database string
}

type MLConfig struct {
	BaseURL string
	Timeout time.Duration
}

type LogConfig struct {
	Level  string
	Format string
}

func Load() (*Config, error) {
	accessTTL, err := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}

	refreshTTL, err := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "720h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}

	mlTimeout, err := time.ParseDuration(getEnv("ML_TIMEOUT_SECONDS", "30") + "s")
	if err != nil {
		return nil, fmt.Errorf("invalid ML_TIMEOUT_SECONDS: %w", err)
	}

	return &Config{
		App: AppConfig{
			Name:    getEnv("APP_NAME", "Holy Home"),
			Env:     getEnv("APP_ENV", "production"),
			Host:    getEnv("APP_HOST", "0.0.0.0"),
			Port:    getEnv("APP_PORT", "8080"),
			BaseURL: getEnv("APP_BASE_URL", "http://localhost:8080"),
		},
		JWT: JWTConfig{
			AccessTTL:     accessTTL,
			RefreshTTL:    refreshTTL,
			Secret:        getEnv("JWT_SECRET", ""),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", ""),
		},
		Admin: AdminConfig{
			Email:        getEnv("ADMIN_EMAIL", ""),
			PasswordHash: getEnv("ADMIN_PASSWORD_HASH", getEnv("ADMIN_PASSWORD", "")),
		},
		Auth: AuthConfig{
			TwoFAEnabled: getEnv("AUTH_2FA_ENABLED", "false") == "true",
		},
		Mongo: MongoConfig{
			URI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGO_DB", "holyhome"),
		},
		ML: MLConfig{
			BaseURL: getEnv("ML_BASE_URL", "http://localhost:8000"),
			Timeout: mlTimeout,
		},
		Logging: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}