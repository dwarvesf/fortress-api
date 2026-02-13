package invoiceemail

import (
	"context"
	"fmt"
	"strings"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/googlemail"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
)

// Processor handles incoming invoice email processing
type Processor struct {
	cfg               *config.Config
	gmailService      googlemail.IService
	extractor         IExtractor
	payablesService   *notion.ContractorPayablesService
	logger            logger.Logger
	processedLabelID  string // Cached label ID
}

// NewProcessor creates a new invoice email processor
func NewProcessor(
	cfg *config.Config,
	gmailService googlemail.IService,
	extractor IExtractor,
	payablesService *notion.ContractorPayablesService,
	l logger.Logger,
) *Processor {
	return &Processor{
		cfg:             cfg,
		gmailService:    gmailService,
		extractor:       extractor,
		payablesService: payablesService,
		logger:          l,
	}
}

// ProcessIncomingInvoices processes unread invoice emails from the monitored inbox
func (p *Processor) ProcessIncomingInvoices(ctx context.Context) (*ProcessorStats, error) {
	l := p.logger.Fields(logger.Fields{
		"service": "invoiceemail",
		"method":  "ProcessIncomingInvoices",
	})

	if !p.cfg.InvoiceListener.Enabled {
		l.Debug("invoice listener is disabled")
		return &ProcessorStats{}, nil
	}

	l.Debug("starting invoice email processing")

	stats := &ProcessorStats{
		Results: make([]ProcessResult, 0),
	}

	// Build Gmail query: unread emails to the target address
	targetEmail := p.cfg.InvoiceListener.Email
	if targetEmail == "" {
		l.Debug("no target email configured")
		return stats, nil
	}

	// Query for emails to the target address that haven't been processed yet
	// We rely ONLY on the processed label to track what's been handled
	// NOT using is:unread - because emails could be read manually but not processed
	processedLabel := p.cfg.InvoiceListener.ProcessedLabel
	if processedLabel == "" {
		processedLabel = "Processed" // Default label
	}

	// Primary filter: emails to target WITHOUT processed label
	// Additional filter: subject contains "INVC-" (Invoice ID pattern) OR has PDF attachment
	// This distinguishes contractor invoice emails from other emails
	// Using in:anywhere to also search spam/junk, since group emails may be flagged as spam
	query := fmt.Sprintf("in:anywhere to:%s -label:%s (subject:INVC- OR has:attachment filename:pdf)",
		targetEmail, strings.ReplaceAll(processedLabel, "/", "-"))

	l.Debugf("filtering for contractor invoice emails (subject:INVC- OR PDF attachment)")

	l.Debugf("querying Gmail with: %s", query)

	maxMessages := p.cfg.InvoiceListener.MaxMessages
	if maxMessages == 0 {
		maxMessages = 50
	}

	messages, err := p.gmailService.ListInboxMessages(ctx, query, maxMessages)
	if err != nil {
		l.Error(err, "failed to list inbox messages")
		return nil, fmt.Errorf("failed to list inbox messages: %w", err)
	}

	stats.TotalEmails = len(messages)
	l.Debugf("found %d unread emails", len(messages))

	if len(messages) == 0 {
		return stats, nil
	}

	// Get or create the processed label
	if processedLabel != "" && p.processedLabelID == "" {
		labelID, err := p.gmailService.GetOrCreateLabel(ctx, processedLabel)
		if err != nil {
			l.Error(err, "failed to get/create processed label")
			// Continue without labeling
		} else {
			p.processedLabelID = labelID
			l.Debugf("using processed label ID: %s", labelID)
		}
	}

	// Process each email
	for _, msg := range messages {
		result := p.processEmail(ctx, msg)
		stats.Results = append(stats.Results, result)

		switch result.Status {
		case "success":
			stats.Processed++
		case "skipped":
			stats.Skipped++
		case "error":
			stats.Errors++
		}
	}

	l.Debugf("processing complete: total=%d processed=%d skipped=%d errors=%d",
		stats.TotalEmails, stats.Processed, stats.Skipped, stats.Errors)

	return stats, nil
}

