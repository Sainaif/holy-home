package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sainaif/holy-home/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BillScanHandler struct {
	scanService *services.BillScanService
}

func NewBillScanHandler(scanService *services.BillScanService) *BillScanHandler {
	return &BillScanHandler{
		scanService: scanService,
	}
}

// UploadBillScan handles uploading a bill image for scanning
func (h *BillScanHandler) UploadBillScan(c *fiber.Ctx) error {
	userID := c.Locals("userId").(primitive.ObjectID)

	// Get uploaded file
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No image file provided",
		})
	}

	// Validate file type
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
	}

	contentType := file.Header.Get("Content-Type")
	if !allowedTypes[contentType] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file type. Only JPEG and PNG images are allowed",
		})
	}

	// Validate file size (max 10MB)
	maxSize := int64(10 * 1024 * 1024)
	if file.Size > maxSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File too large. Maximum size is 10MB",
		})
	}

	// Create uploads directory if it doesn't exist
	uploadsDir := os.Getenv("UPLOADS_DIR")
	if uploadsDir == "" {
		uploadsDir = "./uploads/bill_scans"
	}

	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create uploads directory",
		})
	}

	// Generate unique filename
	timestamp := time.Now().Unix()
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s_%d%s", userID.Hex(), timestamp, ext)
	filePath := filepath.Join(uploadsDir, filename)

	// Save file
	if err := c.SaveFile(file, filePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save file",
		})
	}

	// Create scan record
	scan, err := h.scanService.CreateScan(c.Context(), userID, filePath)
	if err != nil {
		// Clean up uploaded file
		os.Remove(filePath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create scan record",
		})
	}

	// Process scan asynchronously
	go func() {
		ctx := c.Context()
		if err := h.scanService.ProcessScan(ctx, scan.ID); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to process scan %s: %v\n", scan.ID.Hex(), err)
		}
	}()

	return c.Status(fiber.StatusCreated).JSON(scan)
}

// GetBillScan retrieves a specific bill scan by ID
func (h *BillScanHandler) GetBillScan(c *fiber.Ctx) error {
	scanIDStr := c.Params("id")
	scanID, err := primitive.ObjectIDFromHex(scanIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid scan ID",
		})
	}

	scan, err := h.scanService.GetScan(c.Context(), scanID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Scan not found",
		})
	}

	return c.JSON(scan)
}

// GetUserScans retrieves all scans for the current user
func (h *BillScanHandler) GetUserScans(c *fiber.Ctx) error {
	userID := c.Locals("userId").(primitive.ObjectID)

	scans, err := h.scanService.GetUserScans(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve scans",
		})
	}

	return c.JSON(scans)
}
