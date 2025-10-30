package services

import (
	"fmt"
	"os"

	"github.com/otiai10/gosseract/v2"
)

type OCRService struct {
	tesseractPath string
}

func NewOCRService() *OCRService {
	tesseractPath := os.Getenv("TESSERACT_PATH")
	if tesseractPath == "" {
		tesseractPath = "/usr/bin/tesseract"
	}
	return &OCRService{
		tesseractPath: tesseractPath,
	}
}

// ExtractTextFromImage performs OCR on an image file and returns the extracted text
func (s *OCRService) ExtractTextFromImage(imagePath string) (string, error) {
	client := gosseract.NewClient()
	defer client.Close()

	// Set language to Polish and English for better results with Polish bills
	client.SetLanguage("pol", "eng")

	// Set image path
	err := client.SetImage(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to set image: %w", err)
	}

	// Extract text
	text, err := client.Text()
	if err != nil {
		return "", fmt.Errorf("failed to extract text: %w", err)
	}

	return text, nil
}
