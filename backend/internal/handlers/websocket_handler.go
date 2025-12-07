package handlers

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/sainaif/holy-home/internal/config"
	"github.com/sainaif/holy-home/internal/services"
	"github.com/sainaif/holy-home/internal/utils"
)

type WebSocketHandler struct {
	eventService *services.EventService
	config       *config.Config
}

func NewWebSocketHandler(eventService *services.EventService, cfg *config.Config) *WebSocketHandler {
	return &WebSocketHandler{
		eventService: eventService,
		config:       cfg,
	}
}

// AuthMessage is the expected first message from client
type AuthMessage struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

// WSMessage is a generic WebSocket message
type WSMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// UpgradeMiddleware checks if the request is a WebSocket upgrade request
func (h *WebSocketHandler) UpgradeMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}

// HandleWebSocket handles WebSocket connections with post-connection auth
func (h *WebSocketHandler) HandleWebSocket() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		defer c.Close()

		// Set read deadline for auth message (10 seconds)
		c.SetReadDeadline(time.Now().Add(10 * time.Second))

		// Wait for auth message
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Printf("WebSocket: failed to read auth message: %v", err)
			c.WriteJSON(WSMessage{Type: "error", Data: json.RawMessage(`"Authentication timeout"`)})
			return
		}

		var authMsg AuthMessage
		if err := json.Unmarshal(msg, &authMsg); err != nil {
			log.Printf("WebSocket: invalid auth message format: %v", err)
			c.WriteJSON(WSMessage{Type: "error", Data: json.RawMessage(`"Invalid message format"`)})
			return
		}

		if authMsg.Type != "auth" || authMsg.Token == "" {
			c.WriteJSON(WSMessage{Type: "error", Data: json.RawMessage(`"Missing auth token"`)})
			return
		}

		// Validate token
		claims, err := utils.ValidateAccessToken(authMsg.Token, h.config.JWT.Secret)
		if err != nil {
			log.Printf("WebSocket: invalid token: %v", err)
			c.WriteJSON(WSMessage{Type: "error", Data: json.RawMessage(`"Invalid or expired token"`)})
			return
		}

		userID := claims.UserID

		// Clear read deadline for normal operation
		c.SetReadDeadline(time.Time{})

		// Subscribe to events
		eventChan := h.eventService.Subscribe(userID)
		defer h.eventService.Unsubscribe(userID)

		// Send auth success
		c.WriteJSON(WSMessage{Type: "authenticated"})

		// Create done channel for cleanup
		done := make(chan struct{})
		defer close(done)

		// Heartbeat ticker
		heartbeat := time.NewTicker(15 * time.Second)
		defer heartbeat.Stop()

		// Read pump - handles incoming messages (pings, etc)
		go func() {
			defer func() {
				select {
				case <-done:
				default:
				}
			}()
			for {
				select {
				case <-done:
					return
				default:
					_, _, err := c.ReadMessage()
					if err != nil {
						return
					}
					// We don't expect any messages from client after auth
					// but we need to read to detect disconnects
				}
			}
		}()

		// Write pump - sends events to client
		for {
			select {
			case event, ok := <-eventChan:
				if !ok {
					return
				}

				eventData, _ := json.Marshal(event)
				wsMsg := WSMessage{
					Type: "event",
					Data: eventData,
				}

				if err := c.WriteJSON(wsMsg); err != nil {
					log.Printf("WebSocket: failed to write event: %v", err)
					return
				}

			case <-heartbeat.C:
				if err := c.WriteJSON(WSMessage{Type: "heartbeat"}); err != nil {
					log.Printf("WebSocket: failed to write heartbeat: %v", err)
					return
				}

			case <-done:
				return
			}
		}
	})
}
