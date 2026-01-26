package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

// ClientInvoicesDBID is the Notion database ID for Client Invoices
const ClientInvoicesDBID = "2bf64b29b84c80879a52ed2f9d493096"

// notionAPIVersion is the Notion API version used for raw HTTP requests
// Using 2022-06-28 for page updates as it's stable and widely compatible
// Note: 2025-09-03 introduces breaking changes for database operations
const notionAPIVersion = "2022-06-28"

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

	// Build filter for Invoice type with Legacy Number containing the invoice number
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
				Property: "Legacy Number",
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
	req.Header.Set("Notion-Version", notionAPIVersion)

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

// UpdateLineItemsStatus updates the Status of all line items for a given invoice
// Uses concurrent updates with semaphore to prevent API overload
func (n *notionService) UpdateLineItemsStatus(invoicePageID string, status string) error {
	l := n.l.Fields(logger.Fields{
		"service":       "notion",
		"method":        "UpdateLineItemsStatus",
		"invoicePageID": invoicePageID,
		"status":        status,
	})

	l.Debug("updating line items status in Notion")

	if invoicePageID == "" {
		l.Debug("empty invoice page ID provided")
		return fmt.Errorf("invoice page ID is required")
	}

	// Fetch all line items for this invoice
	lineItems, err := n.GetInvoiceLineItems(invoicePageID)
	if err != nil {
		l.Errorf(err, "failed to fetch line items for invoice")
		return fmt.Errorf("failed to fetch line items: %w", err)
	}

	l.Debugf("found %d line items to update", len(lineItems))

	if len(lineItems) == 0 {
		l.Debug("no line items found for invoice")
		return nil
	}

	// Concurrent updates with semaphore (like task order log pattern)
	type updateResult struct {
		pageID string
		err    error
	}

	resultsChan := make(chan updateResult, len(lineItems))
	var wg sync.WaitGroup

	// Concurrency limit to prevent API overload
	const maxConcurrent = 5
	sem := make(chan struct{}, maxConcurrent)

	httpClient := &http.Client{Timeout: 30 * time.Second}

	for _, lineItem := range lineItems {
		wg.Add(1)
		go func(pageID string) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			l.Debugf("updating line item %s status to %s", pageID, status)

			// Build update payload - Line Items use Status type (same as Invoice)
			updatePayload := map[string]any{
				"properties": map[string]any{
					"Status": map[string]any{
						"status": map[string]string{
							"name": status,
						},
					},
				},
			}

			payloadBytes, err := json.Marshal(updatePayload)
			if err != nil {
				resultsChan <- updateResult{pageID: pageID, err: fmt.Errorf("marshal error: %w", err)}
				return
			}

			// Create raw HTTP request to Notion API
			// Note: Using raw HTTP because go-notion doesn't support status property type well
			notionURL := fmt.Sprintf("https://api.notion.com/v1/pages/%s", pageID)
			req, err := http.NewRequest("PATCH", notionURL, bytes.NewReader(payloadBytes))
			if err != nil {
				resultsChan <- updateResult{pageID: pageID, err: fmt.Errorf("request error: %w", err)}
				return
			}

			req.Header.Set("Authorization", "Bearer "+n.secret)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Notion-Version", notionAPIVersion)

			resp, err := httpClient.Do(req)
			if err != nil {
				resultsChan <- updateResult{pageID: pageID, err: fmt.Errorf("http error: %w", err)}
				return
			}

			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				resultsChan <- updateResult{pageID: pageID, err: fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody))}
				return
			}

			l.Debugf("line item %s status updated successfully", pageID)
			resultsChan <- updateResult{pageID: pageID, err: nil}
		}(lineItem.ID)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	var updateErrors []string
	for result := range resultsChan {
		if result.err != nil {
			l.Errorf(result.err, "failed to update line item %s", result.pageID)
			updateErrors = append(updateErrors, fmt.Sprintf("%s: %v", result.pageID, result.err))
		}
	}

	if len(updateErrors) > 0 {
		l.Warnf("some line items failed to update: %v", updateErrors)
		return fmt.Errorf("failed to update %d/%d line items", len(updateErrors), len(lineItems))
	}

	l.Infof("successfully updated %d line items status to %s", len(lineItems), status)
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

	// Extract invoice number from Legacy Number
	if legacyNumberProp, ok := props["Legacy Number"]; ok && legacyNumberProp.RichText != nil {
		var invoiceNumber string
		for _, t := range legacyNumberProp.RichText {
			invoiceNumber += t.PlainText
		}
		invoice.Number = invoiceNumber
		l.Debugf("extracted invoice number from Legacy Number: %s", invoice.Number)
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

// LineItemCommissionData contains commission data extracted from a line item
type LineItemCommissionData struct {
	PageID           string
	DeploymentPageID string

	// Commission percentages
	SalesPercent        float64
	AccountMgrPercent   float64
	DeliveryLeadPercent float64
	HiringRefPercent    float64

	// Calculated amounts
	SalesAmount        float64
	AccountMgrAmount   float64
	DeliveryLeadAmount float64
	HiringRefAmount    float64

	// Person page IDs (from rollups - these are Contractor page IDs)
	SalesPersonIDs    []string
	AccountMgrIDs     []string
	DeliveryLeadIDs   []string
	HiringRefIDs      []string

	// Metadata
	Currency    string
	Month       time.Time
	ProjectCode string
}

// QueryLineItemsWithCommissions fetches line items with commission data for a given invoice
func (n *notionService) QueryLineItemsWithCommissions(invoicePageID string) ([]LineItemCommissionData, error) {
	l := n.l.Fields(logger.Fields{
		"service":       "notion",
		"method":        "QueryLineItemsWithCommissions",
		"invoicePageID": invoicePageID,
	})

	l.Debug("querying line items with commission data")

	// First, fetch the Invoice page to get Currency and Issue Date
	// (These are on the Invoice, not the Line Items)
	ctx := context.Background()
	invoicePage, err := n.notionClient.FindPageByID(ctx, invoicePageID)
	if err != nil {
		l.Error(err, "failed to fetch invoice page")
		return nil, fmt.Errorf("failed to fetch invoice page: %w", err)
	}

	var invoiceCurrency string
	var invoiceMonth time.Time

	if invoiceProps, ok := invoicePage.Properties.(nt.DatabasePageProperties); ok {
		// Extract Currency from Invoice
		if currencyProp, ok := invoiceProps["Currency"]; ok && currencyProp.Select != nil {
			invoiceCurrency = currencyProp.Select.Name
			l.Debugf("extracted invoice currency: %s", invoiceCurrency)
		}

		// Extract Issue Date from Invoice
		if issueDateProp, ok := invoiceProps["Issue Date"]; ok && issueDateProp.Date != nil {
			invoiceMonth = issueDateProp.Date.Start.Time
			l.Debugf("extracted invoice month: %s", invoiceMonth.Format("2006-01"))
		}
	}

	// Get line items for the invoice
	lineItems, err := n.GetInvoiceLineItems(invoicePageID)
	if err != nil {
		l.Error(err, "failed to get invoice line items")
		return nil, fmt.Errorf("failed to get line items: %w", err)
	}

	l.Debugf("fetched %d line items", len(lineItems))

	var results []LineItemCommissionData

	for _, item := range lineItems {
		props, ok := item.Properties.(nt.DatabasePageProperties)
		if !ok {
			l.Debugf("failed to cast properties for line item %s", item.ID)
			continue
		}

		data := LineItemCommissionData{
			PageID: item.ID,
		}

		// Extract Deployment Tracker relation
		if deploymentProp, ok := props["Deployment Tracker"]; ok && deploymentProp.Relation != nil {
			if len(deploymentProp.Relation) > 0 {
				data.DeploymentPageID = deploymentProp.Relation[0].ID
				l.Debugf("extracted deployment page ID: %s", data.DeploymentPageID)
			}
		}

		// Extract commission percentages
		data.SalesPercent = n.extractNumberProp(props, "% Sales")
		data.AccountMgrPercent = n.extractNumberProp(props, "% Account Mgr")
		data.DeliveryLeadPercent = n.extractNumberProp(props, "% Delivery Lead")
		data.HiringRefPercent = n.extractNumberProp(props, "% Hiring Referral")

		l.Debugf("commission percentages: sales=%.2f%%, am=%.2f%%, dl=%.2f%%, hr=%.2f%%",
			data.SalesPercent*100, data.AccountMgrPercent*100, data.DeliveryLeadPercent*100, data.HiringRefPercent*100)

		// Extract calculated commission amounts from formulas
		data.SalesAmount = n.extractFormulaProp(props, "Sales Amount")
		data.AccountMgrAmount = n.extractFormulaProp(props, "Account Amount")
		data.DeliveryLeadAmount = n.extractFormulaProp(props, "Delivery Lead Amount")
		data.HiringRefAmount = n.extractFormulaProp(props, "Hiring Referral Amount")

		l.Debugf("commission amounts: sales=%.2f, am=%.2f, dl=%.2f, hr=%.2f",
			data.SalesAmount, data.AccountMgrAmount, data.DeliveryLeadAmount, data.HiringRefAmount)

		// Fetch Deployment Tracker page to get contractor IDs
		if data.DeploymentPageID != "" {
			contractorIDs, err := n.getContractorIDsFromDeployment(data.DeploymentPageID)
			if err != nil {
				l.Debugf("failed to get contractor IDs from deployment: %v", err)
			} else {
				data.SalesPersonIDs = contractorIDs.SalesIDs
				data.AccountMgrIDs = contractorIDs.AccountMgrIDs
				data.DeliveryLeadIDs = contractorIDs.DeliveryLeadIDs
				data.HiringRefIDs = contractorIDs.HiringRefIDs
			}
		}

		l.Debugf("person IDs: sales=%v, am=%v, dl=%v, hr=%v",
			data.SalesPersonIDs, data.AccountMgrIDs, data.DeliveryLeadIDs, data.HiringRefIDs)

		// Use Currency and Month from Invoice (already extracted above)
		data.Currency = invoiceCurrency
		data.Month = invoiceMonth
		l.Debugf("using invoice currency=%s month=%s", data.Currency, data.Month.Format("2006-01"))

		// Extract Project Code from rollup
		if codeProp, ok := props["Code"]; ok && codeProp.Rollup != nil {
			if len(codeProp.Rollup.Array) > 0 {
				item := codeProp.Rollup.Array[0]
				if item.Formula != nil && item.Formula.String != nil {
					data.ProjectCode = *item.Formula.String
					l.Debugf("extracted project code: %s", data.ProjectCode)
				}
			}
		}

		results = append(results, data)
	}

	l.Debugf("processed %d line items with commission data", len(results))

	return results, nil
}

// DeploymentContractorIDs holds contractor IDs extracted from a Deployment Tracker page
type DeploymentContractorIDs struct {
	SalesIDs        []string
	AccountMgrIDs   []string
	DeliveryLeadIDs []string
	HiringRefIDs    []string
}

// getContractorIDsFromDeployment fetches the Deployment Tracker page and extracts contractor IDs
func (n *notionService) getContractorIDsFromDeployment(deploymentPageID string) (*DeploymentContractorIDs, error) {
	l := n.l.Fields(logger.Fields{
		"service":          "notion",
		"method":           "getContractorIDsFromDeployment",
		"deploymentPageID": deploymentPageID,
	})

	l.Debug("fetching deployment tracker page for contractor IDs")

	ctx := context.Background()
	page, err := n.notionClient.FindPageByID(ctx, deploymentPageID)
	if err != nil {
		l.Error(err, "failed to fetch deployment tracker page")
		return nil, fmt.Errorf("failed to fetch deployment page: %w", err)
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		l.Debug("failed to cast deployment page properties")
		return nil, fmt.Errorf("invalid deployment page properties")
	}

	result := &DeploymentContractorIDs{}

	// Extract Original Sales rollup → relation IDs
	if prop, ok := props["Original Sales"]; ok && prop.Rollup != nil && prop.Rollup.Array != nil {
		for _, item := range prop.Rollup.Array {
			if item.Relation != nil {
				for _, rel := range item.Relation {
					result.SalesIDs = append(result.SalesIDs, rel.ID)
				}
			}
		}
	}
	l.Debugf("extracted sales IDs: %v", result.SalesIDs)

	// Extract Account Managers rollup → relation IDs
	if prop, ok := props["Account Managers"]; ok && prop.Rollup != nil && prop.Rollup.Array != nil {
		for _, item := range prop.Rollup.Array {
			if item.Relation != nil {
				for _, rel := range item.Relation {
					result.AccountMgrIDs = append(result.AccountMgrIDs, rel.ID)
				}
			}
		}
	}
	l.Debugf("extracted AM IDs: %v", result.AccountMgrIDs)

	// Extract Delivery Leads rollup → relation IDs
	if prop, ok := props["Delivery Leads"]; ok && prop.Rollup != nil && prop.Rollup.Array != nil {
		for _, item := range prop.Rollup.Array {
			if item.Relation != nil {
				for _, rel := range item.Relation {
					result.DeliveryLeadIDs = append(result.DeliveryLeadIDs, rel.ID)
				}
			}
		}
	}
	l.Debugf("extracted DL IDs: %v", result.DeliveryLeadIDs)

	// Extract Hiring Referral - check for a relation property
	// Note: Hiring Referral might be in Project, not Deployment Tracker
	// For now, check if there's a Hiring Referral relation
	if prop, ok := props["Hiring Referral"]; ok {
		if prop.Relation != nil {
			for _, rel := range prop.Relation {
				result.HiringRefIDs = append(result.HiringRefIDs, rel.ID)
			}
		} else if prop.Rollup != nil && prop.Rollup.Array != nil {
			for _, item := range prop.Rollup.Array {
				if item.Relation != nil {
					for _, rel := range item.Relation {
						result.HiringRefIDs = append(result.HiringRefIDs, rel.ID)
					}
				}
			}
		}
	}
	l.Debugf("extracted HR IDs: %v", result.HiringRefIDs)

	return result, nil
}

// extractNumberProp extracts a number property value
func (n *notionService) extractNumberProp(props nt.DatabasePageProperties, propName string) float64 {
	if prop, ok := props[propName]; ok && prop.Number != nil {
		return *prop.Number
	}
	return 0
}

// extractFormulaProp extracts a formula property number value
func (n *notionService) extractFormulaProp(props nt.DatabasePageProperties, propName string) float64 {
	if prop, ok := props[propName]; ok && prop.Formula != nil && prop.Formula.Number != nil {
		return *prop.Formula.Number
	}
	return 0
}

// IsSplitsGenerated checks if splits have already been generated for an invoice
func (n *notionService) IsSplitsGenerated(invoicePageID string) (bool, error) {
	l := n.l.Fields(logger.Fields{
		"service":       "notion",
		"method":        "IsSplitsGenerated",
		"invoicePageID": invoicePageID,
	})

	l.Debug("checking if splits already generated")

	ctx := context.Background()
	page, err := n.notionClient.FindPageByID(ctx, invoicePageID)
	if err != nil {
		l.Error(err, "failed to fetch invoice page")
		return false, fmt.Errorf("failed to fetch invoice page: %w", err)
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		l.Error(fmt.Errorf("invalid page properties"), "failed to cast page properties")
		return false, fmt.Errorf("invalid page properties format")
	}

	// Check "Splits Generated" checkbox property
	if prop, ok := props["Splits Generated"]; ok && prop.Checkbox != nil {
		isGenerated := *prop.Checkbox
		l.Debugf("splits generated: %v", isGenerated)
		return isGenerated, nil
	}

	l.Debug("Splits Generated property not found, assuming false")
	return false, nil
}

// MarkSplitsGenerated updates the "Splits Generated" checkbox to true
func (n *notionService) MarkSplitsGenerated(invoicePageID string) error {
	l := n.l.Fields(logger.Fields{
		"service":       "notion",
		"method":        "MarkSplitsGenerated",
		"invoicePageID": invoicePageID,
	})

	l.Debug("marking splits as generated")

	if invoicePageID == "" {
		l.Debug("empty invoice page ID provided")
		return fmt.Errorf("invoice page ID is required")
	}

	// Build the update payload for checkbox property
	updatePayload := map[string]interface{}{
		"properties": map[string]interface{}{
			"Splits Generated": map[string]interface{}{
				"checkbox": true,
			},
		},
	}

	payloadBytes, err := json.Marshal(updatePayload)
	if err != nil {
		l.Error(err, "failed to marshal update payload")
		return fmt.Errorf("failed to marshal update payload: %w", err)
	}

	l.Debugf("update payload: %s", string(payloadBytes))

	// Create raw HTTP request to Notion API
	notionURL := fmt.Sprintf("https://api.notion.com/v1/pages/%s", invoicePageID)
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

	l.Debug("splits generated flag updated successfully")

	return nil
}

// MarkLineItemsSplitsGenerated updates the "Splits Generated" checkbox to true for all line items of an invoice
// Uses concurrent updates with semaphore to prevent API overload
func (n *notionService) MarkLineItemsSplitsGenerated(invoicePageID string) error {
	l := n.l.Fields(logger.Fields{
		"service":       "notion",
		"method":        "MarkLineItemsSplitsGenerated",
		"invoicePageID": invoicePageID,
	})

	l.Debug("marking line items splits as generated")

	if invoicePageID == "" {
		l.Debug("empty invoice page ID provided")
		return fmt.Errorf("invoice page ID is required")
	}

	// Fetch all line items for this invoice
	lineItems, err := n.GetInvoiceLineItems(invoicePageID)
	if err != nil {
		l.Errorf(err, "failed to fetch line items for invoice")
		return fmt.Errorf("failed to fetch line items: %w", err)
	}

	l.Debugf("found %d line items to mark splits generated", len(lineItems))

	if len(lineItems) == 0 {
		l.Debug("no line items found for invoice")
		return nil
	}

	// Concurrent updates with semaphore
	type updateResult struct {
		pageID string
		err    error
	}

	resultsChan := make(chan updateResult, len(lineItems))
	var wg sync.WaitGroup

	// Concurrency limit to prevent API overload
	const maxConcurrent = 5
	sem := make(chan struct{}, maxConcurrent)

	httpClient := &http.Client{Timeout: 30 * time.Second}

	for _, lineItem := range lineItems {
		wg.Add(1)
		go func(pageID string) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			l.Debugf("marking line item %s splits generated", pageID)

			// Build update payload for checkbox property
			updatePayload := map[string]any{
				"properties": map[string]any{
					"Splits Generated": map[string]any{
						"checkbox": true,
					},
				},
			}

			payloadBytes, err := json.Marshal(updatePayload)
			if err != nil {
				resultsChan <- updateResult{pageID: pageID, err: fmt.Errorf("marshal error: %w", err)}
				return
			}

			// Create raw HTTP request to Notion API
			notionURL := fmt.Sprintf("https://api.notion.com/v1/pages/%s", pageID)
			req, err := http.NewRequest("PATCH", notionURL, bytes.NewReader(payloadBytes))
			if err != nil {
				resultsChan <- updateResult{pageID: pageID, err: fmt.Errorf("request error: %w", err)}
				return
			}

			req.Header.Set("Authorization", "Bearer "+n.secret)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Notion-Version", notionAPIVersion)

			resp, err := httpClient.Do(req)
			if err != nil {
				resultsChan <- updateResult{pageID: pageID, err: fmt.Errorf("http error: %w", err)}
				return
			}

			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				resultsChan <- updateResult{pageID: pageID, err: fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody))}
				return
			}

			l.Debugf("line item %s splits generated marked successfully", pageID)
			resultsChan <- updateResult{pageID: pageID, err: nil}
		}(lineItem.ID)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	var updateErrors []string
	for result := range resultsChan {
		if result.err != nil {
			l.Errorf(result.err, "failed to mark line item %s splits generated", result.pageID)
			updateErrors = append(updateErrors, fmt.Sprintf("%s: %v", result.pageID, result.err))
		}
	}

	if len(updateErrors) > 0 {
		l.Warnf("some line items failed to mark splits generated: %v", updateErrors)
		return fmt.Errorf("failed to mark %d/%d line items splits generated", len(updateErrors), len(lineItems))
	}

	l.Infof("successfully marked %d line items splits generated", len(lineItems))
	return nil
}

