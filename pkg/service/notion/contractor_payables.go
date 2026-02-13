package notion

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

// ContractorPayablesService handles contractor payables operations with Notion
type ContractorPayablesService struct {
	client    *nt.Client
	cfg       *config.Config
	logger    logger.Logger
	notionSvc IService // For file upload operations
}

// CreatePayableInput contains the input data for creating a new contractor payable
type CreatePayableInput struct {
	ContractorPageID string   // Relation to Contractor (required)
	Total            float64  // Total amount in USD (required)
	Currency         string   // "USD" or "VND" (required)
	PeriodStart      string   // YYYY-MM-DD start of period - payday of invoice month (required)
	PeriodEnd        string   // YYYY-MM-DD end of period - payday of next month (required)
	InvoiceDate      string   // YYYY-MM-DD (required)
	InvoiceID        string   // Invoice number e.g., INVC-202512-QUANG-A1B2 (required)
	PayoutItemIDs    []string // Relation to Payout Items (required)
	ContractorType   string   // "Individual", "Sole Proprietor", "LLC", etc. (optional, defaults to "Individual")
	ExchangeRate     float64  // Exchange rate for currency conversion (optional)
	PDFBytes         []byte   // PDF file bytes to upload to Notion (optional)
}

// NewContractorPayablesService creates a new Notion contractor payables service
func NewContractorPayablesService(cfg *config.Config, logger logger.Logger, notionSvc IService) *ContractorPayablesService {
	if cfg.Notion.Secret == "" {
		logger.Error(errors.New("notion secret not configured"), "notion secret is empty")
		return nil
	}

	logger.Debug("[DEBUG] contractor_payables: creating new ContractorPayablesService")

	return &ContractorPayablesService{
		client:    nt.NewClient(cfg.Notion.Secret),
		cfg:       cfg,
		logger:    logger,
		notionSvc: notionSvc,
	}
}

// ExistingPayable represents an existing payable record found in Notion
type ExistingPayable struct {
	PageID string
	Status string // "New", "Pending", etc.
}

// hasOverlap checks if any element in slice a exists in slice b
func hasOverlap(a, b []string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	bSet := make(map[string]bool, len(b))
	for _, id := range b {
		bSet[id] = true
	}
	for _, id := range a {
		if bSet[id] {
			return true
		}
	}
	return false
}

// findExistingPayable searches for an existing payable record by contractor, period start date, and payout item overlap.
// It queries all payables for the contractor/period and returns the first one with overlapping payout items.
// If no overlap is found, returns nil (indicating a new payable should be created).
func (s *ContractorPayablesService) findExistingPayable(ctx context.Context, contractorPageID, periodStart string, payoutItemIDs []string) (*ExistingPayable, error) {
	payablesDBID := s.cfg.Notion.Databases.ContractorPayables
	if payablesDBID == "" {
		return nil, errors.New("contractor payables database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: searching for existing payable contractor=%s periodStart=%s payoutItemCount=%d", contractorPageID, periodStart, len(payoutItemIDs)))

	// Build filter: Contractor relation contains contractorPageID AND Period start equals periodStart date
	filter := &nt.DatabaseQueryFilter{
		And: []nt.DatabaseQueryFilter{
			{
				Property: "Contractor",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Relation: &nt.RelationDatabaseQueryFilter{
						Contains: contractorPageID,
					},
				},
			},
			{
				Property: "Period",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Date: &nt.DatePropertyFilter{
						Equals: func() *time.Time {
							t, err := time.Parse("2006-01-02", periodStart)
							if err != nil {
								return nil
							}
							return &t
						}(),
					},
				},
			},
		},
	}

	// Query all payables for contractor + period (no PageSize limit to get all matches)
	resp, err := s.client.QueryDatabase(ctx, payablesDBID, &nt.DatabaseQuery{
		Filter: filter,
	})
	if err != nil {
		s.logger.Error(err, "[DEBUG] contractor_payables: failed to query existing payables")
		return nil, fmt.Errorf("failed to query existing payables: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: found %d payables for contractor=%s periodStart=%s", len(resp.Results), contractorPageID, periodStart))

	if len(resp.Results) == 0 {
		s.logger.Debug("[DEBUG] contractor_payables: no existing payable found")
		return nil, nil
	}

	// Check each payable for payout item overlap
	for _, page := range resp.Results {
		props, ok := page.Properties.(nt.DatabasePageProperties)
		if !ok {
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: failed to cast properties for pageID=%s", page.ID))
			continue
		}

		existingPayoutIDs := s.extractAllRelationIDs(props, "Payout Items")
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: checking payable pageID=%s existingPayoutItemCount=%d", page.ID, len(existingPayoutIDs)))

		// Check for overlap between input payout items and existing payable's payout items
		if hasOverlap(payoutItemIDs, existingPayoutIDs) {
			status := ""
			if statusProp, exists := props["Payment Status"]; exists && statusProp.Status != nil {
				status = statusProp.Status.Name
			}

			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: found matching payable with overlap pageID=%s status=%s", page.ID, status))

			return &ExistingPayable{
				PageID: page.ID,
				Status: status,
			}, nil
		}
	}

	s.logger.Debug("[DEBUG] contractor_payables: no payable found with overlapping payout items")
	return nil, nil
}

