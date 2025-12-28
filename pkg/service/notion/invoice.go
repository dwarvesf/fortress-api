package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

// ClientInvoicesDBID is the Notion database ID for Client Invoices
const ClientInvoicesDBID = "2bf64b29b84c80879a52ed2f9d493096"

// QueryInvoices fetches invoices from Notion Client Invoices database with filters
func (n *notionService) QueryInvoices(filter *InvoiceFilter, pagination model.Pagination) ([]nt.Page, int64, error) {
	l := n.l.Fields(logger.Fields{
		"service": "notion",
		"method":  "QueryInvoices",
		"filter":  filter,
	})

	l.Debug("querying invoices from Notion")

	// Build base filter for Invoice type only (exclude Line Items)
	filters := []nt.DatabaseQueryFilter{
		{
			Property: "Type",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Select: &nt.SelectDatabaseQueryFilter{
					Equals: "Invoice",
				},
			},
		},
	}

	l.Debugf("base filter: Type = Invoice")

	// Add project filter if provided
	if filter != nil && len(filter.ProjectIDs) > 0 {
		l.Debugf("adding project filter: %d projects", len(filter.ProjectIDs))
		projectFilters := []nt.DatabaseQueryFilter{}
		for _, projID := range filter.ProjectIDs {
			projectFilters = append(projectFilters, nt.DatabaseQueryFilter{
				Property: "Project", // Property ID: srY=
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Relation: &nt.RelationDatabaseQueryFilter{
						Contains: projID,
					},
				},
			})
		}
		if len(projectFilters) > 1 {
			filters = append(filters, nt.DatabaseQueryFilter{Or: projectFilters})
		} else {
			filters = append(filters, projectFilters[0])
		}
	}

	// Add status filter if provided
	if filter != nil && len(filter.Statuses) > 0 {
		l.Debugf("adding status filter: %d statuses", len(filter.Statuses))
		statusFilters := []nt.DatabaseQueryFilter{}
		for _, status := range filter.Statuses {
			statusFilters = append(statusFilters, nt.DatabaseQueryFilter{
				Property: "Status", // Property ID: nsb@
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Status: &nt.StatusDatabaseQueryFilter{
						Equals: status,
					},
				},
			})
		}
		if len(statusFilters) > 1 {
			filters = append(filters, nt.DatabaseQueryFilter{Or: statusFilters})
		} else {
			filters = append(filters, statusFilters[0])
		}
	}

	// Add invoice number filter if provided
	// Note: Filtering by title property not supported yet in this implementation
	// Invoice number filtering will be done post-fetch if needed
	if filter != nil && filter.InvoiceNumber != "" {
		l.Debugf("invoice number filter will be applied post-fetch: %s", filter.InvoiceNumber)
	}

	// Combine all filters with AND
	var finalFilter *nt.DatabaseQueryFilter
	if len(filters) > 1 {
		finalFilter = &nt.DatabaseQueryFilter{And: filters}
	} else if len(filters) == 1 {
		finalFilter = &filters[0]
	}

	l.Debugf("final filter constructed with %d conditions", len(filters))

	// Sort by Issue Date descending
	sorts := []nt.DatabaseQuerySort{
		{
			Property:  "Issue Date", // Property ID: U?Nt
			Direction: nt.SortDirDesc,
		},
	}

	// Determine page size - use pagination size if provided, otherwise default
	pageSize := 100
	if pagination.Size > 0 {
		pageSize = int(pagination.Size)
	}

	l.Debugf("querying Notion database: dbID=%s, pageSize=%d", ClientInvoicesDBID, pageSize)

	// Query database using existing GetDatabase method
	response, err := n.GetDatabase(ClientInvoicesDBID, finalFilter, sorts, pageSize)
	if err != nil {
		l.Error(err, "failed to query Notion database")
		return nil, 0, fmt.Errorf("failed to query Notion invoices: %w", err)
	}

	l.Debugf("fetched %d invoices from Notion", len(response.Results))

	// Return results and count
	// Note: Notion API doesn't provide total count, so we return the number of results
	// For proper pagination support, caller would need to handle HasMore and NextCursor
	total := int64(len(response.Results))

	return response.Results, total, nil
}

