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

// InvoiceSplitsDBID is the Notion database ID for Invoice Splits
const InvoiceSplitsDBID = "2c364b29b84c80498a8df7befd22f7fc"

// InvoiceSplitService handles invoice split operations with Notion
type InvoiceSplitService struct {
	*baseService
}

// InvoiceSplitData represents invoice split data from Notion
type InvoiceSplitData struct {
	PageID      string
	Amount      float64
	Role        string // Sales, Account Manager, etc.
	Currency    string
	Description string // From Description formula
}

// PendingCommissionSplit represents a pending commission split for payout processing
type PendingCommissionSplit struct {
	PageID       string
	Name         string
	AutoName     string // Auto Name formula field containing formatted name with ID suffix (e.g., "SPL :: 202511 :: ... :: HJR4Z")
	Amount       float64
	Currency     string
	Role         string
	Type         string // Commission, Bonus, Fee
	PersonPageID string // From Person relation
	Month        string // Date in YYYY-MM-DD format from Month property
	Description  string // Formula field in Notion (read-only, automatically calculated)
}

// CreateCommissionSplitInput contains the data needed to create a commission split
type CreateCommissionSplitInput struct {
	Name              string
	Amount            float64
	Currency          string
	Month             time.Time
	Role              string // Sales, Account Manager, Delivery Lead, Hiring Referral
	Type              string // Commission
	Status            string // Pending
	ContractorPageID  string
	DeploymentPageID  string
	InvoiceItemPageID string
	InvoicePageID     string
	Description       string // Not used (Description is a formula field in Notion)
}

// NewInvoiceSplitService creates a new Notion invoice split service
func NewInvoiceSplitService(cfg *config.Config, l logger.Logger) *InvoiceSplitService {
	base := newBaseService(cfg, l)
	if base == nil {
		return nil
	}

	l.Debug("creating new InvoiceSplitService")

	return &InvoiceSplitService{baseService: base}
}

// GetInvoiceSplitByID fetches invoice split data by page ID
func (s *InvoiceSplitService) GetInvoiceSplitByID(ctx context.Context, splitPageID string) (*InvoiceSplitData, error) {
	s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: fetching page=%s", splitPageID))

	page, err := s.client.FindPageByID(ctx, splitPageID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] invoice_split: failed to fetch page=%s: %v", splitPageID, err))
		return nil, fmt.Errorf("failed to fetch invoice split page: %w", err)
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Debug("[DEBUG] invoice_split: failed to cast page properties")
		return nil, errors.New("failed to cast invoice split page properties")
	}

	// Debug: Log available properties
	s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: Available properties for page %s:", splitPageID))
	for propName := range props {
		s.logger.Debug(fmt.Sprintf("[DEBUG]   - %s", propName))
	}

	// Extract invoice split data
	data := &InvoiceSplitData{
		PageID:   splitPageID,
		Amount:   ExtractNumber(props, "Amount"),
		Role:     ExtractSelect(props, "Role"),
		Currency: ExtractSelect(props, "Currency"),
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: parsed data pageID=%s role=%s amount=%.2f currency=%s",
		data.PageID, data.Role, data.Amount, data.Currency))

	return data, nil
}