// PayableInfo represents payable information for Discord lookup
type PayableInfo struct {
	PageID      string
	Status      string     // "New", "Pending", "Paid"
	Total       float64
	Currency    string
	InvoiceID   string
	FileURL     string     // PDF file URL from Attachments field
	PaymentDate *time.Time // Only for "Paid" status
}

// FindPayableByContractorAndPeriodAllStatus finds all payables by contractor and period start date regardless of status
// Returns all payables (New, Pending, Paid, etc.)
// Returns empty slice if no payables found
func (s *ContractorPayablesService) FindPayableByContractorAndPeriodAllStatus(ctx context.Context, contractorPageID, periodStart string) ([]*PayableInfo, error) {
	payablesDBID := s.cfg.Notion.Databases.ContractorPayables
	if payablesDBID == "" {
		return nil, errors.New("contractor payables database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("looking up all payables (any status): contractor=%s periodStart=%s", contractorPageID, periodStart))

	// Build filter: Contractor + Period start date (NO status filter)
	filter := &nt.DatabaseQueryFilter{
		And: []nt.DatabaseQueryFilter{
			{
				Property: "Contractor",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Relation: &nt.RelationDatabaseQueryFilter{
						Contains: contractorPageID,
					},
				},
			},
			{
				Property: "Period",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Date: &nt.DatePropertyFilter{
						Equals: func() *time.Time {
							t, _ := time.Parse("2006-01-02", periodStart)
							return &t
						}(),
					},
				},
			},
		},
	}

	// Query ALL payables for this contractor + period
	resp, err := s.client.QueryDatabase(ctx, payablesDBID, &nt.DatabaseQuery{
		Filter: filter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query payables: %w", err)
	}

	if len(resp.Results) == 0 {
		s.logger.Debug("no payables found (any status)")
		return []*PayableInfo{}, nil
	}

	s.logger.Debug(fmt.Sprintf("found %d payable(s) for contractor=%s period=%s", len(resp.Results), contractorPageID, periodStart))

	// Extract all payables
	var payables []*PayableInfo
	for _, page := range resp.Results {
		props, ok := page.Properties.(nt.DatabasePageProperties)
		if !ok {
			s.logger.Debug(fmt.Sprintf("failed to cast properties for pageID=%s", page.ID))
			continue
		}

		payableInfo := &PayableInfo{
			PageID:    page.ID,
			Status:    s.extractStatus(props, "Payment Status"),
			Total:     s.extractNumber(props, "Total"),
			Currency:  s.extractSelect(props, "Currency"),
			InvoiceID: s.extractRichText(props, "Invoice ID"),
		}

		// Extract payment date if paid
		if dateProp, ok := props["Payment Date"]; ok && dateProp.Date != nil {
			payableInfo.PaymentDate = &dateProp.Date.Start.Time
		}

		// Extract PDF file URL from Attachments field
		if filesProp, ok := props["Attachments"]; ok && len(filesProp.Files) > 0 {
			// Get the first file (should be the PDF)
			if len(filesProp.Files[0].File.URL) > 0 {
				payableInfo.FileURL = filesProp.Files[0].File.URL
			} else if len(filesProp.Files[0].External.URL) > 0 {
				payableInfo.FileURL = filesProp.Files[0].External.URL
			}
		}

		payables = append(payables, payableInfo)
	}

	s.logger.Debug(fmt.Sprintf("returning %d payable(s) (any status)", len(payables)))
	return payables, nil
}

// FindPayableByContractorAndPeriod finds all PENDING payables by contractor and period start date
// Returns all pending payables (multiple invoices are supported)
// Returns empty slice if no pending payables found
func (s *ContractorPayablesService) FindPayableByContractorAndPeriod(ctx context.Context, contractorPageID, periodStart string) ([]*PayableInfo, error) {
	payablesDBID := s.cfg.Notion.Databases.ContractorPayables
	if payablesDBID == "" {
		return nil, errors.New("contractor payables database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("looking up payable: contractor=%s periodStart=%s", contractorPageID, periodStart))

	// Build filter: Contractor + Period start date + Status = "Pending"
	filter := &nt.DatabaseQueryFilter{
		And: []nt.DatabaseQueryFilter{
			{
				Property: "Contractor",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Relation: &nt.RelationDatabaseQueryFilter{
						Contains: contractorPageID,
					},
				},
			},
			{
				Property: "Period",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Date: &nt.DatePropertyFilter{
						Equals: func() *time.Time {
							t, _ := time.Parse("2006-01-02", periodStart)
							return &t
						}(),
					},
				},
			},
			{
				Property: "Payment Status",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Status: &nt.StatusDatabaseQueryFilter{
						Equals: "Pending",
					},
				},
			},
		},
	}

	// Query ALL pending payables for this contractor + period
	resp, err := s.client.QueryDatabase(ctx, payablesDBID, &nt.DatabaseQuery{
		Filter: filter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query payables: %w", err)
	}

	if len(resp.Results) == 0 {
		s.logger.Debug("no pending payable found")
		return []*PayableInfo{}, nil
	}

	// If multiple payables found, log info (this is expected - contractor can have multiple invoices)
	s.logger.Debug(fmt.Sprintf("found %d pending payable(s) for contractor=%s period=%s", len(resp.Results), contractorPageID, periodStart))

	// Extract all pending payables
	var payables []*PayableInfo
	for _, page := range resp.Results {
		props, ok := page.Properties.(nt.DatabasePageProperties)
		if !ok {
			s.logger.Debug(fmt.Sprintf("failed to cast properties for pageID=%s", page.ID))
			continue
		}

		payableInfo := &PayableInfo{
			PageID:    page.ID,
			Status:    s.extractSelect(props, "Payment Status"),
			Total:     s.extractNumber(props, "Total"),
			Currency:  s.extractSelect(props, "Currency"),
			InvoiceID: s.extractRichText(props, "Invoice ID"),
		}

		// Extract payment date if paid
		if dateProp, ok := props["Payment Date"]; ok && dateProp.Date != nil {
			payableInfo.PaymentDate = &dateProp.Date.Start.Time
		}

		// Extract PDF file URL from Attachments field
		if filesProp, ok := props["Attachments"]; ok && len(filesProp.Files) > 0 {
			// Get the first file (should be the PDF)
			if len(filesProp.Files[0].File.URL) > 0 {
				payableInfo.FileURL = filesProp.Files[0].File.URL
			} else if len(filesProp.Files[0].External.URL) > 0 {
				payableInfo.FileURL = filesProp.Files[0].External.URL
			}
		}

		payables = append(payables, payableInfo)
	}

	s.logger.Debug(fmt.Sprintf("returning %d pending payable(s)", len(payables)))
	return payables, nil
}

// CreatePayable creates a new payable record or updates existing one in the Contractor Payables database
// - If existing record with status "New" found: updates it
// - If existing record with status "Pending" or other: skips (returns existing page ID)
// - If no existing record: creates new
// Returns the page ID (created or existing)
func (s *ContractorPayablesService) CreatePayable(ctx context.Context, input CreatePayableInput) (string, error) {
	payablesDBID := s.cfg.Notion.Databases.ContractorPayables
	if payablesDBID == "" {
		return "", errors.New("contractor payables database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: creating payable contractor=%s total=%.2f currency=%s invoiceID=%s payoutItems=%d",
		input.ContractorPageID, input.Total, input.Currency, input.InvoiceID, len(input.PayoutItemIDs)))

	// Check for existing payable with overlapping payout items
	existing, err := s.findExistingPayable(ctx, input.ContractorPageID, input.PeriodStart, input.PayoutItemIDs)
	if err != nil {
		s.logger.Error(err, "[DEBUG] contractor_payables: failed to check existing payable - proceeding with create")
		// Continue with creation on error
	} else if existing != nil {
		if existing.Status == "New" {
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: found existing payable with status=New, updating pageID=%s", existing.PageID))
			return s.updatePayable(ctx, existing.PageID, input)
		}
		// Status is Pending or other - skip
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: existing payable has status=%s, skipping update", existing.Status))
		return existing.PageID, nil
	}

	// Build properties for the new payable
	props := nt.DatabasePageProperties{
		// Title: Payable name (empty, Auto Name formula will fill it)
		"Payable Title": nt.DatabasePageProperty{
			Title: []nt.RichText{
				{Text: &nt.Text{Content: ""}},
			},
		},
		// Total
		"Total": nt.DatabasePageProperty{
			Number: &input.Total,
		},
		// Payment Status: New
		"Payment Status": nt.DatabasePageProperty{
			Status: &nt.SelectOptions{
				Name: "New",
			},
		},
		// Contractor relation
		"Contractor": nt.DatabasePageProperty{
			Relation: []nt.Relation{
				{ID: input.ContractorPageID},
			},
		},
	}

	// Add Currency
	if input.Currency != "" {
		props["Currency"] = nt.DatabasePageProperty{
			Select: &nt.SelectOptions{
				Name: input.Currency,
			},
		}
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: set currency=%s", input.Currency))
	}

	// Add Period date range (start and end)
	if input.PeriodStart != "" && input.PeriodEnd != "" {
		startDateObj, startErr := nt.ParseDateTime(input.PeriodStart)
		endDateObj, endErr := nt.ParseDateTime(input.PeriodEnd)
		if startErr == nil && endErr == nil {
			props["Period"] = nt.DatabasePageProperty{
				Date: &nt.Date{
					Start: startDateObj,
					End:   &endDateObj,
				},
			}
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: set period=%s to %s", input.PeriodStart, input.PeriodEnd))
		} else {
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: failed to parse period start=%s end=%s: startErr=%v endErr=%v", input.PeriodStart, input.PeriodEnd, startErr, endErr))
		}
	}

	// Add Invoice Date
	if input.InvoiceDate != "" {
		dateObj, err := nt.ParseDateTime(input.InvoiceDate)
		if err == nil {
			props["Invoice Date"] = nt.DatabasePageProperty{
				Date: &nt.Date{
					Start: dateObj,
				},
			}
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: set invoiceDate=%s", input.InvoiceDate))
		} else {
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: failed to parse invoiceDate=%s: %v", input.InvoiceDate, err))
		}
	}

	// Add Invoice ID
	if input.InvoiceID != "" {
		props["Invoice ID"] = nt.DatabasePageProperty{
			RichText: []nt.RichText{
				{Text: &nt.Text{Content: input.InvoiceID}},
			},
		}
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: set invoiceID=%s", input.InvoiceID))
	}

	// Add Payout Items relation
	if len(input.PayoutItemIDs) > 0 {
		relations := make([]nt.Relation, len(input.PayoutItemIDs))
		for i, id := range input.PayoutItemIDs {
			relations[i] = nt.Relation{ID: id}
		}
		props["Payout Items"] = nt.DatabasePageProperty{
			Relation: relations,
		}
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: set payoutItems count=%d", len(input.PayoutItemIDs)))
	}

	// Add Contractor Type (default to "Individual" if not provided)
	contractorType := input.ContractorType
	if contractorType == "" {
		contractorType = "Individual"
	}
	props["Contractor Type"] = nt.DatabasePageProperty{
		Select: &nt.SelectOptions{
			Name: contractorType,
		},
	}
	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: set contractorType=%s", contractorType))

	// Add Exchange Rate (optional)
	if input.ExchangeRate > 0 {
		props["Exchange Rate"] = nt.DatabasePageProperty{
			Number: &input.ExchangeRate,
		}
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: set exchangeRate=%.2f", input.ExchangeRate))
	}

	params := nt.CreatePageParams{
		ParentType:             nt.ParentTypeDatabase,
		ParentID:               payablesDBID,
		DatabasePageProperties: &props,
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: creating page in database=%s", payablesDBID))

	page, err := s.client.CreatePage(ctx, params)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_payables: failed to create payable: %v", err))
		return "", fmt.Errorf("failed to create payable: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: created payable pageID=%s", page.ID))

	// Upload PDF attachment if provided
	if len(input.PDFBytes) > 0 && s.notionSvc != nil {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: uploading PDF attachment size=%d bytes", len(input.PDFBytes)))

		filename := input.InvoiceID + ".pdf"
		fileUploadID, uploadErr := s.notionSvc.UploadFile(filename, "application/pdf", input.PDFBytes)
		if uploadErr != nil {
			s.logger.Error(uploadErr, "[DEBUG] contractor_payables: failed to upload PDF - continuing without attachment")
			// Non-fatal: page is created, just without attachment
		} else {
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: PDF uploaded fileUploadID=%s", fileUploadID))

			// Attach the uploaded file to the page
			attachErr := s.notionSvc.UpdatePagePropertiesWithFileUpload(page.ID, "Attachments", fileUploadID, filename)
			if attachErr != nil {
				s.logger.Error(attachErr, "[DEBUG] contractor_payables: failed to attach PDF to page - continuing without attachment")
				// Non-fatal: page is created, just without attachment
			} else {
				s.logger.Debug("[DEBUG] contractor_payables: PDF attached successfully")
			}
		}
	}

	return page.ID, nil
}

// updatePayable updates an existing payable record with new data
func (s *ContractorPayablesService) updatePayable(ctx context.Context, pageID string, input CreatePayableInput) (string, error) {
	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: updating payable pageID=%s", pageID))

	// Build properties for update
	props := nt.DatabasePageProperties{
		// Total
		"Total": nt.DatabasePageProperty{
			Number: &input.Total,
		},
	}

	// Add Currency
	if input.Currency != "" {
		props["Currency"] = nt.DatabasePageProperty{
			Select: &nt.SelectOptions{
				Name: input.Currency,
			},
		}
	}

	// Add Period date range (start and end)
	if input.PeriodStart != "" && input.PeriodEnd != "" {
		startDateObj, startErr := nt.ParseDateTime(input.PeriodStart)
		endDateObj, endErr := nt.ParseDateTime(input.PeriodEnd)
		if startErr == nil && endErr == nil {
			props["Period"] = nt.DatabasePageProperty{
				Date: &nt.Date{
					Start: startDateObj,
					End:   &endDateObj,
				},
			}
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: updating period=%s to %s", input.PeriodStart, input.PeriodEnd))
		}
	}

	// Add Invoice Date
	if input.InvoiceDate != "" {
		dateObj, err := nt.ParseDateTime(input.InvoiceDate)
		if err == nil {
			props["Invoice Date"] = nt.DatabasePageProperty{
				Date: &nt.Date{
					Start: dateObj,
				},
			}
		}
	}

	// Add Invoice ID
	if input.InvoiceID != "" {
		props["Invoice ID"] = nt.DatabasePageProperty{
			RichText: []nt.RichText{
				{Text: &nt.Text{Content: input.InvoiceID}},
			},
		}
	}

	// Add Payout Items relation
	if len(input.PayoutItemIDs) > 0 {
		relations := make([]nt.Relation, len(input.PayoutItemIDs))
		for i, id := range input.PayoutItemIDs {
			relations[i] = nt.Relation{ID: id}
		}
		props["Payout Items"] = nt.DatabasePageProperty{
			Relation: relations,
		}
	}

	// Add Contractor Type (default to "Individual" if not provided)
	contractorType := input.ContractorType
	if contractorType == "" {
		contractorType = "Individual"
	}
	props["Contractor Type"] = nt.DatabasePageProperty{
		Select: &nt.SelectOptions{
			Name: contractorType,
		},
	}

	// Add Exchange Rate (optional)
	if input.ExchangeRate > 0 {
		props["Exchange Rate"] = nt.DatabasePageProperty{
			Number: &input.ExchangeRate,
		}
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: updating exchangeRate=%.2f", input.ExchangeRate))
	}

	// Update the page
	_, err := s.client.UpdatePage(ctx, pageID, nt.UpdatePageParams{
		DatabasePageProperties: props,
	})
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_payables: failed to update payable: %v", err))
		return "", fmt.Errorf("failed to update payable: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: updated payable pageID=%s", pageID))

	// Upload PDF attachment if provided
	if len(input.PDFBytes) > 0 && s.notionSvc != nil {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: uploading PDF attachment size=%d bytes", len(input.PDFBytes)))

		filename := input.InvoiceID + ".pdf"
		fileUploadID, uploadErr := s.notionSvc.UploadFile(filename, "application/pdf", input.PDFBytes)
		if uploadErr != nil {
			s.logger.Error(uploadErr, "[DEBUG] contractor_payables: failed to upload PDF - continuing without attachment")
		} else {
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: PDF uploaded fileUploadID=%s", fileUploadID))

			attachErr := s.notionSvc.UpdatePagePropertiesWithFileUpload(pageID, "Attachments", fileUploadID, filename)
			if attachErr != nil {
				s.logger.Error(attachErr, "[DEBUG] contractor_payables: failed to attach PDF to page - continuing without attachment")
			} else {
				s.logger.Debug("[DEBUG] contractor_payables: PDF attached successfully")
			}
		}
	}

	return pageID, nil
}