// processEmail processes a single email message
func (p *Processor) processEmail(ctx context.Context, msg googlemail.InboxMessage) ProcessResult {
	l := p.logger.Fields(logger.Fields{
		"service":   "invoiceemail",
		"method":    "processEmail",
		"messageID": msg.ID,
	})

	result := ProcessResult{
		MessageID: msg.ID,
		Status:    "error",
	}

	// Get full message details
	fullMsg, err := p.gmailService.GetMessage(ctx, msg.ID)
	if err != nil {
		l.Error(err, "failed to get message details")
		result.Error = fmt.Sprintf("failed to get message: %v", err)
		return result
	}

	l.Debugf("processing email subject: %s", fullMsg.Subject)

	// Try to extract Invoice ID from subject first
	invoiceID, err := p.extractor.ExtractInvoiceIDFromSubject(fullMsg.Subject)
	if err != nil {
		l.Debug("no Invoice ID in subject, checking PDF attachment")

		// Check if there's a PDF attachment
		if !fullMsg.HasPDF {
			l.Debug("no PDF attachment found")
			result.Status = "skipped"
			result.Error = "no Invoice ID in subject and no PDF attachment"
			p.markAsProcessed(ctx, msg.ID, l)
			return result
		}

		// Get PDF attachment
		pdfBytes, err := p.gmailService.GetAttachment(ctx, msg.ID, fullMsg.PDFPartID)
		if err != nil {
			l.Error(err, "failed to get PDF attachment")
			result.Error = fmt.Sprintf("failed to get PDF: %v", err)
			return result
		}

		// Check PDF size limit
		maxSizeMB := p.cfg.InvoiceListener.PDFMaxSizeMB
		if maxSizeMB == 0 {
			maxSizeMB = 5
		}
		if len(pdfBytes) > maxSizeMB*1024*1024 {
			l.Debugf("PDF too large: %d bytes (max %d MB)", len(pdfBytes), maxSizeMB)
			result.Status = "skipped"
			result.Error = fmt.Sprintf("PDF too large: %d bytes", len(pdfBytes))
			p.markAsProcessed(ctx, msg.ID, l)
			return result
		}

		// Extract Invoice ID from PDF
		invoiceID, err = p.extractor.ExtractInvoiceIDFromPDF(pdfBytes)
		if err != nil {
			l.Debug("no Invoice ID found in PDF")
			result.Status = "skipped"
			result.Error = "no Invoice ID found in subject or PDF"
			p.markAsProcessed(ctx, msg.ID, l)
			return result
		}
	}

	result.InvoiceID = invoiceID
	l.Debugf("extracted Invoice ID: %s", invoiceID)

	// Find the payable in Notion
	payable, err := p.payablesService.FindPayableByInvoiceID(ctx, invoiceID)
	if err != nil {
		l.Error(err, "failed to find payable by Invoice ID")
		result.Error = fmt.Sprintf("failed to find payable: %v", err)
		return result
	}

	if payable == nil {
		l.Debugf("no payable found with Invoice ID: %s", invoiceID)
		result.Status = "skipped"
		result.Error = "no matching payable found with status 'New'"
		p.markAsProcessed(ctx, msg.ID, l)
		return result
	}

	l.Debugf("found payable: pageID=%s status=%s", payable.PageID, payable.Status)

	// Update payable status to "Pending"
	err = p.payablesService.UpdatePayableStatus(ctx, payable.PageID, "Pending", "")
	if err != nil {
		l.Error(err, "failed to update payable status")
		result.Error = fmt.Sprintf("failed to update payable: %v", err)
		return result
	}

	result.PageID = payable.PageID
	result.Status = "success"
	l.Debugf("updated payable %s to Pending", payable.PageID)

	// Mark email as processed
	p.markAsProcessed(ctx, msg.ID, l)

	return result
}

// markAsProcessed adds the processed label to the email
func (p *Processor) markAsProcessed(ctx context.Context, messageID string, l logger.Logger) {
	if p.processedLabelID == "" {
		return
	}

	err := p.gmailService.AddLabel(ctx, messageID, p.processedLabelID)
	if err != nil {
		l.Error(err, "failed to add processed label")
		// Non-fatal error, continue
	} else {
		l.Debug("marked email as processed")
	}
}
