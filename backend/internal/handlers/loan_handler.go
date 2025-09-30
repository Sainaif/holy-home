package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sainaif/holy-home/internal/middleware"
	"github.com/sainaif/holy-home/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LoanHandler struct {
	loanService *services.LoanService
}

func NewLoanHandler(loanService *services.LoanService) *LoanHandler {
	return &LoanHandler{loanService: loanService}
}

// CreateLoan creates a new loan
func (h *LoanHandler) CreateLoan(c *fiber.Ctx) error {
	var req services.CreateLoanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	loan, err := h.loanService.CreateLoan(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(loan)
}

// CreateLoanPayment records a loan repayment
func (h *LoanHandler) CreateLoanPayment(c *fiber.Ctx) error {
	var req services.CreateLoanPaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	payment, err := h.loanService.CreateLoanPayment(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(payment)
}

// GetLoans retrieves all loans
func (h *LoanHandler) GetLoans(c *fiber.Ctx) error {
	loans, err := h.loanService.GetLoans(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(loans)
}

// GetBalances retrieves pairwise balances
func (h *LoanHandler) GetBalances(c *fiber.Ctx) error {
	balances, err := h.loanService.GetBalances(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(balances)
}

// GetMyBalance retrieves the current user's balance
func (h *LoanHandler) GetMyBalance(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	balance, err := h.loanService.GetUserBalance(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(balance)
}

// GetUserBalance retrieves a specific user's balance (ADMIN)
func (h *LoanHandler) GetUserBalance(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	balance, err := h.loanService.GetUserBalance(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(balance)
}