// PendingPayable represents a payable record with Pending status
type PendingPayable struct {
	PageID            string   // Payable page ID
	ContractorPageID  string   // From Contractor relation
	ContractorName    string   // Rollup from Contractor (Full Name)
	Discord           string   // Discord username from Contractor
	Total             float64  // Total amount
	Currency          string   // USD or VND
	Period            string   // YYYY-MM-DD
	PayoutItemPageIDs []string // From Payout Items relation (multiple)
}

// QueryPendingPayablesByPeriod queries all contractor payables with Payment Status="Pending" for a given month and batch.
// Month should be in YYYY-MM format (e.g., "2025-01").
// Batch is the PayDay (1 or 15) which determines the period start date:
//   - batch 1: Period starts on 1st (YYYY-MM-01)
//   - batch 15: Period starts on 15th (YYYY-MM-15)
// Returns empty slice if no results found (not an error).
func (s *ContractorPayablesService) QueryPendingPayablesByPeriod(ctx context.Context, month string, batch int) ([]PendingPayable, error) {
	payablesDBID := s.cfg.Notion.Databases.ContractorPayables
	if payablesDBID == "" {
		return nil, errors.New("contractor payables database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: querying pending payables month=%s batch=%d", month, batch))

	// Calculate period start date based on batch
	// batch 1 → 1st of month, batch 15 → 15th of month
	day := "01"
	if batch == 15 {
		day = "15"
	}
	periodStr := fmt.Sprintf("%s-%s", month, day)

	periodStart, err := time.Parse("2006-01-02", periodStr)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_payables: failed to parse period=%s", periodStr))
		return nil, fmt.Errorf("invalid period format: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: using Period start date filter=%s", periodStart.Format("2006-01-02")))

	// Build filter: Payment Status = Pending AND Period start date equals the calculated date
	filter := &nt.DatabaseQueryFilter{
		And: []nt.DatabaseQueryFilter{
			{
				Property: "Payment Status",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Status: &nt.StatusDatabaseQueryFilter{
						Equals: "Pending",
					},
				},
			},
			{
				Property: "Period",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Date: &nt.DatePropertyFilter{
						Equals: &periodStart,
					},
				},
			},
		},
	}

	query := &nt.DatabaseQuery{
		Filter:   filter,
		PageSize: 100,
	}

	var payables []PendingPayable

	// Query with pagination
	for {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: executing query on database=%s", payablesDBID))

		resp, err := s.client.QueryDatabase(ctx, payablesDBID, query)
		if err != nil {
			s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_payables: failed to query database: %v", err))
			return nil, fmt.Errorf("failed to query contractor payables database: %w", err)
		}

		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: found %d payable entries in this page", len(resp.Results)))

		for _, page := range resp.Results {
			props, ok := page.Properties.(nt.DatabasePageProperties)
			if !ok {
				s.logger.Debug("[DEBUG] contractor_payables: failed to cast page properties")
				continue
			}

			// Extract payable data
			payable := PendingPayable{
				PageID:            page.ID,
				ContractorPageID:  s.extractFirstRelationID(props, "Contractor"),
				Total:             s.extractNumber(props, "Total"),
				Currency:          s.extractSelect(props, "Currency"),
				Period:            periodStr,
				PayoutItemPageIDs: s.extractAllRelationIDs(props, "Payout Items"),
			}

			// Extract contractor name from rollup if available
			payable.ContractorName = s.extractRollupTitle(props, "Contractor Name")

			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: parsed payable pageID=%s contractor=%s contractorName=%s total=%.2f currency=%s payoutItems=%d",
				payable.PageID, payable.ContractorPageID, payable.ContractorName, payable.Total, payable.Currency, len(payable.PayoutItemPageIDs)))

			payables = append(payables, payable)
		}

		if !resp.HasMore || resp.NextCursor == nil {
			break
		}

		query.StartCursor = *resp.NextCursor
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: fetching next page with cursor=%s", *resp.NextCursor))
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: total pending payables found=%d", len(payables)))

	// Fetch contractor info in parallel for payables missing contractor name or discord
	s.logger.Debug("[DEBUG] contractor_payables: starting parallel contractor info fetch")
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := range payables {
		if payables[i].ContractorPageID != "" && (payables[i].ContractorName == "" || payables[i].Discord == "") {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				info := s.getContractorInfo(ctx, payables[idx].ContractorPageID)
				mu.Lock()
				if payables[idx].ContractorName == "" && info.Name != "" {
					payables[idx].ContractorName = info.Name
				}
				if payables[idx].Discord == "" && info.Discord != "" {
					payables[idx].Discord = info.Discord
				}
				mu.Unlock()
				s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: fetched contractor info idx=%d name=%s discord=%s", idx, info.Name, info.Discord))
			}(i)
		}
	}

	wg.Wait()
	s.logger.Debug("[DEBUG] contractor_payables: parallel contractor info fetch completed")

	return payables, nil
}

