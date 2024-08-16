package invoiceemail

import (
	"log"
	"regexp"
	"strings"
)

type InvoiceDetector struct {
	// Add fields for invoice detection configuration
}

func NewInvoiceDetector() *InvoiceDetector {
	return &InvoiceDetector{}
}

func (d *InvoiceDetector) DetectInvoice(email interface{}) bool {
	// Basic invoice detection logic
	e, ok := email.(InvoiceEmail)
	if !ok {
		return false
	}

	subject := strings.ToLower(e.Subject)
	content := strings.ToLower(e.Content)

	// Check if subject or content contains invoice-related keywords
	keywords := []string{"invoice", "bill", "payment", "due"}
	for _, keyword := range keywords {
		if strings.Contains(subject, keyword) || strings.Contains(content, keyword) {
			return true
		}
	}

	return false
}

func (d *InvoiceDetector) ExtractInvoiceData(email interface{}) map[string]interface{} {
	e, ok := email.(InvoiceEmail)
	if !ok {
		return nil
	}

	// Basic invoice data extraction logic
	invoiceData := make(map[string]interface{})
	invoiceData["sender"] = e.Sender
	invoiceData["subject"] = e.Subject
	invoiceData["received_at"] = e.ReceivedAt

	// Extract invoice number
	invoiceNumberRegex := regexp.MustCompile(`(?i)invoice\s*#?\s*(\w+)`)
	if match := invoiceNumberRegex.FindStringSubmatch(e.Content); len(match) > 1 {
		invoiceData["invoice_number"] = match[1]
	}

	// Extract invoice amount
	amountRegex := regexp.MustCompile(`(?i)amount.*?\$?\s*([\d,]+\.?\d*)`)
	if match := amountRegex.FindStringSubmatch(e.Content); len(match) > 1 {
		invoiceData["invoice_amount"] = match[1]
	}

	// Extract invoice date
	dateRegex := regexp.MustCompile(`(?i)date.*?(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})`)
	if match := dateRegex.FindStringSubmatch(e.Content); len(match) > 1 {
		invoiceData["invoice_date"] = match[1]
	}

	invoiceData["content"] = e.Content

	return invoiceData
}