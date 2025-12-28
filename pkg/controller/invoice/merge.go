package invoice

import (
	"sort"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

// mergeInvoices merges invoices from PostgreSQL and Notion databases
// Deduplication strategy: use invoice number as key, Notion version takes precedence over PG
// Returns merged list sorted by InvoicedAt descending
func mergeInvoices(pgInvoices []*model.Invoice, notionInvoices []*model.Invoice, l logger.Logger) []*model.Invoice {
	l = l.Fields(logger.Fields{
		"function":      "mergeInvoices",
		"pgCount":       len(pgInvoices),
		"notionCount":   len(notionInvoices),
	})

	l.Debug("starting invoice merge process")

	// Use invoice number as deduplication key
	invoiceMap := make(map[string]*model.Invoice)

	// First, add all PG invoices to map
	for _, inv := range pgInvoices {
		if inv == nil {
			l.Debug("skipping nil PG invoice")
			continue
		}
		if inv.Number == "" {
			l.AddField("invoiceID", inv.ID).Debug("skipping PG invoice with empty number")
			continue
		}
		invoiceMap[inv.Number] = inv
		l.Debugf("added PG invoice to map: number=%s", inv.Number)
	}

	duplicateCount := 0

	// Then, iterate through Notion invoices
	// If duplicate found (by invoice number), replace with Notion version (Notion takes precedence)
	for _, inv := range notionInvoices {
		if inv == nil {
			l.Debug("skipping nil Notion invoice")
			continue
		}
		if inv.Number == "" {
			l.Debug("skipping Notion invoice with empty number")
			continue
		}

		if _, exists := invoiceMap[inv.Number]; exists {
			// Duplicate found - Notion takes precedence
			duplicateCount++
			l.Debugf("duplicate invoice found, preferring Notion: number=%s", inv.Number)
		} else {
			l.Debugf("added new Notion invoice to map: number=%s", inv.Number)
		}

		// Add or replace with Notion version
		invoiceMap[inv.Number] = inv
	}

	// Convert map values to slice
	merged := make([]*model.Invoice, 0, len(invoiceMap))
	for _, inv := range invoiceMap {
		merged = append(merged, inv)
	}

	// Sort by InvoicedAt descending (most recent first)
	sort.Slice(merged, func(i, j int) bool {
		// Handle nil InvoicedAt fields - put them at the end
		if merged[i].InvoicedAt == nil {
			return false
		}
		if merged[j].InvoicedAt == nil {
			return true
		}
		return merged[i].InvoicedAt.After(*merged[j].InvoicedAt)
	})

	l.Debugf("merged results: PG=%d, Notion=%d, deduplicated=%d, final=%d",
		len(pgInvoices), len(notionInvoices), duplicateCount, len(merged))

	l.Debug("invoice merge process completed")

	return merged
}