// CreateCommissionSplit creates a new invoice split record in Notion
func (s *InvoiceSplitService) CreateCommissionSplit(ctx context.Context, input CreateCommissionSplitInput) (*InvoiceSplitData, error) {
	l := s.logger.Fields(logger.Fields{
		"service": "invoice_split",
		"method":  "CreateCommissionSplit",
		"name":    input.Name,
		"role":    input.Role,
		"amount":  input.Amount,
	})

	l.Debug("creating commission split in Notion")

	// Build page properties
	props := nt.DatabasePageProperties{
		"Name": nt.DatabasePageProperty{
			Title: []nt.RichText{
				{
					Type: nt.RichTextTypeText,
					Text: &nt.Text{Content: input.Name},
				},
			},
		},
		"Amount": nt.DatabasePageProperty{
			Number: &input.Amount,
		},
		"Currency": nt.DatabasePageProperty{
			Select: &nt.SelectOptions{
				Name: input.Currency,
			},
		},
		"Month": nt.DatabasePageProperty{
			Date: &nt.Date{
				Start: nt.NewDateTime(input.Month, false),
			},
		},
		"Role": nt.DatabasePageProperty{
			Select: &nt.SelectOptions{
				Name: input.Role,
			},
		},
		"Type": nt.DatabasePageProperty{
			Select: &nt.SelectOptions{
				Name: input.Type,
			},
		},
		"Status": nt.DatabasePageProperty{
			Select: &nt.SelectOptions{
				Name: input.Status,
			},
		},
	}

	// Add relations if provided
	if input.ContractorPageID != "" {
		// Person is a relation to Contractors database
		props["Person"] = nt.DatabasePageProperty{
			Relation: []nt.Relation{
				{ID: input.ContractorPageID},
			},
		}
		l.Debug(fmt.Sprintf("setting Person relation to contractor: %s", input.ContractorPageID))
	}

	if input.DeploymentPageID != "" {
		props["Deployment"] = nt.DatabasePageProperty{
			Relation: []nt.Relation{
				{ID: input.DeploymentPageID},
			},
		}
	}

	if input.InvoiceItemPageID != "" {
		props["Invoice Item"] = nt.DatabasePageProperty{
			Relation: []nt.Relation{
				{ID: input.InvoiceItemPageID},
			},
		}
	}

	// Note: Description column is a formula field in Notion and cannot be set manually
	// It will be automatically calculated based on other properties

	// Create the page
	page, err := s.client.CreatePage(ctx, nt.CreatePageParams{
		ParentType:             nt.ParentTypeDatabase,
		ParentID:               InvoiceSplitsDBID,
		DatabasePageProperties: &props,
	})
	if err != nil {
		l.Error(err, "failed to create commission split")
		return nil, fmt.Errorf("failed to create commission split: %w", err)
	}

	l.Info("commission split created successfully")

	// Fetch the created page to get the Description formula value
	var description string
	if props, ok := page.Properties.(nt.DatabasePageProperties); ok {
		description = ExtractFormulaString(props, "Description")
		l.Debug(fmt.Sprintf("extracted Description formula: %s", description))
	}

	return &InvoiceSplitData{
		PageID:      page.ID,
		Amount:      input.Amount,
		Role:        input.Role,
		Currency:    input.Currency,
		Description: description,
	}, nil
}

// QueryPendingInvoiceSplits queries invoice splits with Status=Pending and Type in (Commission, Bonus, Fee, Service Fee)
func (s *InvoiceSplitService) QueryPendingInvoiceSplits(ctx context.Context) ([]PendingCommissionSplit, error) {
	s.logger.Debug("[DEBUG] invoice_split: querying pending invoice splits (Commission, Bonus, Fee, Service Fee)")

	// Build filter: Status=Pending AND (Type=Commission OR Type=Bonus OR Type=Fee OR Type=Service Fee)
	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			And: []nt.DatabaseQueryFilter{
				{
					Property: "Status",
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Select: &nt.SelectDatabaseQueryFilter{
							Equals: "Pending",
						},
					},
				},
				{
					Or: []nt.DatabaseQueryFilter{
						{
							Property: "Type",
							DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
								Select: &nt.SelectDatabaseQueryFilter{
									Equals: "Commission",
								},
							},
						},
						{
							Property: "Type",
							DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
								Select: &nt.SelectDatabaseQueryFilter{
									Equals: "Bonus",
								},
							},
						},
						{
							Property: "Type",
							DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
								Select: &nt.SelectDatabaseQueryFilter{
									Equals: "Fee",
								},
							},
						},
						{
							Property: "Type",
							DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
								Select: &nt.SelectDatabaseQueryFilter{
									Equals: "Service Fee",
								},
							},
						},
					},
				},
			},
		},
		PageSize: 100,
	}

	var splits []PendingCommissionSplit

	// Query with pagination
	for {
		s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: executing invoice split query on database=%s", InvoiceSplitsDBID))

		resp, err := s.client.QueryDatabase(ctx, InvoiceSplitsDBID, query)
		if err != nil {
			s.logger.Error(err, fmt.Sprintf("[DEBUG] invoice_split: failed to query database for invoice splits: %v", err))
			return nil, fmt.Errorf("failed to query invoice splits database: %w", err)
		}

		s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: found %d invoice splits in this page", len(resp.Results)))

		for _, page := range resp.Results {
			props, ok := page.Properties.(nt.DatabasePageProperties)
			if !ok {
				s.logger.Debug("[DEBUG] invoice_split: failed to cast page properties for invoice split")
				continue
			}

			split := PendingCommissionSplit{
				PageID:       page.ID,
				Name:         ExtractTitle(props, "Name"),
				AutoName:     ExtractFormulaString(props, "Auto Name"),
				Amount:       ExtractNumber(props, "Amount"),
				Currency:     ExtractSelect(props, "Currency"),
				Role:         ExtractSelect(props, "Role"),
				Type:         ExtractSelect(props, "Type"),
				PersonPageID: ExtractFirstRelationID(props, "Person"),
				Month:        ExtractDateString(props, "Month"),
				Description:  ExtractFormulaString(props, "Description"),
			}

			s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: parsed invoice split pageID=%s name=%s type=%s amount=%.2f currency=%s role=%s personID=%s month=%s notes=%s",
				split.PageID, split.Name, split.Type, split.Amount, split.Currency, split.Role, split.PersonPageID, split.Month, split.Description))

			splits = append(splits, split)
		}

		if !resp.HasMore || resp.NextCursor == nil {
			break
		}

		query.StartCursor = *resp.NextCursor
		s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: fetching next page of invoice splits with cursor=%s", *resp.NextCursor))
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: total pending invoice splits found=%d", len(splits)))

	return splits, nil
}

