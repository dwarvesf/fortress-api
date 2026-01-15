package invoice

import (
	"fmt"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/dwarvesf/fortress-api/pkg/worker"
)

// GenerateInvoiceSplitsByLegacyNumber generates invoice splits by querying Notion Client Invoices
// and enqueuing a worker job for async processing
func (c *controller) GenerateInvoiceSplitsByLegacyNumber(legacyNumber string) (*view.GenerateSplitsResponse, error) {
	l := c.logger.Fields(logger.Fields{
		"controller":   "invoice",
		"method":       "GenerateInvoiceSplitsByLegacyNumber",
		"legacyNumber": legacyNumber,
	})

	l.Debug("starting generate invoice splits by legacy number")

	if legacyNumber == "" {
		l.Debug("empty legacy number provided")
		return nil, fmt.Errorf("legacy number is required")
	}

	// 1. Query Notion Client Invoices database
	l.Debug("querying Notion Client Invoices database")
	notionPage, err := c.service.Notion.QueryClientInvoiceByNumber(legacyNumber)
	if err != nil {
		l.Errorf(err, "failed to query Notion for invoice with legacy number: %s", legacyNumber)
		return nil, fmt.Errorf("invoice %s not found: %w", legacyNumber, err)
	}

	if notionPage == nil {
		l.Debugf("invoice with legacy number %s not found in Notion", legacyNumber)
		return nil, fmt.Errorf("invoice %s not found", legacyNumber)
	}

	l.Debugf("found invoice in Notion: pageID=%s", notionPage.ID)

	// 2. Enqueue worker job for invoice splits generation
	l.Debug("enqueuing invoice splits generation job")
	c.worker.Enqueue(worker.GenerateInvoiceSplitsMsg, worker.GenerateInvoiceSplitsPayload{
		InvoicePageID: notionPage.ID,
	})
	l.Infof("invoice splits generation job enqueued for invoice: %s (pageID: %s)", legacyNumber, notionPage.ID)

	// 3. Return response
	response := &view.GenerateSplitsResponse{
		LegacyNumber:  legacyNumber,
		InvoicePageID: notionPage.ID,
		JobEnqueued:   true,
		Message:       "Invoice splits generation job enqueued successfully",
	}

	l.Debug("generate invoice splits completed successfully")
	return response, nil
}
