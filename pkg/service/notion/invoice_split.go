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
	client *nt.Client
	cfg    *config.Config
	logger logger.Logger
}

// InvoiceSplitData represents invoice split data from Notion
type InvoiceSplitData struct {
	PageID   string
	Amount   float64
	Role     string // Sales, Account Manager, etc.
	Currency string
}

// PendingCommissionSplit represents a pending commission split for payout processing
type PendingCommissionSplit struct {
	PageID       string
	Name         string
	Amount       float64
	Currency     string
	Role         string
	PersonPageID string // From Person relation
	Month        string // Date in YYYY-MM-DD format from Month property
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
}

// NewInvoiceSplitService creates a new Notion invoice split service
func NewInvoiceSplitService(cfg *config.Config, logger logger.Logger) *InvoiceSplitService {
	if cfg.Notion.Secret == "" {
		logger.Error(errors.New("notion secret not configured"), "notion secret is empty")
		return nil
	}

	logger.Debug("creating new InvoiceSplitService")

	return &InvoiceSplitService{
		client: nt.NewClient(cfg.Notion.Secret),
		cfg:    cfg,
		logger: logger,
	}
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
	fmt.Printf("[DEBUG] invoice_split: Available properties for page %s:\n", splitPageID)
	for propName := range props {
		fmt.Printf("[DEBUG]   - %s\n", propName)
	}

	// Extract invoice split data
	data := &InvoiceSplitData{
		PageID:   splitPageID,
		Amount:   s.extractNumber(props, "Amount"),
		Role:     s.extractSelect(props, "Role"),
		Currency: s.extractSelect(props, "Currency"),
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: parsed data pageID=%s role=%s amount=%.2f currency=%s",
		data.PageID, data.Role, data.Amount, data.Currency))

	return data, nil
}

// Helper functions for extracting properties

func (s *InvoiceSplitService) extractNumber(props nt.DatabasePageProperties, propName string) float64 {
	prop, ok := props[propName]
	if !ok || prop.Number == nil {
		s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: number property %s not found or nil", propName))
		return 0
	}
	s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: number %s value=%.2f", propName, *prop.Number))
	return *prop.Number
}

func (s *InvoiceSplitService) extractSelect(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Select == nil {
		s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: select property %s not found or nil", propName))
		return ""
	}
	s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: select %s value=%s", propName, prop.Select.Name))
	return prop.Select.Name
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

	if input.InvoicePageID != "" {
		props["Client Invoices"] = nt.DatabasePageProperty{
			Relation: []nt.Relation{
				{ID: input.InvoicePageID},
			},
		}
	}

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

	return &InvoiceSplitData{
		PageID:   page.ID,
		Amount:   input.Amount,
		Role:     input.Role,
		Currency: input.Currency,
	}, nil
}

// QueryPendingCommissionSplits queries invoice splits with Status=Pending and Type=Commission
func (s *InvoiceSplitService) QueryPendingCommissionSplits(ctx context.Context) ([]PendingCommissionSplit, error) {
	s.logger.Debug("[DEBUG] invoice_split: querying pending commission splits")

	// Build filter: Status=Pending AND Type=Commission
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
					Property: "Type",
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Select: &nt.SelectDatabaseQueryFilter{
							Equals: "Commission",
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
		s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: executing query on database=%s", InvoiceSplitsDBID))

		resp, err := s.client.QueryDatabase(ctx, InvoiceSplitsDBID, query)
		if err != nil {
			s.logger.Error(err, fmt.Sprintf("[DEBUG] invoice_split: failed to query database: %v", err))
			return nil, fmt.Errorf("failed to query invoice splits database: %w", err)
		}

		s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: found %d splits in this page", len(resp.Results)))

		for _, page := range resp.Results {
			props, ok := page.Properties.(nt.DatabasePageProperties)
			if !ok {
				s.logger.Debug("[DEBUG] invoice_split: failed to cast page properties")
				continue
			}

			split := PendingCommissionSplit{
				PageID:       page.ID,
				Name:         s.extractTitle(props, "Name"),
				Amount:       s.extractNumber(props, "Amount"),
				Currency:     s.extractSelect(props, "Currency"),
				Role:         s.extractSelect(props, "Role"),
				PersonPageID: s.extractFirstRelationID(props, "Person"),
				Month:        s.extractDate(props, "Month"),
			}

			s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: parsed split pageID=%s name=%s amount=%.2f currency=%s role=%s personID=%s month=%s",
				split.PageID, split.Name, split.Amount, split.Currency, split.Role, split.PersonPageID, split.Month))

			splits = append(splits, split)
		}

		if !resp.HasMore || resp.NextCursor == nil {
			break
		}

		query.StartCursor = *resp.NextCursor
		s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: fetching next page with cursor=%s", *resp.NextCursor))
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: total pending commission splits found=%d", len(splits)))

	return splits, nil
}

// extractTitle extracts title property value
func (s *InvoiceSplitService) extractTitle(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || len(prop.Title) == 0 {
		s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: title property %s not found or empty", propName))
		return ""
	}
	var result string
	for _, rt := range prop.Title {
		result += rt.PlainText
	}
	s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: title %s value=%s", propName, result))
	return result
}

// extractFirstRelationID extracts the first relation ID from a relation property
func (s *InvoiceSplitService) extractFirstRelationID(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || len(prop.Relation) == 0 {
		s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: relation property %s not found or empty", propName))
		return ""
	}
	s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: relation %s first ID=%s", propName, prop.Relation[0].ID))
	return prop.Relation[0].ID
}

// extractDate extracts date property value as YYYY-MM-DD string
func (s *InvoiceSplitService) extractDate(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Date == nil {
		s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: date property %s not found or nil", propName))
		return ""
	}
	dateStr := prop.Date.Start.Format("2006-01-02")
	s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: date %s value=%s", propName, dateStr))
	return dateStr
}
