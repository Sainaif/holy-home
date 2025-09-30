package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sainaif/holy-home/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GroupHandler struct {
	groupService *services.GroupService
}

func NewGroupHandler(groupService *services.GroupService) *GroupHandler {
	return &GroupHandler{groupService: groupService}
}

// CreateGroup creates a new group (ADMIN only)
func (h *GroupHandler) CreateGroup(c *fiber.Ctx) error {
	var req services.CreateGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	group, err := h.groupService.CreateGroup(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(group)
}

// GetGroups retrieves all groups
func (h *GroupHandler) GetGroups(c *fiber.Ctx) error {
	groups, err := h.groupService.GetGroups(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(groups)
}

// GetGroup retrieves a specific group
func (h *GroupHandler) GetGroup(c *fiber.Ctx) error {
	id := c.Params("id")
	groupID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid group ID",
		})
	}

	group, err := h.groupService.GetGroup(c.Context(), groupID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(group)
}

// UpdateGroup updates a group (ADMIN only)
func (h *GroupHandler) UpdateGroup(c *fiber.Ctx) error {
	id := c.Params("id")
	groupID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid group ID",
		})
	}

	var req services.UpdateGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.groupService.UpdateGroup(c.Context(), groupID, req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Group updated successfully",
	})
}

// DeleteGroup deletes a group (ADMIN only)
func (h *GroupHandler) DeleteGroup(c *fiber.Ctx) error {
	id := c.Params("id")
	groupID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid group ID",
		})
	}

	if err := h.groupService.DeleteGroup(c.Context(), groupID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Group deleted successfully",
	})
}