package pdfparser

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/ledongthuc/pdf"

	"github.com/dwarvesf/fortress-api/pkg/logger"
)

var (
	// ErrEmptyPDF is returned when the PDF has no content
	ErrEmptyPDF = errors.New("PDF has no content")
	// ErrInvalidPDF is returned when the PDF is malformed or cannot be parsed
	ErrInvalidPDF = errors.New("invalid or corrupted PDF")
	// ErrEncryptedPDF is returned when the PDF is password-protected
	ErrEncryptedPDF = errors.New("PDF is password-protected")
)

type service struct {
	logger logger.Logger
}

// New creates a new PDF parser service
func New(l logger.Logger) IService {
	return &service{
		logger: l,
	}
}

// ExtractText extracts text content from a PDF byte slice
func (s *service) ExtractText(pdfBytes []byte) (string, error) {
	l := s.logger.Fields(logger.Fields{
		"service": "pdfparser",
		"method":  "ExtractText",
		"size":    len(pdfBytes),
	})
	l.Debug("extracting text from PDF")

	if len(pdfBytes) == 0 {
		l.Debug("empty PDF bytes provided")
		return "", ErrEmptyPDF
	}

	// Create a reader from bytes
	reader := bytes.NewReader(pdfBytes)

	// Open PDF from reader
	pdfReader, err := pdf.NewReader(reader, int64(len(pdfBytes)))
	if err != nil {
		l.Error(err, "failed to open PDF")

		// Check for common error types
		errStr := err.Error()
		if strings.Contains(errStr, "encrypted") || strings.Contains(errStr, "password") {
			return "", ErrEncryptedPDF
		}

		return "", fmt.Errorf("%w: %v", ErrInvalidPDF, err)
	}

	// Extract text from all pages
	var textBuilder strings.Builder
	numPages := pdfReader.NumPage()
	l.Debugf("PDF has %d pages", numPages)

	for i := 1; i <= numPages; i++ {
		page := pdfReader.Page(i)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			l.Debugf("failed to extract text from page %d: %v", i, err)
			continue
		}

		textBuilder.WriteString(text)
		textBuilder.WriteString("\n")
	}

	text := textBuilder.String()
	l.Debugf("extracted %d characters from PDF", len(text))

	if strings.TrimSpace(text) == "" {
		l.Debug("no text content found in PDF")
		return "", nil
	}

	return text, nil
}