// UpdatePayableStatus updates a payable's Payment Status and Payment Date.
// pageID: Payable page ID to update
// status: New status value (e.g., "Paid")
// paymentDate: Payment date in YYYY-MM-DD format
func (s *ContractorPayablesService) UpdatePayableStatus(ctx context.Context, pageID string, status string, paymentDate string) error {
	if pageID == "" {
		return errors.New("page ID is required")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: updating payable status pageID=%s status=%s paymentDate=%s", pageID, status, paymentDate))

	// Build update parameters
	props := nt.DatabasePageProperties{
		"Payment Status": nt.DatabasePageProperty{
			Status: &nt.SelectOptions{
				Name: status,
			},
		},
	}

	// Add Payment Date if provided
	if paymentDate != "" {
		dateObj, err := nt.ParseDateTime(paymentDate)
		if err == nil {
			props["Payment Date"] = nt.DatabasePageProperty{
				Date: &nt.Date{
					Start: dateObj,
				},
			}
		} else {
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: failed to parse paymentDate=%s: %v", paymentDate, err))
		}
	}

	params := nt.UpdatePageParams{
		DatabasePageProperties: props,
	}

	_, err := s.client.UpdatePage(ctx, pageID, params)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_payables: failed to update payable status: %v", err))
		return fmt.Errorf("failed to update payable status: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: updated payable pageID=%s status=%s successfully", pageID, status))
	return nil
}

