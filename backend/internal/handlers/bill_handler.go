package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sainaif/holy-home/internal/middleware"
	"github.com/sainaif/holy-home/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BillHandler struct {
	billService        *services.BillService
	consumptionService *services.ConsumptionService
}

func NewBillHandler(billService *services.BillService, consumptionService *services.ConsumptionService) *BillHandler {
	return &BillHandler{
		billService:        billService,
		consumptionService: consumptionService,
	}
}

// CreateBill creates a new bill (ADMIN only)
func (h *BillHandler) CreateBill(c *fiber.Ctx) error {
	var req services.CreateBillRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	bill, err := h.billService.CreateBill(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(bill)
}

// GetBills retrieves bills with optional filters
func (h *BillHandler) GetBills(c *fiber.Ctx) error {
	billType := c.Query("type")
	fromStr := c.Query("from")
	toStr := c.Query("to")

	var billTypePtr *string
	if billType != "" {
		billTypePtr = &billType
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

	bills, err := h.billService.GetBills(c.Context(), billTypePtr, fromPtr, toPtr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(bills)
}

// GetBill retrieves a specific bill
func (h *BillHandler) GetBill(c *fiber.Ctx) error {
	id := c.Params("id")
	billID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid bill ID",
		})
	}

	bill, err := h.billService.GetBill(c.Context(), billID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(bill)
}

// AllocateBill performs cost allocation (ADMIN only)
func (h *BillHandler) AllocateBill(c *fiber.Ctx) error {
	id := c.Params("id")
	billID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid bill ID",
		})
	}

	var req services.AllocateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.billService.AllocateBill(c.Context(), billID, req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Bill allocated successfully",
	})
}

// PostBill changes status to posted (ADMIN only)
func (h *BillHandler) PostBill(c *fiber.Ctx) error {
	id := c.Params("id")
	billID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid bill ID",
		})
	}

	if err := h.billService.PostBill(c.Context(), billID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Bill posted successfully",
	})
}

// CloseBill changes status to closed (ADMIN only)
func (h *BillHandler) CloseBill(c *fiber.Ctx) error {
	id := c.Params("id")
	billID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid bill ID",
		})
	}

	if err := h.billService.CloseBill(c.Context(), billID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Bill closed successfully",
	})
}

// CreateConsumption records a consumption reading
func (h *BillHandler) CreateConsumption(c *fiber.Ctx) error {
	var req services.CreateConsumptionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Determine source based on user role
	role, _ := middleware.GetUserRole(c)
	source := "user"
	if role == "ADMIN" {
		source = "admin"
	}

	consumption, err := h.consumptionService.CreateConsumption(c.Context(), req, source)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(consumption)
}

// GetConsumptions retrieves consumptions for a bill (or all consumptions if no billId)
func (h *BillHandler) GetConsumptions(c *fiber.Ctx) error {
	billIDStr := c.Query("billId")

	var billID *primitive.ObjectID
	if billIDStr != "" {
		id, err := primitive.ObjectIDFromHex(billIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid bill ID",
			})
		}
		billID = &id
	}

	consumptions, err := h.consumptionService.GetConsumptions(c.Context(), billID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(consumptions)
}

// GetAllocations retrieves allocations for a bill
func (h *BillHandler) GetAllocations(c *fiber.Ctx) error {
	billIDStr := c.Query("billId")
	if billIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "billId query parameter is required",
		})
	}

	billID, err := primitive.ObjectIDFromHex(billIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid bill ID",
		})
	}

	allocations, err := h.consumptionService.GetAllocations(c.Context(), billID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(allocations)
}