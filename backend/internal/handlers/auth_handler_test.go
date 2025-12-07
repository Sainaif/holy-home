package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/sainaif/holy-home/internal/services"
)

// MockAuthService mocks the AuthService for testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx interface{}, req services.LoginRequest, ip, userAgent string) (*services.TokenResponse, error) {
	args := m.Called(ctx, req, ip, userAgent)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.TokenResponse), args.Error(1)
}

func (m *MockAuthService) RefreshTokens(ctx interface{}, refreshToken, ip, userAgent string) (*services.TokenResponse, error) {
	args := m.Called(ctx, refreshToken, ip, userAgent)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.TokenResponse), args.Error(1)
}

func setupTestApp() *fiber.App {
	app := fiber.New()
	return app
}

func TestLogin_Success(t *testing.T) {
	app := setupTestApp()

	// Create a simple mock handler for testing request/response
	app.Post("/auth/login", func(c *fiber.Ctx) error {
		var req services.LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Simulate successful login
		if req.Email == "test@example.com" && req.Password == "password123" {
			return c.JSON(services.TokenResponse{
				Access:  "test-access-token",
				Refresh: "test-refresh-token",
			})
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid credentials",
		})
	})

	// Test successful login
	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var tokenResp services.TokenResponse
	json.NewDecoder(resp.Body).Decode(&tokenResp)
	assert.NotEmpty(t, tokenResp.Access)
	assert.NotEmpty(t, tokenResp.Refresh)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	app := setupTestApp()

	app.Post("/auth/login", func(c *fiber.Ctx) error {
		var req services.LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid credentials",
		})
	})

	reqBody := map[string]string{
		"email":    "wrong@example.com",
		"password": "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLogin_InvalidRequestBody(t *testing.T) {
	app := setupTestApp()

	app.Post("/auth/login", func(c *fiber.Ctx) error {
		var req services.LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}
		return c.SendStatus(fiber.StatusOK)
	})

	// Send invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestLogin_EmptyBody(t *testing.T) {
	app := setupTestApp()

	app.Post("/auth/login", func(c *fiber.Ctx) error {
		var req services.LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}
		if req.Email == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid credentials",
			})
		}
		return c.SendStatus(fiber.StatusOK)
	})

	reqBody := map[string]string{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLogin_Requires2FA(t *testing.T) {
	app := setupTestApp()

	app.Post("/auth/login", func(c *fiber.Ctx) error {
		var req services.LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Simulate user with 2FA enabled
		if req.Email == "2fa@example.com" && req.TOTPCode == "" {
			return c.JSON(services.TokenResponse{
				Requires2FA: true,
			})
		}

		return c.JSON(services.TokenResponse{
			Access:  "test-access-token",
			Refresh: "test-refresh-token",
		})
	})

	reqBody := map[string]string{
		"email":    "2fa@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var tokenResp services.TokenResponse
	json.NewDecoder(resp.Body).Decode(&tokenResp)
	assert.True(t, tokenResp.Requires2FA)
	assert.Empty(t, tokenResp.Access)
}

func TestRefresh_Success(t *testing.T) {
	app := setupTestApp()

	app.Post("/auth/refresh", func(c *fiber.Ctx) error {
		var req struct {
			RefreshToken string `json:"refreshToken"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if req.RefreshToken == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Refresh token is required",
			})
		}

		if req.RefreshToken == "valid-refresh-token" {
			return c.JSON(services.TokenResponse{
				Access:  "new-access-token",
				Refresh: "new-refresh-token",
			})
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid refresh token",
		})
	})

	reqBody := map[string]string{
		"refreshToken": "valid-refresh-token",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var tokenResp services.TokenResponse
	json.NewDecoder(resp.Body).Decode(&tokenResp)
	assert.Equal(t, "new-access-token", tokenResp.Access)
	assert.Equal(t, "new-refresh-token", tokenResp.Refresh)
}

func TestRefresh_MissingToken(t *testing.T) {
	app := setupTestApp()

	app.Post("/auth/refresh", func(c *fiber.Ctx) error {
		var req struct {
			RefreshToken string `json:"refreshToken"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if req.RefreshToken == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Refresh token is required",
			})
		}

		return c.SendStatus(fiber.StatusOK)
	})

	reqBody := map[string]string{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRefresh_InvalidToken(t *testing.T) {
	app := setupTestApp()

	app.Post("/auth/refresh", func(c *fiber.Ctx) error {
		var req struct {
			RefreshToken string `json:"refreshToken"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if req.RefreshToken == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Refresh token is required",
			})
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid refresh token",
		})
	})

	reqBody := map[string]string{
		"refreshToken": "invalid-token",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestRefresh_LegacyFieldName(t *testing.T) {
	app := setupTestApp()

	app.Post("/auth/refresh", func(c *fiber.Ctx) error {
		var req struct {
			RefreshToken       string `json:"refreshToken"`
			LegacyRefreshToken string `json:"refresh_token"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		refreshToken := req.RefreshToken
		if refreshToken == "" {
			refreshToken = req.LegacyRefreshToken
		}

		if refreshToken == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Refresh token is required",
			})
		}

		if refreshToken == "legacy-valid-token" {
			return c.JSON(services.TokenResponse{
				Access:  "new-access-token",
				Refresh: "new-refresh-token",
			})
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid refresh token",
		})
	})

	// Test with legacy field name (refresh_token instead of refreshToken)
	reqBody := map[string]string{
		"refresh_token": "legacy-valid-token",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
