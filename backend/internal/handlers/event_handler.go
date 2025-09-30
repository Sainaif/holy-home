package handlers

import (
	"bufio"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sainaif/holy-home/internal/middleware"
	"github.com/sainaif/holy-home/internal/services"
	"github.com/valyala/fasthttp"
)

type EventHandler struct {
	eventService *services.EventService
}

func NewEventHandler(eventService *services.EventService) *EventHandler {
	return &EventHandler{eventService: eventService}
}

// StreamEvents handles SSE connections
func (h *EventHandler) StreamEvents(c *fiber.Ctx) error {
	// Get user ID from context
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Set SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Subscribe to events
	eventChan := h.eventService.Subscribe(userID)
	defer h.eventService.Unsubscribe(userID)

	// Stream events
	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		// Send initial connection event
		fmt.Fprintf(w, "data: {\"type\":\"connected\",\"timestamp\":\"%s\"}\n\n", time.Now().Format(time.RFC3339))
		w.Flush()

		// Keep connection alive with heartbeat
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case event, ok := <-eventChan:
				if !ok {
					// Channel closed
					return
				}

				// Send event to client
				formatted := event.FormatSSE()
				if _, err := fmt.Fprint(w, formatted); err != nil {
					return
				}
				w.Flush()

			case <-ticker.C:
				// Send heartbeat to keep connection alive
				if _, err := fmt.Fprintf(w, ": heartbeat\n\n"); err != nil {
					return
				}
				w.Flush()

			case <-c.Context().Done():
				// Client disconnected
				return
			}
		}
	}))

	return nil
}