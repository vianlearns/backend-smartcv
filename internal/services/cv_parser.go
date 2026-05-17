package services

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ledongthuc/pdf"
)

// ParsePDF extracts text content from a PDF file
func (s *AIService) ParsePDF(data []byte) (string, error) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("failed to create PDF reader: %v", err)
	}

	var text strings.Builder
	numPages := reader.NumPage()

	for i := 1; i <= numPages; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}

		// Extract plain text from page
		plainText, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		text.WriteString(plainText)
		text.WriteString("\n")
	}

	return text.String(), nil
}

// ParseDOCX extracts text content from a DOCX file
// Note: For DOCX, we'll use a simpler approach - extract from XML
func (s *AIService) ParseDOCX(data []byte) (string, error) {
	// DOCX files are ZIP archives containing XML
	// For simplicity, we'll use archive/zip to extract document.xml
	// This is a basic implementation - for production use a proper DOCX library
	
	return "", fmt.Errorf("DOCX parsing not yet implemented - please use PDF or plain text")
}

// ParseAndExtract parses a file (PDF or text) and extracts CV data using AI
func (s *AIService) ParseAndExtract(data []byte, filename string) (string, error) {
	var content string
	var err error

	lowerFilename := strings.ToLower(filename)

	if strings.HasSuffix(lowerFilename, ".pdf") {
		content, err = s.ParsePDF(data)
		if err != nil {
			return "", fmt.Errorf("failed to parse PDF: %v", err)
		}
	} else if strings.HasSuffix(lowerFilename, ".docx") {
		return "", fmt.Errorf("DOCX files are not supported yet. Please convert to PDF or paste text directly.")
	} else {
		// Try to parse as plain text
		content = string(data)
	}

	// Use AI to extract structured data
	return s.ExtractCVData(content)
}