// GetContractorPayDay gets the PayDay value from a contractor's Service Rate (ContractorRates).
// Returns the PayDay value (1 or 15) or error if not found.
func (s *ContractorPayablesService) GetContractorPayDay(ctx context.Context, contractorPageID string) (int, error) {
	if contractorPageID == "" {
		return 0, errors.New("contractor page ID is required")
	}

	contractorRatesDBID := s.cfg.Notion.Databases.ContractorRates
	if contractorRatesDBID == "" {
		return 0, errors.New("contractor rates database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: fetching PayDay for contractor=%s", contractorPageID))

	// Query Service Rate database for records with Contractor relation and Active status
	filter := &nt.DatabaseQueryFilter{
		And: []nt.DatabaseQueryFilter{
			{
				Property: "Contractor",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Relation: &nt.RelationDatabaseQueryFilter{
						Contains: contractorPageID,
					},
				},
			},
			{
				Property: "Status",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Status: &nt.StatusDatabaseQueryFilter{
						Equals: "Active",
					},
				},
			},
		},
	}

	query := &nt.DatabaseQuery{
		Filter:   filter,
		PageSize: 1,
	}

	resp, err := s.client.QueryDatabase(ctx, contractorRatesDBID, query)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_payables: failed to query contractor rates for contractor=%s", contractorPageID))
		return 0, fmt.Errorf("failed to query contractor rates: %w", err)
	}

	if len(resp.Results) == 0 {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: no active service rate found for contractor=%s", contractorPageID))
		return 0, fmt.Errorf("no active service rate found for contractor %s", contractorPageID)
	}

	page := resp.Results[0]
	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		return 0, errors.New("failed to cast service rate page properties")
	}

	// Extract PayDay from Select property (property name is "Payday" with lowercase 'd')
	payDayStr := s.extractSelect(props, "Payday")
	if payDayStr == "" {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: PayDay property not found or empty for contractor=%s", contractorPageID))
		return 0, fmt.Errorf("PayDay property not found for contractor %s", contractorPageID)
	}

	// Parse PayDay value ("01" or "15" as string)
	var payDay int
	_, err = fmt.Sscanf(payDayStr, "%d", &payDay)
	if err != nil {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: invalid PayDay value=%s for contractor=%s", payDayStr, contractorPageID))
		return 0, fmt.Errorf("invalid PayDay value: %s", payDayStr)
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: contractor=%s PayDay=%d", contractorPageID, payDay))
	return payDay, nil
}

