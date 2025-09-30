package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sainaif/holy-home/internal/services"
)

type PredictionHandler struct {
	predictionService *services.PredictionService
}

func NewPredictionHandler(predictionService *services.PredictionService) *PredictionHandler {
	return &PredictionHandler{predictionService: predictionService}
}

// RecomputePrediction generates a new prediction (ADMIN only)
func (h *PredictionHandler) RecomputePrediction(c *fiber.Ctx) error {
	var req services.RecomputeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	prediction, err := h.predictionService.RecomputePrediction(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(prediction)
}

// GetPredictions retrieves predictions with optional filters
func (h *PredictionHandler) GetPredictions(c *fiber.Ctx) error {
	target := c.Query("target")
	fromStr := c.Query("from")
	toStr := c.Query("to")

	var targetPtr *string
	if target != "" {
		targetPtr = &target
	}

	var fromPtr, toPtr *time.Time
	if fromStr != "" {
		from, err := time.Parse(time.RFC3339, fromStr)
		if err == nil {
			fromPtr = &from
		}
	}
	if toStr != "" {
		to, err := time.Parse(time.RFC3339, toStr)
		if err == nil {
			toPtr = &to
		}
	}

	predictions, err := h.predictionService.GetPredictions(c.Context(), targetPtr, fromPtr, toPtr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(predictions)
}