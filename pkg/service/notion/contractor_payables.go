package notion

import (
	"context"
	"errors"
	"fmt"
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
	Period           string   // YYYY-MM-DD start of month (required)
	InvoiceDate      string   // YYYY-MM-DD (required)
	InvoiceID        string   // Invoice number e.g., CONTR-202512-A1B2 (required)
	PayoutItemIDs    []string // Relation to Payout Items (required)
	ContractorType   string   // "Individual", "Sole Proprietor", "LLC", etc. (optional, defaults to "Individual")
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

// findExistingPayable searches for an existing payable record by contractor and period
func (s *ContractorPayablesService) findExistingPayable(ctx context.Context, contractorPageID, period string) (*ExistingPayable, error) {
	payablesDBID := s.cfg.Notion.Databases.ContractorPayables
	if payablesDBID == "" {
		return nil, errors.New("contractor payables database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: searching for existing payable contractor=%s period=%s", contractorPageID, period))

	// Build filter: Contractor relation contains contractorPageID AND Period equals period date
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
							t, err := time.Parse("2006-01-02", period)
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

	resp, err := s.client.QueryDatabase(ctx, payablesDBID, &nt.DatabaseQuery{
		Filter:   filter,
		PageSize: 1,
	})
	if err != nil {
		s.logger.Error(err, "[DEBUG] contractor_payables: failed to query existing payables")
		return nil, fmt.Errorf("failed to query existing payables: %w", err)
	}

	if len(resp.Results) == 0 {
		s.logger.Debug("[DEBUG] contractor_payables: no existing payable found")
		return nil, nil
	}

	page := resp.Results[0]
	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: found existing payable pageID=%s", page.ID))

	// Extract status
	status := ""
	if props, ok := page.Properties.(nt.DatabasePageProperties); ok {
		if statusProp, exists := props["Payment Status"]; exists && statusProp.Status != nil {
			status = statusProp.Status.Name
		}
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: existing payable status=%s", status))

	return &ExistingPayable{
		PageID: page.ID,
		Status: status,
	}, nil
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

	// Check for existing payable
	existing, err := s.findExistingPayable(ctx, input.ContractorPageID, input.Period)
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
		"Payable": nt.DatabasePageProperty{
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

	// Add Period date
	if input.Period != "" {
		dateObj, err := nt.ParseDateTime(input.Period)
		if err == nil {
			props["Period"] = nt.DatabasePageProperty{
				Date: &nt.Date{
					Start: dateObj,
				},
			}
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: set period=%s", input.Period))
		} else {
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: failed to parse period=%s: %v", input.Period, err))
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