// QueryInvoicesByMonth fetches invoices from Notion for a specific month with filters
func (n *notionService) QueryInvoicesByMonth(year, month int, statuses []string, projectID string) ([]nt.Page, error) {
	l := n.l.Fields(logger.Fields{
		"service":   "notion",
		"method":    "QueryInvoicesByMonth",
		"year":      year,
		"month":     month,
		"statuses":  statuses,
		"projectID": projectID,
	})

	l.Debug("querying invoices by month from Notion")

	// Calculate first and last day of the month
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)

	// Build base filters
	filters := []nt.DatabaseQueryFilter{
		// Type = Invoice (exclude Line Items)
		{
			Property: "Type",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Select: &nt.SelectDatabaseQueryFilter{
					Equals: "Invoice",
				},
			},
		},
		// Issue Date >= first day of month
		{
			Property: "Issue Date",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Date: &nt.DatePropertyFilter{
					OnOrAfter: &firstDay,
				},
			},
		},
		// Issue Date <= last day of month
		{
			Property: "Issue Date",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Date: &nt.DatePropertyFilter{
					OnOrBefore: &lastDay,
				},
			},
		},
	}

	// Add status filter if provided
	if len(statuses) > 0 {
		l.Debugf("adding status filter: %v", statuses)
		statusFilters := []nt.DatabaseQueryFilter{}
		for _, status := range statuses {
			statusFilters = append(statusFilters, nt.DatabaseQueryFilter{
				Property: "Status",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Status: &nt.StatusDatabaseQueryFilter{
						Equals: status,
					},
				},
			})
		}
		if len(statusFilters) > 1 {
			filters = append(filters, nt.DatabaseQueryFilter{Or: statusFilters})
		} else if len(statusFilters) == 1 {
			filters = append(filters, statusFilters[0])
		}
	}

	// Add project filter if provided
	if projectID != "" {
		l.Debugf("adding project filter: %s", projectID)
		filters = append(filters, nt.DatabaseQueryFilter{
			Property: "Project",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Relation: &nt.RelationDatabaseQueryFilter{
					Contains: projectID,
				},
			},
		})
	}

	// Combine all filters with AND
	finalFilter := &nt.DatabaseQueryFilter{And: filters}

	l.Debugf("querying Notion database with %d filter conditions", len(filters))

	// Query database
	response, err := n.GetDatabase(ClientInvoicesDBID, finalFilter, nil, 100)
	if err != nil {
		l.Error(err, "failed to query Notion database")
		return nil, fmt.Errorf("failed to query invoices by month: %w", err)
	}

	l.Debugf("fetched %d invoices for %d-%02d", len(response.Results), year, month)

	return response.Results, nil
}