// FindPayableByInvoiceID finds a payable by Invoice ID.
// Returns the payable if found with status "New", or nil if not found.
func (s *ContractorPayablesService) FindPayableByInvoiceID(ctx context.Context, invoiceID string) (*ExistingPayable, error) {
	payablesDBID := s.cfg.Notion.Databases.ContractorPayables
	if payablesDBID == "" {
		return nil, errors.New("contractor payables database ID not configured")
	}

	if invoiceID == "" {
		return nil, errors.New("invoice ID is required")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: searching for payable by invoiceID=%s", invoiceID))

	// Build filter: Invoice ID equals invoiceID AND Payment Status = "New"
	filter := &nt.DatabaseQueryFilter{
		And: []nt.DatabaseQueryFilter{
			{
				Property: "Invoice ID",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					RichText: &nt.TextPropertyFilter{
						Equals: invoiceID,
					},
				},
			},
			{
				Property: "Payment Status",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Status: &nt.StatusDatabaseQueryFilter{
						Equals: "New",
					},
				},
			},
		},
	}

	resp, err := s.client.QueryDatabase(ctx, payablesDBID, &nt.DatabaseQuery{
		Filter:   filter,
		PageSize: 1,
	})
	if err != nil {
		s.logger.Error(err, "[DEBUG] contractor_payables: failed to query payables by invoice ID")
		return nil, fmt.Errorf("failed to query payables by invoice ID: %w", err)
	}

	if len(resp.Results) == 0 {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: no payable found with invoiceID=%s and status=New", invoiceID))
		return nil, nil
	}

	page := resp.Results[0]
	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: found payable by invoiceID=%s pageID=%s", invoiceID, page.ID))

	// Extract status
	status := ""
	if props, ok := page.Properties.(nt.DatabasePageProperties); ok {
		if statusProp, exists := props["Payment Status"]; exists && statusProp.Status != nil {
			status = statusProp.Status.Name
		}
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: payable invoiceID=%s status=%s", invoiceID, status))

	return &ExistingPayable{
		PageID: page.ID,
		Status: status,
	}, nil
}

