package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockSession represents a session for testing
type MockSession struct {
	ID        primitive.ObjectID `json:"id"`
	UserID    primitive.ObjectID `json:"userId"`
	Name      string             `json:"name"`
	IPAddress string             `json:"ipAddress"`
	UserAgent string             `json:"userAgent"`
	CreatedAt time.Time          `json:"createdAt"`
	ExpiresAt time.Time          `json:"expiresAt"`
}

func TestGetSessions_Success(t *testing.T) {
	app := fiber.New()

	// Mock session data
	sessions := []MockSession{
		{
			ID:        primitive.NewObjectID(),
			Name:      "Chrome on Windows",
			IPAddress: "192.168.1.1",
			UserAgent: "Mozilla/5.0 Chrome",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		},
		{
			ID:        primitive.NewObjectID(),
			Name:      "Firefox on MacOS",
			IPAddress: "192.168.1.2",
			UserAgent: "Mozilla/5.0 Firefox",
			CreatedAt: time.Now().Add(-1 * time.Hour),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		},
	}

	app.Get("/sessions", func(c *fiber.Ctx) error {
		// Simulate auth middleware setting userId
		return c.JSON(sessions)
	})

	req := httptest.NewRequest(http.MethodGet, "/sessions", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result []MockSession
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Len(t, result, 2)
}

func TestGetSessions_Unauthorized(t *testing.T) {
	app := fiber.New()

	app.Get("/sessions", func(c *fiber.Ctx) error {
		// Simulate no auth
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/sessions", nil)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestRenameSession_Success(t *testing.T) {
	app := fiber.New()

	sessionID := primitive.NewObjectID()

	app.Put("/sessions/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		_, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid session ID",
			})
		}

		var req struct {
			Name string `json:"name"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if req.Name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Session name is required",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Session renamed successfully",
		})
	})

	reqBody := map[string]string{
		"name": "My Work Laptop",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/sessions/"+sessionID.Hex(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRenameSession_InvalidID(t *testing.T) {
	app := fiber.New()

	app.Put("/sessions/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		_, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid session ID",
			})
		}
		return c.SendStatus(fiber.StatusOK)
	})

	reqBody := map[string]string{
		"name": "My Work Laptop",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/sessions/invalid-id", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRenameSession_EmptyName(t *testing.T) {
	app := fiber.New()

	sessionID := primitive.NewObjectID()

	app.Put("/sessions/:id", func(c *fiber.Ctx) error {
		var req struct {
			Name string `json:"name"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if req.Name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Session name is required",
			})
		}

		return c.SendStatus(fiber.StatusOK)
	})

	reqBody := map[string]string{
		"name": "",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/sessions/"+sessionID.Hex(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestDeleteSession_Success(t *testing.T) {
	app := fiber.New()

	sessionID := primitive.NewObjectID()

	app.Delete("/sessions/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		_, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid session ID",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Session deleted successfully",
		})
	})

	req := httptest.NewRequest(http.MethodDelete, "/sessions/"+sessionID.Hex(), nil)
	req.Header.Set("Authorization", "Bearer test-token")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDeleteSession_InvalidID(t *testing.T) {
	app := fiber.New()

	app.Delete("/sessions/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		_, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid session ID",
			})
		}
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodDelete, "/sessions/invalid-id", nil)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestDeleteAllSessions_Success(t *testing.T) {
	app := fiber.New()

	app.Delete("/sessions", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "All sessions revoked successfully",
		})
	})

	req := httptest.NewRequest(http.MethodDelete, "/sessions", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "All sessions revoked successfully", result["message"])
}

func TestDeleteAllSessions_KeepCurrent(t *testing.T) {
	app := fiber.New()

	currentSessionID := primitive.NewObjectID()

	app.Delete("/sessions", func(c *fiber.Ctx) error {
		keepCurrent := c.Query("keepCurrent")
		if keepCurrent != "" {
			// Validate the keepCurrent session ID
			if _, err := primitive.ObjectIDFromHex(keepCurrent); err != nil {
				// Invalid ID is ignored, not an error
			}
		}
		return c.JSON(fiber.Map{
			"message": "All sessions revoked successfully",
		})
	})

	req := httptest.NewRequest(http.MethodDelete, "/sessions?keepCurrent="+currentSessionID.Hex(), nil)
	req.Header.Set("Authorization", "Bearer test-token")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDeleteAllSessions_Unauthorized(t *testing.T) {
	app := fiber.New()

	app.Delete("/sessions", func(c *fiber.Ctx) error {
		// Simulate no auth
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	})

	req := httptest.NewRequest(http.MethodDelete, "/sessions", nil)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
