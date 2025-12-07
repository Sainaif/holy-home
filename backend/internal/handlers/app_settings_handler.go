package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sainaif/holy-home/internal/services"
)

type AppSettingsHandler struct {
	appSettingsService *services.AppSettingsService
}

func NewAppSettingsHandler(appSettingsService *services.AppSettingsService) *AppSettingsHandler {
	return &AppSettingsHandler{
		appSettingsService: appSettingsService,
	}
}

// GetSettings retrieves app settings (PUBLIC - no auth required for branding)
func (h *AppSettingsHandler) GetSettings(c *fiber.Ctx) error {
	settings, err := h.appSettingsService.GetSettings(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(settings)
}

// UpdateSettings updates app settings (ADMIN only)
func (h *AppSettingsHandler) UpdateSettings(c *fiber.Ctx) error {
	var req services.UpdateSettingsInput

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.appSettingsService.UpdateSettings(c.Context(), req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "App settings updated successfully",
	})
}

// GetSupportedLanguages returns the list of supported languages
func (h *AppSettingsHandler) GetSupportedLanguages(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"languages": services.SupportedLanguages,
	})
}