// GetInvoiceLineItems fetches line items for a specific invoice from Notion
func (n *notionService) GetInvoiceLineItems(invoicePageID string) ([]nt.Page, error) {
	l := n.l.Fields(logger.Fields{
		"service":       "notion",
		"method":        "GetInvoiceLineItems",
		"invoicePageID": invoicePageID,
	})

	l.Debugf("fetching line items for invoice: %s", invoicePageID)

	// Build filter for Line Items that have this invoice as parent
	filter := &nt.DatabaseQueryFilter{
		And: []nt.DatabaseQueryFilter{
			{
				Property: "Type",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Select: &nt.SelectDatabaseQueryFilter{
						Equals: "Line Item",
					},
				},
			},
			{
				Property: "Parent item", // Relation to parent invoice
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Relation: &nt.RelationDatabaseQueryFilter{
						Contains: invoicePageID,
					},
				},
			},
		},
	}

	l.Debug("querying Notion for line items")

	// Query database for line items
	response, err := n.GetDatabase(ClientInvoicesDBID, filter, nil, 100)
	if err != nil {
		l.Error(err, "failed to query line items from Notion")
		return nil, fmt.Errorf("failed to query invoice line items: %w", err)
	}

	l.Debugf("fetched %d line items for invoice %s", len(response.Results), invoicePageID)

	return response.Results, nil
}

// QueryClientInvoiceByNumber finds an invoice in Notion by its invoice number
// Returns nil, nil if not found (not an error)
func (n *notionService) QueryClientInvoiceByNumber(invoiceNumber string) (*nt.Page, error) {
	l := n.l.Fields(logger.Fields{
		"service":       "notion",
		"method":        "QueryClientInvoiceByNumber",
		"invoiceNumber": invoiceNumber,
	})

	l.Debug("searching for invoice by number in Notion")

	if invoiceNumber == "" {
		l.Debug("empty invoice number provided")
		return nil, fmt.Errorf("invoice number is required")
	}

	// Build filter for Invoice type with title containing the invoice number
	filter := &nt.DatabaseQueryFilter{
		And: []nt.DatabaseQueryFilter{
			{
				Property: "Type",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Select: &nt.SelectDatabaseQueryFilter{
						Equals: "Invoice",
					},
				},
			},
			{
				Property: "(auto) Invoice Number",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					RichText: &nt.TextPropertyFilter{
						Contains: invoiceNumber,
					},
				},
			},
		},
	}

	l.Debugf("querying Notion database: dbID=%s, filter by invoice number=%s", ClientInvoicesDBID, invoiceNumber)

	// Query database
	response, err := n.GetDatabase(ClientInvoicesDBID, filter, nil, 10)
	if err != nil {
		l.Error(err, "failed to query Notion database for invoice")
		return nil, fmt.Errorf("failed to query Notion invoice: %w", err)
	}

	l.Debugf("found %d results for invoice number %s", len(response.Results), invoiceNumber)

	// Return first result if found
	if len(response.Results) == 0 {
		l.Debug("no invoice found with this number")
		return nil, nil
	}

	// If multiple found, log warning and return first
	if len(response.Results) > 1 {
		l.Warnf("multiple invoices found with number %s, returning first", invoiceNumber)
	}

	result := &response.Results[0]
	l.Debugf("returning invoice: pageID=%s", result.ID)

	return result, nil
}

