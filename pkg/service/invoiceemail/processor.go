package invoiceemail

import (
	"context"
	"fmt"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/discord"
	"github.com/dwarvesf/fortress-api/pkg/service/googlemail"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
)

// Processor handles incoming invoice email processing
type Processor struct {
	cfg             *config.Config
	gmailService    googlemail.IService
	extractor       IExtractor
	payablesService *notion.ContractorPayablesService
	discordService  discord.IService
	logger          logger.Logger
}

// NewProcessor creates a new invoice email processor
func NewProcessor(
	cfg *config.Config,
	gmailService googlemail.IService,
	extractor IExtractor,
	payablesService *notion.ContractorPayablesService,
	discordService discord.IService,
	l logger.Logger,
) *Processor {
	return &Processor{
		cfg:             cfg,
		gmailService:    gmailService,
		extractor:       extractor,
		payablesService: payablesService,
		discordService:  discordService,
		logger:          l,
	}
}

// ProcessIncomingInvoices fetches "New" payables from Notion and searches Gmail for matching emails.
// Payable-first flow: Notion payables drive the search, not inbox scanning.
func (p *Processor) ProcessIncomingInvoices(ctx context.Context) (*ProcessorStats, error) {
	l := p.logger.Fields(logger.Fields{
		"service": "invoiceemail",
		"method":  "ProcessIncomingInvoices",
	})

	if !p.cfg.InvoiceListener.Enabled {
		l.Debug("invoice listener is disabled")
		return &ProcessorStats{}, nil
	}

	l.Debug("starting payable-first invoice email processing")

	stats := &ProcessorStats{
		Results: make([]ProcessResult, 0),
	}

	targetEmail := p.cfg.InvoiceListener.Email
	if targetEmail == "" {
		l.Debug("no target email configured")
		return stats, nil
	}

	// Step 1: Query all "New" payables from Notion
	newPayables, err := p.payablesService.QueryNewPayables(ctx)
	if err != nil {
		l.Error(err, "failed to query new payables from Notion")
		return nil, fmt.Errorf("failed to query new payables: %w", err)
	}

	stats.TotalEmails = len(newPayables)
	l.Debugf("found %d new payables to match", len(newPayables))

	if len(newPayables) == 0 {
		return stats, nil
	}

	// Step 2: For each payable, search Gmail for matching email
	for _, payable := range newPayables {
		result := p.matchPayableToEmail(ctx, payable, targetEmail)
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

// matchPayableToEmail searches Gmail for an email matching the payable's Invoice ID.
// First checks subject lines, then falls back to PDF attachment content.
func (p *Processor) matchPayableToEmail(ctx context.Context, payable notion.NewPayable, targetEmail string) ProcessResult {
	l := p.logger.Fields(logger.Fields{
		"service":   "invoiceemail",
		"method":    "matchPayableToEmail",
		"invoiceID": payable.InvoiceID,
		"pageID":    payable.PageID,
	})

	result := ProcessResult{
		InvoiceID: payable.InvoiceID,
		Status:    "skipped",
	}

	// Search Gmail: subject contains the Invoice ID
	subjectQuery := fmt.Sprintf("in:anywhere to:%s subject:%s",
		targetEmail, payable.InvoiceID)

	l.Debugf("searching Gmail with subject query: %s", subjectQuery)

	messages, err := p.gmailService.ListInboxMessages(ctx, subjectQuery, 1)
	if err != nil {
		l.Error(err, "failed to search Gmail by subject")
		result.Status = "error"
		result.Error = fmt.Sprintf("failed to search Gmail: %v", err)
		return result
	}

	// If no subject match, optionally fall back to PDF attachment scanning
	if len(messages) == 0 {
		if !p.cfg.InvoiceListener.PDFFallback {
			l.Debug("no subject match and PDF fallback disabled, will retry next poll")
			result.Error = "no matching email found (PDF fallback disabled)"
			return result
		}

		l.Debug("no subject match, searching for emails with PDF attachments")

		pdfQuery := fmt.Sprintf("in:anywhere to:%s has:attachment filename:pdf",
			targetEmail)

		messages, err = p.gmailService.ListInboxMessages(ctx, pdfQuery, 20)
		if err != nil {
			l.Error(err, "failed to search Gmail for PDF emails")
			result.Status = "error"
			result.Error = fmt.Sprintf("failed to search Gmail for PDFs: %v", err)
			return result
		}

		// Check each PDF email for the Invoice ID
		var matchedMsg *googlemail.InboxMessage
		for _, msg := range messages {
			fullMsg, err := p.gmailService.GetMessage(ctx, msg.ID)
			if err != nil {
				l.Debugf("failed to get message %s: %v", msg.ID, err)
				continue
			}

			if !fullMsg.HasPDF {
				continue
			}

			pdfBytes, err := p.gmailService.GetAttachment(ctx, msg.ID, fullMsg.PDFPartID)
			if err != nil {
				l.Debugf("failed to get PDF from message %s: %v", msg.ID, err)
				continue
			}

			// Check PDF size limit
			maxSizeMB := p.cfg.InvoiceListener.PDFMaxSizeMB
			if maxSizeMB == 0 {
				maxSizeMB = 5
			}
			if len(pdfBytes) > maxSizeMB*1024*1024 {
				l.Debugf("PDF too large in message %s: %d bytes", msg.ID, len(pdfBytes))
				continue
			}

			extractedID, err := p.extractor.ExtractInvoiceIDFromPDF(pdfBytes)
			if err != nil {
				continue
			}

			if extractedID == payable.InvoiceID {
				l.Debugf("found matching Invoice ID in PDF of message %s", msg.ID)
				matchedMsg = fullMsg
				break
			}
		}

		if matchedMsg == nil {
			l.Debug("no matching email found, will retry next poll")
			result.Error = "no matching email found"
			return result
		}

		// Found via PDF — proceed with this message
		return p.handleMatchedEmail(ctx, matchedMsg, payable, l)
	}

	// Found via subject — get full message
	fullMsg, err := p.gmailService.GetMessage(ctx, messages[0].ID)
	if err != nil {
		l.Error(err, "failed to get matched message details")
		result.Status = "error"
		result.Error = fmt.Sprintf("failed to get message: %v", err)
		return result
	}

	return p.handleMatchedEmail(ctx, fullMsg, payable, l)
}

// handleMatchedEmail processes a matched email: updates payable to Pending, labels email, sends Discord notification.
func (p *Processor) handleMatchedEmail(ctx context.Context, msg *googlemail.InboxMessage, payable notion.NewPayable, l logger.Logger) ProcessResult {
	result := ProcessResult{
		InvoiceID: payable.InvoiceID,
		MessageID: msg.ID,
		Status:    "error",
	}

	l.Debugf("matched email: messageID=%s subject=%s from=%s", msg.ID, msg.Subject, msg.From)

	// Update payable status to "Pending"
	err := p.payablesService.UpdatePayableStatus(ctx, payable.PageID, "Pending", "")
	if err != nil {
		l.Error(err, "failed to update payable status")
		result.Error = fmt.Sprintf("failed to update payable: %v", err)
		return result
	}

	result.PageID = payable.PageID
	result.Status = "success"
	l.Debugf("updated payable %s to Pending", payable.PageID)

	// Send Discord notification
	p.sendDiscordNotification(msg, payable.InvoiceID, l)

	return result
}

// sendDiscordNotification sends a notification to the Discord audit log channel
func (p *Processor) sendDiscordNotification(msg *googlemail.InboxMessage, invoiceID string, l logger.Logger) {
	if p.discordService == nil {
		l.Debug("discord service not available, skipping notification")
		return
	}

	webhookURL := p.cfg.Discord.Webhooks.AuditLog
	if webhookURL == "" {
		l.Debug("no audit log webhook URL configured, skipping notification")
		return
	}

	content := fmt.Sprintf("**Contractor Invoice Received**\nInvoice ID: `%s`\nFrom: %s\nSubject: %s\nPayable status updated to **Pending**",
		invoiceID, msg.From, msg.Subject)

	_, err := p.discordService.SendMessage(model.DiscordMessage{
		Content: content,
	}, webhookURL)
	if err != nil {
		l.Error(err, "failed to send Discord notification")
	}
}

// StartPolling starts a background loop that polls for invoice emails at the configured interval.
// It blocks until the context is cancelled.
func (p *Processor) StartPolling(ctx context.Context) {
	l := p.logger.Fields(logger.Fields{
		"service": "invoiceemail",
		"method":  "StartPolling",
	})

	interval := p.cfg.InvoiceListener.PollInterval
	if interval == 0 {
		interval = 5 * time.Minute
	}

	l.Infof("invoice email polling started, interval=%s", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run immediately on start
	p.poll(ctx, l)

	for {
		select {
		case <-ctx.Done():
			l.Info("invoice email polling stopped")
			return
		case <-ticker.C:
			p.poll(ctx, l)
		}
	}
}

func (p *Processor) poll(ctx context.Context, l logger.Logger) {
	stats, err := p.ProcessIncomingInvoices(ctx)
	if err != nil {
		l.Error(err, "invoice email poll failed")
		return
	}

	if stats.TotalEmails > 0 {
		l.Infof("invoice email poll complete: total=%d processed=%d skipped=%d errors=%d",
			stats.TotalEmails, stats.Processed, stats.Skipped, stats.Errors)
	}
}

