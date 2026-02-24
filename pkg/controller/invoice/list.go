package invoice

import (
	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	// "github.com/dwarvesf/fortress-api/pkg/store/invoice" // TEMPORARILY DISABLED
)

type GetListInvoiceInput struct {
	model.Pagination
	ProjectIDs    []string
	Statuses      []string
	InvoiceNumber string
	OnProgress    func(completed, total int) // optional progress callback, nil-safe
}

func (c *controller) List(in GetListInvoiceInput) ([]*model.Invoice, int64, error) {
	l := c.logger.Fields(logger.Fields{
		"controller": "invoice",
		"method":     "List",
		"input":      in,
	})

	l.Debug("fetching invoices from Notion only (PostgreSQL temporarily disabled)")

	// Step 1: Fetch invoices from PostgreSQL (TEMPORARILY DISABLED)
	// TODO: Re-enable PostgreSQL fetch when ready
	// l.Debug("fetching invoices from PostgreSQL")
	// pgInvoices, _, err := c.store.Invoice.All(c.repo.DB(), invoice.GetInvoicesFilter{
	// 	Preload:       true,
	// 	ProjectIDs:    in.ProjectIDs,
	// 	Statuses:      in.Statuses,
	// 	InvoiceNumber: in.InvoiceNumber,
	// }, in.Pagination)
	// if err != nil {
	// 	l.Error(err, "failed to get invoice list from PostgreSQL")
	// 	return nil, 0, err
	// }
	// l.Debugf("fetched %d invoices from PostgreSQL", len(pgInvoices))
	pgInvoices := []*model.Invoice{}

	// Step 2: Fetch invoices from Notion
	l.Debug("fetching invoices from Notion")
	notionFilter := &notion.InvoiceFilter{
		ProjectIDs:    in.ProjectIDs,
		Statuses:      in.Statuses,
		InvoiceNumber: in.InvoiceNumber,
	}

	notionPages, _, err := c.service.Notion.QueryInvoices(notionFilter, in.Pagination)
	if err != nil {
		l.Error(err, "failed to query invoices from Notion")
		// Don't fail the entire request if Notion fails - continue with PG data only
		l.Debug("continuing with PostgreSQL data only due to Notion error")
		return pgInvoices, int64(len(pgInvoices)), nil
	}
	l.Debugf("fetched %d invoice pages from Notion", len(notionPages))

	// Step 3: Transform Notion pages to API Invoice models (concurrent processing)
	l.Debug("transforming Notion pages to Invoice models using concurrent workers")

	// Use a buffered channel to collect results
	type invoiceResult struct {
		invoice *model.Invoice
		pageID  string
		err     error
	}

	results := make(chan invoiceResult, len(notionPages))

	// Process each invoice concurrently
	for _, page := range notionPages {
		go func(p nt.Page) {
			// Fetch line items for this invoice
			lineItems, err := c.service.Notion.GetInvoiceLineItems(p.ID)
			if err != nil {
				results <- invoiceResult{pageID: p.ID, err: err}
				return
			}

			// Transform to API model
			apiInvoice, err := NotionPageToInvoice(p, lineItems, c.service.Notion, l)
			if err != nil {
				results <- invoiceResult{pageID: p.ID, err: err}
				return
			}

			results <- invoiceResult{invoice: apiInvoice, pageID: p.ID}
		}(page)
	}

	// Collect results
	notionInvoices := make([]*model.Invoice, 0, len(notionPages))
	for i := 0; i < len(notionPages); i++ {
		result := <-results
		if result.err != nil {
			l.AddField("pageID", result.pageID).Error(result.err, "failed to process Notion invoice")
		} else {
			notionInvoices = append(notionInvoices, result.invoice)
		}
		if in.OnProgress != nil {
			in.OnProgress(i+1, len(notionPages))
		}
	}
	close(results)

	l.Debugf("successfully transformed %d Notion invoices (concurrent processing)", len(notionInvoices))

	// Step 4: Merge invoices from both sources
	l.Debug("merging invoices from PostgreSQL and Notion")
	mergedInvoices := mergeInvoices(pgInvoices, notionInvoices, l)
	l.Debugf("merge complete: final count=%d", len(mergedInvoices))

	// Step 5: Apply pagination to merged results
	// Note: Since we merged data, we need to re-apply pagination
	total := int64(len(mergedInvoices))

	// Calculate pagination bounds
	offset := int(in.Pagination.Page-1) * int(in.Pagination.Size)
	if offset < 0 {
		offset = 0
	}

	end := offset + int(in.Pagination.Size)
	if end > len(mergedInvoices) {
		end = len(mergedInvoices)
	}

	// Apply pagination slice
	var paginatedInvoices []*model.Invoice
	if offset < len(mergedInvoices) {
		paginatedInvoices = mergedInvoices[offset:end]
	} else {
		paginatedInvoices = []*model.Invoice{}
	}

	l.Debugf("pagination applied: offset=%d, end=%d, returned=%d, total=%d",
		offset, end, len(paginatedInvoices), total)

	return paginatedInvoices, total, nil
}