// UpdateClientInvoiceStatus updates the Status and Paid Date of an invoice in Notion
func (n *notionService) UpdateClientInvoiceStatus(pageID string, status string, paidDate *time.Time) error {
	l := n.l.Fields(logger.Fields{
		"service":  "notion",
		"method":   "UpdateClientInvoiceStatus",
		"pageID":   pageID,
		"status":   status,
		"paidDate": paidDate,
	})

	l.Debug("updating invoice status in Notion")

	if pageID == "" {
		l.Debug("empty page ID provided")
		return fmt.Errorf("page ID is required")
	}

	// Build the raw HTTP request since go-notion doesn't support status property updates well
	updatePayload := map[string]interface{}{
		"properties": map[string]interface{}{
			"Status": map[string]interface{}{
				"status": map[string]string{
					"name": status,
				},
			},
		},
	}

	// Add Paid Date if provided
	if paidDate != nil {
		l.Debugf("setting Paid Date to %s", paidDate.Format("2006-01-02"))
		updatePayload["properties"].(map[string]interface{})["Paid Date"] = map[string]interface{}{
			"date": map[string]string{
				"start": paidDate.Format("2006-01-02"),
			},
		}
	}

	payloadBytes, err := json.Marshal(updatePayload)
	if err != nil {
		l.Error(err, "failed to marshal update payload")
		return fmt.Errorf("failed to marshal update payload: %w", err)
	}

	l.Debugf("update payload: %s", string(payloadBytes))

	// Create raw HTTP request to Notion API
	notionURL := fmt.Sprintf("https://api.notion.com/v1/pages/%s", pageID)
	req, err := http.NewRequest("PATCH", notionURL, bytes.NewReader(payloadBytes))
	if err != nil {
		l.Error(err, "failed to create HTTP request for Notion update")
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+n.secret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		l.Error(err, "failed to send HTTP request to Notion")
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		l.Errorf(fmt.Errorf("notion update failed with status %d", resp.StatusCode),
			fmt.Sprintf("response body: %s", string(respBody)))
		return fmt.Errorf("Notion update failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	l.Debug("invoice status updated successfully in Notion")

	return nil
}

// ExtractClientInvoiceData extracts invoice data from a Notion page for email/GDrive operations
func (n *notionService) ExtractClientInvoiceData(page *nt.Page) (*model.Invoice, error) {
	l := n.l.Fields(logger.Fields{
		"service": "notion",
		"method":  "ExtractClientInvoiceData",
		"pageID":  page.ID,
	})

	l.Debug("extracting invoice data from Notion page")

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		l.Error(fmt.Errorf("invalid page properties"), "failed to cast page properties")
		return nil, fmt.Errorf("invalid page properties format")
	}

	invoice := &model.Invoice{}

	// Extract invoice number from title
	if titleProp, ok := props["(auto) Invoice Number"]; ok && titleProp.Title != nil {
		var titleText string
		for _, t := range titleProp.Title {
			titleText += t.PlainText
		}
		invoice.Number = titleText
		l.Debugf("extracted invoice number: %s", invoice.Number)
	}

	// Extract Month and Year from Issue Date
	if issueDateProp, ok := props["Issue Date"]; ok && issueDateProp.Date != nil {
		issueDate, err := time.Parse("2006-01-02", issueDateProp.Date.Start.Time.Format("2006-01-02"))
		if err == nil {
			invoice.Month = int(issueDate.Month())
			invoice.Year = issueDate.Year()
			l.Debugf("extracted month=%d year=%d from Issue Date", invoice.Month, invoice.Year)
		} else {
			l.Debugf("failed to parse Issue Date: %v", err)
		}
	} else {
		l.Debug("Issue Date not found in Notion properties")
	}

	// Extract status
	if statusProp, ok := props["Status"]; ok && statusProp.Status != nil {
		statusName := statusProp.Status.Name
		l.Debugf("extracted status: %s", statusName)
		// Map Notion status to model status
		switch statusName {
		case "Draft":
			invoice.Status = model.InvoiceStatusDraft
		case "Sent":
			invoice.Status = model.InvoiceStatusSent
		case "Overdue":
			invoice.Status = model.InvoiceStatusOverdue
		case "Paid":
			invoice.Status = model.InvoiceStatusPaid
		case "Canceled":
			invoice.Status = model.InvoiceStatusError
		}
	}

	// Extract recipients from rollup for email
	// Note: Rollup properties need to be fetched via GetPagePropByID for full data
	if recipientsProp, ok := props["Recipients"]; ok {
		l.Debugf("found Recipients property, type: %s", recipientsProp.Type)
		// Recipients is a rollup - we need to handle the array of rich text
		if recipientsProp.Rollup != nil && recipientsProp.Rollup.Array != nil {
			var emails []string
			for _, item := range recipientsProp.Rollup.Array {
				if item.RichText != nil {
					for _, rt := range item.RichText {
						if rt.PlainText != "" {
							emails = append(emails, rt.PlainText)
						}
					}
				}
			}
			if len(emails) > 0 {
				invoice.Email = emails[0]
				// Build CC list: remaining recipients + accounting@d.foundation
				ccList := []string{}
				if len(emails) > 1 {
					ccList = append(ccList, emails[1:]...)
				}
				// Always include accounting@d.foundation in CC
				ccList = append(ccList, "accounting@d.foundation")
				ccBytes, _ := json.Marshal(ccList)
				invoice.CC = ccBytes
				l.Debugf("extracted recipients: to=%s, cc=%v", invoice.Email, ccList)
			}
		}
	}

	// Extract Final Total
	if totalProp, ok := props["Final Total"]; ok && totalProp.Formula != nil {
		if totalProp.Formula.Number != nil {
			invoice.Total = *totalProp.Formula.Number
			l.Debugf("extracted total: %.2f", invoice.Total)
		}
	}

	// Extract Currency
	if currencyProp, ok := props["Currency"]; ok && currencyProp.Select != nil {
		l.Debugf("extracted currency: %s", currencyProp.Select.Name)
		// Currency will be used in email template
	}

	// Extract project/client name for email template
	var projectName string

	// Try extracting from Client rollup first
	if clientProp, ok := props["Client"]; ok && clientProp.Rollup != nil {
		if len(clientProp.Rollup.Array) > 0 {
			item := clientProp.Rollup.Array[0]
			if len(item.RichText) > 0 {
				projectName = item.RichText[0].PlainText
				l.Debugf("extracted client name from Client rollup: %s", projectName)
			}
		}
	}

	// Fallback to Code rollup if Client is empty
	if projectName == "" {
		if codeProp, ok := props["Code"]; ok && codeProp.Rollup != nil {
			if len(codeProp.Rollup.Array) > 0 {
				item := codeProp.Rollup.Array[0]
				// Code is a formula that returns string
				if item.Formula != nil && item.Formula.String != nil {
					projectName = *item.Formula.String
					l.Debugf("extracted project name from Code rollup: %s", projectName)
				}
			}
		}
	}

	// Fallback to Redacted Codename if Code is also empty
	if projectName == "" {
		if codenameProp, ok := props["Redacted Codename"]; ok && codenameProp.Rollup != nil {
			if len(codenameProp.Rollup.Array) > 0 {
				item := codenameProp.Rollup.Array[0]
				if item.Formula != nil && item.Formula.String != nil {
					projectName = *item.Formula.String
					l.Debugf("extracted project name from Redacted Codename rollup: %s", projectName)
				}
			}
		}
	}

	// Set project info for email template
	if projectName != "" {
		invoice.Project = &model.Project{
			Name: projectName,
		}
		l.Debugf("set project name for email: %s", projectName)
	} else {
		l.Debug("no project name found, email template may fail")
	}

	// Extract Thread ID (stored after sending invoice email)
	if threadIDProp, ok := props["Thread ID"]; ok && threadIDProp.RichText != nil {
		if len(threadIDProp.RichText) > 0 {
			invoice.ThreadID = threadIDProp.RichText[0].PlainText
			l.Debugf("extracted Thread ID: %s", invoice.ThreadID)
		} else {
			l.Debug("Thread ID property exists but is empty")
		}
	} else {
		l.Debug("Thread ID not found in Notion properties")
	}

	l.Debugf("extracted invoice data: number=%s, status=%s, email=%s, projectName=%s, total=%.2f, threadID=%s, notionPageID=%s",
		invoice.Number, invoice.Status, invoice.Email, projectName, invoice.Total, invoice.ThreadID, page.ID)

	return invoice, nil
}

// GetNotionInvoiceStatus extracts status from a Notion invoice page
func (n *notionService) GetNotionInvoiceStatus(page *nt.Page) (string, error) {
	l := n.l.Fields(logger.Fields{
		"service": "notion",
		"method":  "GetNotionInvoiceStatus",
		"pageID":  page.ID,
	})

	l.Debug("extracting status from Notion invoice page")

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		l.Error(fmt.Errorf("invalid page properties"), "failed to cast page properties")
		return "", fmt.Errorf("invalid page properties format")
	}

	if statusProp, ok := props["Status"]; ok && statusProp.Status != nil {
		status := statusProp.Status.Name
		l.Debugf("extracted status: %s", status)
		return status, nil
	}

	l.Debug("no status property found")
	return "", fmt.Errorf("status property not found")
}