// FindPayableByInvoiceIDAnyStatus finds a payable by Invoice ID regardless of status.
// Returns the payable if found, or nil if not found.
func (s *ContractorPayablesService) FindPayableByInvoiceIDAnyStatus(ctx context.Context, invoiceID string) (*ExistingPayable, error) {
	payablesDBID := s.cfg.Notion.Databases.ContractorPayables
	if payablesDBID == "" {
		return nil, errors.New("contractor payables database ID not configured")
	}

	if invoiceID == "" {
		return nil, errors.New("invoice ID is required")
	}

	filter := &nt.DatabaseQueryFilter{
		Property: "Invoice ID",
		DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
			RichText: &nt.TextPropertyFilter{
				Equals: invoiceID,
			},
		},
	}

	resp, err := s.client.QueryDatabase(ctx, payablesDBID, &nt.DatabaseQuery{
		Filter:   filter,
		PageSize: 1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query payables by invoice ID: %w", err)
	}

	if len(resp.Results) == 0 {
		return nil, nil
	}

	page := resp.Results[0]
	status := ""
	if props, ok := page.Properties.(nt.DatabasePageProperties); ok {
		if statusProp, exists := props["Payment Status"]; exists && statusProp.Status != nil {
			status = statusProp.Status.Name
		}
	}

	return &ExistingPayable{
		PageID: page.ID,
		Status: status,
	}, nil
}

