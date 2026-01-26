package invoice

import (
	"errors"
	"fmt"
	"sync"
	"time"

	nt "github.com/dstotijn/go-notion"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	sInvoice "github.com/dwarvesf/fortress-api/pkg/store/invoice"
	"github.com/dwarvesf/fortress-api/pkg/worker"
)

// MarkPaidResult contains the result of marking an invoice as paid
type MarkPaidResult struct {
	InvoiceNumber   string    `json:"invoice_number"`
	Source          string    `json:"source"` // "postgres", "notion", or "both"
	PaidAt          time.Time `json:"paid_at"`
	PostgresUpdated bool      `json:"postgres_updated"`
	NotionUpdated   bool      `json:"notion_updated"`
}

// determineSource returns the source string based on where the invoice was found
func determineSource(pgFound, notionFound bool) string {
	if pgFound && notionFound {
		return "both"
	}
	if pgFound {
		return "postgres"
	}
	return "notion"
}

// MarkInvoiceAsPaidByNumber handles marking invoice as paid from Discord command.
// It searches both PostgreSQL and Notion, updates where found.
func (c *controller) MarkInvoiceAsPaidByNumber(invoiceNumber string) (*MarkPaidResult, error) {
	l := c.logger.Fields(logger.Fields{
		"controller":    "invoice",
		"method":        "MarkInvoiceAsPaidByNumber",
		"invoiceNumber": invoiceNumber,
	})

	l.Debug("starting mark invoice as paid by number")

	if invoiceNumber == "" {
		l.Debug("empty invoice number provided")
		return nil, errors.New("invoice number is required")
	}

	// 1. Search PostgreSQL
	l.Debug("searching invoice in PostgreSQL")
	pgInvoice, pgErr := c.store.Invoice.One(c.repo.DB(), &sInvoice.Query{Number: invoiceNumber})
	if pgErr != nil {
		if errors.Is(pgErr, gorm.ErrRecordNotFound) {
			// Record not found - set to nil explicitly
			pgInvoice = nil
			l.Debug("invoice not found in PostgreSQL")
		} else {
			l.Errorf(pgErr, "failed to query PostgreSQL for invoice")
			return nil, fmt.Errorf("failed to query PostgreSQL: %w", pgErr)
		}
	}
	if pgInvoice != nil {
		l.Debugf("found invoice in PostgreSQL: id=%s status=%s", pgInvoice.ID, pgInvoice.Status)
	}

	// 2. Search Notion
	l.Debug("searching invoice in Notion Client Invoices")
	notionPage, notionErr := c.service.Notion.QueryClientInvoiceByNumber(invoiceNumber)
	if notionErr != nil {
		l.Debugf("error searching invoice in Notion: %v", notionErr)
	}
	if notionPage != nil {
		l.Debugf("found invoice in Notion: pageID=%s", notionPage.ID)
	} else {
		l.Debug("invoice not found in Notion")
	}

	// 3. Check if found anywhere
	if pgInvoice == nil && notionPage == nil {
		l.Debug("invoice not found in either PostgreSQL or Notion")
		return nil, fmt.Errorf("invoice %s not found", invoiceNumber)
	}

	result := &MarkPaidResult{
		InvoiceNumber: invoiceNumber,
		PaidAt:        time.Now(),
	}

	// 4. Validate statuses before proceeding
	if pgInvoice != nil {
		if pgInvoice.Status != model.InvoiceStatusSent && pgInvoice.Status != model.InvoiceStatusOverdue {
			l.Debugf("PostgreSQL invoice status validation failed: currentStatus=%v", pgInvoice.Status)
			return nil, fmt.Errorf("cannot mark as paid: PostgreSQL invoice status is %s (must be sent or overdue)", pgInvoice.Status)
		}
	}

	if notionPage != nil {
		notionStatus, err := c.service.Notion.GetNotionInvoiceStatus(notionPage)
		if err != nil {
			l.Debugf("failed to get Notion invoice status: %v", err)
		} else {
			l.Debugf("Notion invoice status: %s", notionStatus)
			if notionStatus != "Sent" && notionStatus != "Overdue" {
				return nil, fmt.Errorf("cannot mark as paid: Notion invoice status is %s (must be Sent or Overdue)", notionStatus)
			}
		}
	}

	// 5. Update PostgreSQL if exists (TEMPORARILY DISABLED)
	// TODO: Re-enable PostgreSQL update when ready
	// if pgInvoice != nil {
	// 	l.Debug("processing PostgreSQL invoice")
	// 	// Use existing logic (includes commission, accounting, email, GDrive)
	// 	_, err := c.MarkInvoiceAsPaidWithTaskRef(pgInvoice, nil, true)
	// 	if err != nil {
	// 		l.Errorf(err, "failed to mark PostgreSQL invoice as paid")
	// 		return nil, fmt.Errorf("failed to mark PostgreSQL invoice as paid: %w", err)
	// 	}
	// 	result.PostgresUpdated = true
	// 	l.Debug("PostgreSQL invoice marked as paid")
	// }
	l.Debug("PostgreSQL update temporarily disabled")

	// 6. Update Notion if exists
	// Note: skipEmailAndGDrive=false since PostgreSQL update is disabled
	if notionPage != nil {
		l.Debug("processing Notion invoice")
		err := c.processNotionInvoicePaid(l, notionPage, false)
		if err != nil {
			l.Errorf(err, "failed to process Notion invoice")
			// Continue - don't fail the whole operation
		} else {
			result.NotionUpdated = true
			l.Debug("Notion invoice marked as paid")
		}
	}

	// 7. Determine source
	result.Source = determineSource(pgInvoice != nil, notionPage != nil)
	l.Debugf("mark invoice as paid completed: source=%s", result.Source)

	return result, nil
}

