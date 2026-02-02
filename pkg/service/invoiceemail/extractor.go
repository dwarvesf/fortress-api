package invoiceemail

import (
	"errors"
	"regexp"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/pdfparser"
)

var (
	// ErrInvoiceIDNotFound is returned when no Invoice ID is found
	ErrInvoiceIDNotFound = errors.New("invoice ID not found")
)

// Invoice ID pattern: INVC-YYYYMM-NAME-XXXX (e.g., INVC-202601-QUANG-4DRE)
// YYYYMM is a 6-digit date, NAME is identifier, XXXX is random suffix
var invoiceIDPattern = regexp.MustCompile(`INVC-\d{6}-[A-Z0-9]+-[A-Z0-9]+`)

// Extractor handles Invoice ID extraction from various sources
type Extractor struct {
	pdfParser pdfparser.IService
	logger    logger.Logger
}

// NewExtractor creates a new Invoice ID extractor
func NewExtractor(pdfParser pdfparser.IService, l logger.Logger) *Extractor {
	return &Extractor{
		pdfParser: pdfParser,
		logger:    l,
	}
}

// ExtractInvoiceIDFromSubject extracts Invoice ID from email subject line
func (e *Extractor) ExtractInvoiceIDFromSubject(subject string) (string, error) {
	l := e.logger.Fields(logger.Fields{
		"service": "invoiceemail",
		"method":  "ExtractInvoiceIDFromSubject",
		"subject": subject,
	})
	l.Debug("extracting Invoice ID from subject")

	match := invoiceIDPattern.FindString(subject)
	if match == "" {
		l.Debug("no Invoice ID found in subject")
		return "", ErrInvoiceIDNotFound
	}

	l.Debugf("found Invoice ID in subject: %s", match)
	return match, nil
}

// ExtractInvoiceIDFromPDF extracts Invoice ID from PDF content
func (e *Extractor) ExtractInvoiceIDFromPDF(pdfBytes []byte) (string, error) {
	l := e.logger.Fields(logger.Fields{
		"service": "invoiceemail",
		"method":  "ExtractInvoiceIDFromPDF",
		"size":    len(pdfBytes),
	})
	l.Debug("extracting Invoice ID from PDF")

	// Extract text from PDF
	text, err := e.pdfParser.ExtractText(pdfBytes)
	if err != nil {
		l.Error(err, "failed to extract text from PDF")
		return "", err
	}

	if text == "" {
		l.Debug("no text content in PDF")
		return "", ErrInvoiceIDNotFound
	}

	// Search for Invoice ID pattern in extracted text
	match := invoiceIDPattern.FindString(text)
	if match == "" {
		l.Debug("no Invoice ID found in PDF content")
		return "", ErrInvoiceIDNotFound
	}

	l.Debugf("found Invoice ID in PDF: %s", match)
	return match, nil
}

// ExtractInvoiceID tries to extract Invoice ID from subject first, then falls back to PDF
func (e *Extractor) ExtractInvoiceID(subject string, pdfBytes []byte) (string, error) {
	l := e.logger.Fields(logger.Fields{
		"service": "invoiceemail",
		"method":  "ExtractInvoiceID",
		"subject": subject,
		"hasPDF":  len(pdfBytes) > 0,
	})
	l.Debug("extracting Invoice ID with fallback")

	// Try subject first (faster)
	invoiceID, err := e.ExtractInvoiceIDFromSubject(subject)
	if err == nil {
		l.Debugf("extracted Invoice ID from subject: %s", invoiceID)
		return invoiceID, nil
	}

	// Fallback to PDF if available
	if len(pdfBytes) > 0 {
		l.Debug("subject extraction failed, trying PDF fallback")
		invoiceID, err = e.ExtractInvoiceIDFromPDF(pdfBytes)
		if err == nil {
			l.Debugf("extracted Invoice ID from PDF: %s", invoiceID)
			return invoiceID, nil
		}
	}

	l.Debug("failed to extract Invoice ID from both subject and PDF")
	return "", ErrInvoiceIDNotFound
}
