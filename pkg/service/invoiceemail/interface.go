package invoiceemail

import "context"

// ProcessResult contains the result of processing an email
type ProcessResult struct {
	MessageID  string
	InvoiceID  string
	Status     string // "success", "skipped", "error"
	Error      string
	PageID     string // Notion page ID if updated
}

// ProcessorStats contains statistics from a processing run
type ProcessorStats struct {
	TotalEmails    int
	Processed      int
	Skipped        int
	Errors         int
	Results        []ProcessResult
}

// IProcessor defines the interface for invoice email processing
type IProcessor interface {
	// ProcessIncomingInvoices processes unread invoice emails from the monitored inbox
	// Returns processing statistics
	ProcessIncomingInvoices(ctx context.Context) (*ProcessorStats, error)

	// StartPolling starts a background loop that polls for invoice emails at the configured interval.
	// It blocks until the context is cancelled.
	StartPolling(ctx context.Context)
}

// IExtractor defines the interface for Invoice ID extraction
type IExtractor interface {
	// ExtractInvoiceIDFromSubject extracts Invoice ID from email subject line
	ExtractInvoiceIDFromSubject(subject string) (string, error)

	// ExtractInvoiceIDFromPDF extracts Invoice ID from PDF content
	ExtractInvoiceIDFromPDF(pdfBytes []byte) (string, error)

	// ExtractInvoiceID tries to extract Invoice ID from subject first, then falls back to PDF
	ExtractInvoiceID(subject string, pdfBytes []byte) (string, error)
}