// UpdateInvoiceSplitStatus updates an invoice split's Status to a new value
// CRITICAL: Invoice Split uses Select property type (NOT Status type like other tables)
func (s *InvoiceSplitService) UpdateInvoiceSplitStatus(ctx context.Context, pageID string, status string) error {
	if pageID == "" {
		return errors.New("invoice split page ID is empty")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: updating status pageID=%s status=%s", pageID, status))

	// IMPORTANT: Invoice Split uses Select type for Status, not Status type
	// This is different from other tables (Contractor Payables, Contractor Payouts, Refund Request)
	params := nt.UpdatePageParams{
		DatabasePageProperties: nt.DatabasePageProperties{
			"Status": nt.DatabasePageProperty{
				Select: &nt.SelectOptions{
					Name: status,
				},
			},
		},
	}

	_, err := s.client.UpdatePage(ctx, pageID, params)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] invoice_split: failed to update status pageID=%s: %v", pageID, err))
		return fmt.Errorf("failed to update invoice split status: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: updated pageID=%s status=%s successfully", pageID, status))

	return nil
}

// InvoiceSplitSyncData represents syncable data from an Invoice Split record
// Extensible struct - add more fields as needed for syncing
type InvoiceSplitSyncData struct {
	PageID      string
	Description string  // From "Description" formula field
	Amount      float64 // For future use
}

// GetInvoiceSplitSyncData fetches syncable data from an Invoice Split record
// Returns struct with all syncable fields for use in payout syncing
func (s *InvoiceSplitService) GetInvoiceSplitSyncData(ctx context.Context, pageID string) (*InvoiceSplitSyncData, error) {
	if pageID == "" {
		return nil, errors.New("invoice split page ID is empty")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: fetching sync data for pageID=%s", pageID))

	page, err := s.client.FindPageByID(ctx, pageID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] invoice_split: failed to fetch page=%s: %v", pageID, err))
		return nil, fmt.Errorf("failed to fetch invoice split page: %w", err)
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Debug("[DEBUG] invoice_split: failed to cast page properties")
		return nil, errors.New("failed to cast invoice split page properties")
	}

	data := &InvoiceSplitSyncData{
		PageID:      pageID,
		Description: ExtractFormulaString(props, "Description"),
		Amount:      ExtractNumber(props, "Amount"),
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: sync data pageID=%s description=%s amount=%.2f",
		data.PageID, data.Description, data.Amount))

	return data, nil
}