// processNotionInvoicePaid handles updating a Notion invoice as paid and performing post-processing.
// If skipEmailAndGDrive is true, it skips email and GDrive operations (used when PostgreSQL was also updated).
func (c *controller) processNotionInvoicePaid(l logger.Logger, page *nt.Page, skipEmailAndGDrive bool) error {
	l.Debugf("processNotionInvoicePaid: pageID=%s skipEmailAndGDrive=%v", page.ID, skipEmailAndGDrive)

	// 1. Update Notion status to "Paid" and set Paid Date
	paidDate := time.Now()
	l.Debug("updating Notion invoice status to Paid")
	if err := c.service.Notion.UpdateClientInvoiceStatus(page.ID, "Paid", &paidDate); err != nil {
		l.Errorf(err, "failed to update Notion invoice status")
		return fmt.Errorf("failed to update Notion status: %w", err)
	}
	l.Debug("Notion invoice status updated to Paid")

	// 1a. Update Line Items status to "Paid"
	l.Debug("updating Line Items status to Paid")
	if err := c.service.Notion.UpdateLineItemsStatus(page.ID, "Paid"); err != nil {
		l.Errorf(err, "failed to update Line Items status")
		// Log error but don't fail the whole operation - invoice itself was already updated
	} else {
		l.Debug("Line Items status updated to Paid")
	}

	// 1b. Enqueue invoice splits generation job
	l.Debug("enqueuing invoice splits generation job")
	c.worker.Enqueue(worker.GenerateInvoiceSplitsMsg, worker.GenerateInvoiceSplitsPayload{
		InvoicePageID: page.ID,
	})
	l.Debug("invoice splits generation job enqueued")

	// If PostgreSQL was also updated, skip email and GDrive to avoid duplicates
	if skipEmailAndGDrive {
		l.Debug("skipping email and GDrive operations (PostgreSQL was updated)")
		return nil
	}

	// 2. Extract invoice data for email and GDrive
	l.Debug("extracting Notion invoice data for post-processing")
	notionInvoice, err := c.service.Notion.ExtractClientInvoiceData(page)
	if err != nil {
		l.Errorf(err, "failed to extract Notion invoice data")
		return fmt.Errorf("failed to extract invoice data: %w", err)
	}

	// 3. Run email and GDrive operations in parallel
	wg := &sync.WaitGroup{}
	wg.Add(2)

	// 3a. Send thank you email
	go func() {
		defer wg.Done()
		l.Debug("sending thank you email for Notion invoice")
		if err := c.service.GoogleMail.SendInvoiceThankYouMail(notionInvoice); err != nil {
			l.Errorf(err, "failed to send thank you email for Notion invoice")
		} else {
			l.Debug("thank you email sent for Notion invoice")
		}
	}()

	// 3b. Move PDF in GDrive (Sent → Paid)
	go func() {
		defer wg.Done()
		l.Debug("moving Notion invoice PDF to Paid folder in GDrive")
		if err := c.service.GoogleDrive.MoveInvoicePDF(notionInvoice, "Sent", "Paid"); err != nil {
			l.Errorf(err, "failed to move Notion invoice PDF in GDrive")
		} else {
			l.Debug("Notion invoice PDF moved to Paid folder")
		}
	}()

	wg.Wait()

	return nil
}