// UpdatePayableStatusByInvoiceID updates a payable's Payment Status by Invoice ID.
// invoiceID: Invoice ID to find the payable
// status: New status value (e.g., "Pending")
// Returns error if payable not found or update fails.
func (s *ContractorPayablesService) UpdatePayableStatusByInvoiceID(ctx context.Context, invoiceID string, status string) error {
	if invoiceID == "" {
		return errors.New("invoice ID is required")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: updating payable by invoiceID=%s to status=%s", invoiceID, status))

	// First, find the payable by invoice ID
	payable, err := s.FindPayableByInvoiceID(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to find payable: %w", err)
	}

	if payable == nil {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: no payable found with invoiceID=%s to update", invoiceID))
		return fmt.Errorf("payable with invoice ID %s not found", invoiceID)
	}

	// Update the status (without payment date, as this is just status change)
	return s.UpdatePayableStatus(ctx, payable.PageID, status, "")
}

// NewPayable represents a payable with status "New" fetched from Notion
type NewPayable struct {
	PageID           string
	InvoiceID        string
	ContractorPageID string
}

// QueryNewPayables queries all contractor payables with Payment Status="New" and a non-empty Invoice ID.
// Returns paginated results of payables ready for email matching.
func (s *ContractorPayablesService) QueryNewPayables(ctx context.Context) ([]NewPayable, error) {
	payablesDBID := s.cfg.Notion.Databases.ContractorPayables
	if payablesDBID == "" {
		return nil, errors.New("contractor payables database ID not configured")
	}

	s.logger.Debug("[DEBUG] contractor_payables: querying new payables for email matching")

	// Build filter: Payment Status = "New" AND Invoice ID is not empty
	filter := &nt.DatabaseQueryFilter{
		And: []nt.DatabaseQueryFilter{
			{
				Property: "Payment Status",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Status: &nt.StatusDatabaseQueryFilter{
						Equals: "New",
					},
				},
			},
			{
				Property: "Invoice ID",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					RichText: &nt.TextPropertyFilter{
						IsNotEmpty: true,
					},
				},
			},
		},
	}

	query := &nt.DatabaseQuery{
		Filter:   filter,
		PageSize: 100,
	}

	var payables []NewPayable

	// Query with pagination
	for {
		resp, err := s.client.QueryDatabase(ctx, payablesDBID, query)
		if err != nil {
			s.logger.Error(err, "[DEBUG] contractor_payables: failed to query new payables")
			return nil, fmt.Errorf("failed to query new payables: %w", err)
		}

		for _, page := range resp.Results {
			props, ok := page.Properties.(nt.DatabasePageProperties)
			if !ok {
				s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: failed to cast properties for pageID=%s", page.ID))
				continue
			}

			invoiceID := s.extractRichText(props, "Invoice ID")
			if invoiceID == "" {
				continue
			}

			payable := NewPayable{
				PageID:           page.ID,
				InvoiceID:        invoiceID,
				ContractorPageID: s.extractFirstRelationID(props, "Contractor"),
			}

			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: found new payable pageID=%s invoiceID=%s contractor=%s",
				payable.PageID, payable.InvoiceID, payable.ContractorPageID))

			payables = append(payables, payable)
		}

		if !resp.HasMore || resp.NextCursor == nil {
			break
		}

		query.StartCursor = *resp.NextCursor
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: total new payables found=%d", len(payables)))
	return payables, nil
}

// Helper functions for property extraction

func (s *ContractorPayablesService) extractFirstRelationID(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || len(prop.Relation) == 0 {
		return ""
	}
	return prop.Relation[0].ID
}

func (s *ContractorPayablesService) extractAllRelationIDs(props nt.DatabasePageProperties, propName string) []string {
	prop, ok := props[propName]
	if !ok || len(prop.Relation) == 0 {
		return nil
	}
	ids := make([]string, len(prop.Relation))
	for i, rel := range prop.Relation {
		ids[i] = rel.ID
	}
	return ids
}

func (s *ContractorPayablesService) extractNumber(props nt.DatabasePageProperties, propName string) float64 {
	prop, ok := props[propName]
	if !ok || prop.Number == nil {
		return 0
	}
	return *prop.Number
}

func (s *ContractorPayablesService) extractSelect(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Select == nil {
		return ""
	}
	return prop.Select.Name
}

func (s *ContractorPayablesService) extractStatus(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Status == nil {
		return ""
	}
	return prop.Status.Name
}

func (s *ContractorPayablesService) extractRichText(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || len(prop.RichText) == 0 {
		return ""
	}
	return prop.RichText[0].PlainText
}

func (s *ContractorPayablesService) extractRollupTitle(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Rollup == nil {
		return ""
	}

	// Handle array type rollup (most common for relation rollups)
	if prop.Rollup.Type == "array" && len(prop.Rollup.Array) > 0 {
		firstItem := prop.Rollup.Array[0]
		if len(firstItem.Title) > 0 {
			var result string
			for _, rt := range firstItem.Title {
				result += rt.PlainText
			}
			return result
		}
	}

	return ""
}

// contractorInfo holds contractor details fetched from Contractor page
type contractorInfo struct {
	Name    string
	Discord string
}

// getContractorInfo fetches the Full Name and Discord from a Contractor page
func (s *ContractorPayablesService) getContractorInfo(ctx context.Context, pageID string) contractorInfo {
	info := contractorInfo{}

	page, err := s.client.FindPageByID(ctx, pageID)
	if err != nil {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: getContractorInfo failed to fetch contractor page %s: %v", pageID, err))
		return info
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: getContractorInfo failed to cast page properties for %s", pageID))
		return info
	}

	// Try to get Full Name from Title property
	if prop, ok := props["Full Name"]; ok && len(prop.Title) > 0 {
		info.Name = prop.Title[0].PlainText
	} else if prop, ok := props["Name"]; ok && len(prop.Title) > 0 {
		info.Name = prop.Title[0].PlainText
	}

	// Get Discord username from rich_text property
	if prop, ok := props["Discord"]; ok && len(prop.RichText) > 0 {
		info.Discord = prop.RichText[0].PlainText
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: getContractorInfo pageID=%s name=%s discord=%s", pageID, info.Name, info.Discord))
	return info
}