// CloneInvoiceToNextMonth clones an invoice and its line items to a new month
func (n *notionService) CloneInvoiceToNextMonth(sourceInvoicePageID string, targetIssueDate time.Time) (*ClonedInvoiceResult, error) {
	l := n.l.Fields(logger.Fields{
		"service":             "notion",
		"method":              "CloneInvoiceToNextMonth",
		"sourceInvoicePageID": sourceInvoicePageID,
		"targetIssueDate":     targetIssueDate.Format("2006-01-02"),
	})

	l.Debug("cloning invoice to next month")

	// Fetch source invoice page
	ctx := context.Background()
	sourcePage, err := n.notionClient.FindPageByID(ctx, sourceInvoicePageID)
	if err != nil {
		l.Error(err, "failed to fetch source invoice page")
		return nil, fmt.Errorf("failed to fetch source invoice: %w", err)
	}

	sourceProps, ok := sourcePage.Properties.(nt.DatabasePageProperties)
	if !ok {
		l.Error(fmt.Errorf("invalid properties"), "failed to cast source invoice properties")
		return nil, fmt.Errorf("invalid source invoice properties")
	}

	// Extract source invoice data
	var projectRelationID string
	if projectProp, ok := sourceProps["Project"]; ok && projectProp.Relation != nil && len(projectProp.Relation) > 0 {
		projectRelationID = projectProp.Relation[0].ID
	}
	if projectRelationID == "" {
		return nil, fmt.Errorf("source invoice has no project relation")
	}

	// Extract other properties to clone
	var currency, discountType, billingType, paymentMethod string
	var taxRate, discountValue float64
	var bankAccountID string

	if currencyProp, ok := sourceProps["Currency"]; ok && currencyProp.Select != nil {
		currency = currencyProp.Select.Name
	}
	if taxRateProp, ok := sourceProps["Tax Rate"]; ok && taxRateProp.Number != nil {
		taxRate = *taxRateProp.Number
	}
	if discountTypeProp, ok := sourceProps["Discount Type"]; ok && discountTypeProp.Select != nil {
		discountType = discountTypeProp.Select.Name
	}
	if discountValueProp, ok := sourceProps["Discount Value"]; ok && discountValueProp.Number != nil {
		discountValue = *discountValueProp.Number
	}
	if bankProp, ok := sourceProps["Bank Account"]; ok && bankProp.Relation != nil && len(bankProp.Relation) > 0 {
		bankAccountID = bankProp.Relation[0].ID
	}
	if paymentMethodProp, ok := sourceProps["Payment Method"]; ok && paymentMethodProp.Select != nil {
		paymentMethod = paymentMethodProp.Select.Name
	}
	if billingTypeProp, ok := sourceProps["Billing Type"]; ok && billingTypeProp.Select != nil {
		billingType = billingTypeProp.Select.Name
	}

	// Extract description and notes
	var description, notes string
	if descProp, ok := sourceProps["Description"]; ok && len(descProp.RichText) > 0 {
		for _, rt := range descProp.RichText {
			description += rt.PlainText
		}
	}
	if notesProp, ok := sourceProps["Notes"]; ok && len(notesProp.RichText) > 0 {
		for _, rt := range notesProp.RichText {
			notes += rt.PlainText
		}
	}

	l.Debugf("source invoice data: project=%s, currency=%s, taxRate=%.2f, discountType=%s, discountValue=%.2f",
		projectRelationID, currency, taxRate, discountType, discountValue)

	// Build new invoice properties
	newInvoiceProps := map[string]interface{}{
		"properties": map[string]interface{}{
			"Type": map[string]interface{}{
				"select": map[string]string{
					"name": "Invoice",
				},
			},
			"Status": map[string]interface{}{
				"status": map[string]string{
					"name": "Draft",
				},
			},
			"Issue Date": map[string]interface{}{
				"date": map[string]string{
					"start": targetIssueDate.Format("2006-01-02"),
				},
			},
			"Project": map[string]interface{}{
				"relation": []map[string]string{
					{"id": projectRelationID},
				},
			},
			"Splits Generated": map[string]interface{}{
				"checkbox": false,
			},
		},
		"parent": map[string]interface{}{
			"database_id": ClientInvoicesDBID,
		},
	}

	propsMap := newInvoiceProps["properties"].(map[string]interface{})

	// Add optional properties if they exist
	if currency != "" {
		propsMap["Currency"] = map[string]interface{}{
			"select": map[string]string{
				"name": currency,
			},
		}
	}
	if taxRate > 0 {
		propsMap["Tax Rate"] = map[string]interface{}{
			"number": taxRate,
		}
	}
	if discountType != "" && discountType != "None" {
		propsMap["Discount Type"] = map[string]interface{}{
			"select": map[string]string{
				"name": discountType,
			},
		}
	}
	if discountValue > 0 {
		propsMap["Discount Value"] = map[string]interface{}{
			"number": discountValue,
		}
	}
	if bankAccountID != "" {
		propsMap["Bank Account"] = map[string]interface{}{
			"relation": []map[string]string{
				{"id": bankAccountID},
			},
		}
	}
	if paymentMethod != "" {
		propsMap["Payment Method"] = map[string]interface{}{
			"select": map[string]string{
				"name": paymentMethod,
			},
		}
	}
	if billingType != "" {
		propsMap["Billing Type"] = map[string]interface{}{
			"select": map[string]string{
				"name": billingType,
			},
		}
	}
	if description != "" {
		propsMap["Description"] = map[string]interface{}{
			"rich_text": []map[string]interface{}{
				{
					"type": "text",
					"text": map[string]string{
						"content": description,
					},
				},
			},
		}
	}
	if notes != "" {
		propsMap["Notes"] = map[string]interface{}{
			"rich_text": []map[string]interface{}{
				{
					"type": "text",
					"text": map[string]string{
						"content": notes,
					},
				},
			},
		}
	}

	// Create new invoice page via raw HTTP request
	payloadBytes, err := json.Marshal(newInvoiceProps)
	if err != nil {
		l.Error(err, "failed to marshal new invoice payload")
		return nil, fmt.Errorf("failed to marshal new invoice payload: %w", err)
	}

	l.Debugf("creating new invoice page with payload: %s", string(payloadBytes))

	req, err := http.NewRequest("POST", "https://api.notion.com/v1/pages", bytes.NewReader(payloadBytes))
	if err != nil {
		l.Error(err, "failed to create HTTP request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+n.secret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		l.Error(err, "failed to send HTTP request to Notion")
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		l.Errorf(fmt.Errorf("notion create failed with status %d", resp.StatusCode),
			fmt.Sprintf("response body: %s", string(respBody)))
		return nil, fmt.Errorf("Notion create failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	// Parse response to get new page ID
	var createResp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respBody, &createResp); err != nil {
		l.Error(err, "failed to parse create response")
		return nil, fmt.Errorf("failed to parse create response: %w", err)
	}

	newInvoicePageID := createResp.ID
	l.Debugf("created new invoice page: %s", newInvoicePageID)

	// Fetch and clone line items
	sourceLineItems, err := n.GetInvoiceLineItems(sourceInvoicePageID)
	if err != nil {
		l.Error(err, "failed to get source line items")
		// Continue without line items - invoice is already created
		return &ClonedInvoiceResult{
			NewInvoicePageID: newInvoicePageID,
			LineItemsCloned:  0,
		}, nil
	}

	lineItemsCloned := 0
	for _, sourceItem := range sourceLineItems {
		itemProps, ok := sourceItem.Properties.(nt.DatabasePageProperties)
		if !ok {
			l.Debugf("failed to cast line item properties: %s", sourceItem.ID)
			continue
		}

		// Clone line item
		err := n.cloneLineItem(itemProps, newInvoicePageID)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to clone line item: %s", sourceItem.ID))
			continue
		}
		lineItemsCloned++
	}

	l.Debugf("cloned %d line items", lineItemsCloned)

	return &ClonedInvoiceResult{
		NewInvoicePageID: newInvoicePageID,
		LineItemsCloned:  lineItemsCloned,
	}, nil
}

// cloneLineItem creates a new line item linked to a new invoice
func (n *notionService) cloneLineItem(sourceProps nt.DatabasePageProperties, newInvoicePageID string) error {
	l := n.l.Fields(logger.Fields{
		"service":          "notion",
		"method":           "cloneLineItem",
		"newInvoicePageID": newInvoicePageID,
	})

	// Build new line item properties
	newLineItemProps := map[string]interface{}{
		"properties": map[string]interface{}{
			"Type": map[string]interface{}{
				"select": map[string]string{
					"name": "Line Item",
				},
			},
			"Parent item": map[string]interface{}{
				"relation": []map[string]string{
					{"id": newInvoicePageID},
				},
			},
		},
		"parent": map[string]interface{}{
			"database_id": ClientInvoicesDBID,
		},
	}

	propsMap := newLineItemProps["properties"].(map[string]interface{})

	// Copy Quantity
	if quantityProp, ok := sourceProps["Quantity"]; ok && quantityProp.Number != nil {
		propsMap["Quantity"] = map[string]interface{}{
			"number": *quantityProp.Number,
		}
	}

	// Copy Unit Price
	if unitPriceProp, ok := sourceProps["Unit Price"]; ok && unitPriceProp.Number != nil {
		propsMap["Unit Price"] = map[string]interface{}{
			"number": *unitPriceProp.Number,
		}
	}

	// Copy Fixed Amount
	if fixedAmountProp, ok := sourceProps["Fixed Amount"]; ok && fixedAmountProp.Number != nil {
		propsMap["Fixed Amount"] = map[string]interface{}{
			"number": *fixedAmountProp.Number,
		}
	}

	// Copy Discount Type
	if discountTypeProp, ok := sourceProps["Discount Type"]; ok && discountTypeProp.Select != nil {
		propsMap["Discount Type"] = map[string]interface{}{
			"select": map[string]string{
				"name": discountTypeProp.Select.Name,
			},
		}
	}

	// Copy Discount Value
	if discountValueProp, ok := sourceProps["Discount Value"]; ok && discountValueProp.Number != nil {
		propsMap["Discount Value"] = map[string]interface{}{
			"number": *discountValueProp.Number,
		}
	}

	// Copy Description
	if descProp, ok := sourceProps["Description"]; ok && len(descProp.RichText) > 0 {
		var desc string
		for _, rt := range descProp.RichText {
			desc += rt.PlainText
		}
		if desc != "" {
			propsMap["Description"] = map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{
						"type": "text",
						"text": map[string]string{
							"content": desc,
						},
					},
				},
			}
		}
	}

	// Copy Deployment Tracker relation if exists
	if deploymentProp, ok := sourceProps["Deployment Tracker"]; ok && deploymentProp.Relation != nil && len(deploymentProp.Relation) > 0 {
		relations := []map[string]string{}
		for _, rel := range deploymentProp.Relation {
			relations = append(relations, map[string]string{"id": rel.ID})
		}
		propsMap["Deployment Tracker"] = map[string]interface{}{
			"relation": relations,
		}
	}

	// Copy commission percentages
	commissionFields := []string{"% Sales", "% Account Mgr", "% Delivery Lead", "% Hiring Referral"}
	for _, field := range commissionFields {
		if prop, ok := sourceProps[field]; ok && prop.Number != nil {
			propsMap[field] = map[string]interface{}{
				"number": *prop.Number,
			}
		}
	}

	// Create line item via raw HTTP request
	payloadBytes, err := json.Marshal(newLineItemProps)
	if err != nil {
		return fmt.Errorf("failed to marshal line item payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.notion.com/v1/pages", bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+n.secret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		l.Errorf(fmt.Errorf("notion create failed with status %d", resp.StatusCode),
			fmt.Sprintf("response body: %s", string(respBody)))
		return fmt.Errorf("Notion create failed: status=%d", resp.StatusCode)
	}

	return nil
}

// CheckInvoiceExistsForMonth checks if an invoice already exists for a project in a given month
func (n *notionService) CheckInvoiceExistsForMonth(projectPageID string, year, month int) (bool, string, error) {
	l := n.l.Fields(logger.Fields{
		"service":       "notion",
		"method":        "CheckInvoiceExistsForMonth",
		"projectPageID": projectPageID,
		"year":          year,
		"month":         month,
	})

	l.Debug("checking if invoice exists for project in month")

	// Calculate first and last day of the month
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)

	// Build filter
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
				Property: "Project",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Relation: &nt.RelationDatabaseQueryFilter{
						Contains: projectPageID,
					},
				},
			},
			{
				Property: "Issue Date",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Date: &nt.DatePropertyFilter{
						OnOrAfter: &firstDay,
					},
				},
			},
			{
				Property: "Issue Date",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Date: &nt.DatePropertyFilter{
						OnOrBefore: &lastDay,
					},
				},
			},
		},
	}

	// Query database
	response, err := n.GetDatabase(ClientInvoicesDBID, filter, nil, 1)
	if err != nil {
		l.Error(err, "failed to query Notion database")
		return false, "", fmt.Errorf("failed to check invoice existence: %w", err)
	}

	if len(response.Results) > 0 {
		existingID := response.Results[0].ID
		l.Debugf("found existing invoice: %s", existingID)
		return true, existingID, nil
	}

	l.Debug("no existing invoice found")
	return false, "", nil
}
