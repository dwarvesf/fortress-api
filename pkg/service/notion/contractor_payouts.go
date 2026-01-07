package notion

import (
	"context"
	"errors"
	"fmt"
	"sync"

	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

// ContractorPayoutsService handles contractor payouts operations with Notion
type ContractorPayoutsService struct {
	client *nt.Client
	cfg    *config.Config
	logger logger.Logger
}

// PayoutEntry represents a single payout entry from the Contractor Payouts database
type PayoutEntry struct {
	PageID          string
	Name            string           // Title/Name of the payout
	Description     string           // From Description rich_text field
	PersonPageID    string           // From Person relation
	SourceType      PayoutSourceType // Determined by which relation is set
	Amount          float64
	Currency        string
	Status          string
	TaskOrderID     string // From "00 Task Order" relation (was ContractorFeesID/Billing)
	InvoiceSplitID  string // From "02 Invoice Split" relation
	RefundRequestID string // From "01 Refund" relation
	WorkDetails     string // From "00 Work Details" formula (proof of works for Service Fee)

	// Commission-specific fields (populated from Invoice Split relation)
	CommissionRole    string // From Invoice Split "Role" select (Sales, Account Manager, etc.)
	CommissionProject string // From Invoice Split "Project" rollup (via Deployment)
}

// NewContractorPayoutsService creates a new Notion contractor payouts service
func NewContractorPayoutsService(cfg *config.Config, logger logger.Logger) *ContractorPayoutsService {
	if cfg.Notion.Secret == "" {
		logger.Error(errors.New("notion secret not configured"), "notion secret is empty")
		return nil
	}

	logger.Debug("creating new ContractorPayoutsService")

	return &ContractorPayoutsService{
		client: nt.NewClient(cfg.Notion.Secret),
		cfg:    cfg,
		logger: logger,
	}
}

// QueryPendingPayoutsByContractor queries all pending payouts for a specific contractor
func (s *ContractorPayoutsService) QueryPendingPayoutsByContractor(ctx context.Context, contractorPageID string) ([]PayoutEntry, error) {
	payoutsDBID := s.cfg.Notion.Databases.ContractorPayouts
	if payoutsDBID == "" {
		return nil, errors.New("contractor payouts database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: querying pending payouts for contractor=%s", contractorPageID))

	// Build filter: Person relation contains contractorPageID AND Status=Pending
	// Note: Direction was removed from schema
	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			And: []nt.DatabaseQueryFilter{
				{
					Property: "Person",
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
							Equals: "Pending",
						},
					},
				},
			},
		},
		PageSize: 100,
	}

	var payouts []PayoutEntry

	// Query with pagination
	for {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: executing query on database=%s", payoutsDBID))

		resp, err := s.client.QueryDatabase(ctx, payoutsDBID, query)
		if err != nil {
			s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_payouts: failed to query database: %v", err))
			return nil, fmt.Errorf("failed to query contractor payouts database: %w", err)
		}

		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: found %d payout entries", len(resp.Results)))

		for _, page := range resp.Results {
			props, ok := page.Properties.(nt.DatabasePageProperties)
			if !ok {
				s.logger.Debug("[DEBUG] contractor_payouts: failed to cast page properties")
				continue
			}

			// Debug: Log available properties
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: Available properties for page %s:", page.ID))
			for propName := range props {
				s.logger.Debug(fmt.Sprintf("[DEBUG]   - %s", propName))
			}

			// Debug: log all select properties to find Currency
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: page %s available select properties:", page.ID))
			for propName, prop := range props {
				if prop.Select != nil {
					s.logger.Debug(fmt.Sprintf("[DEBUG]   - %s (select) = %s", propName, prop.Select.Name))
				}
			}

			// Extract payout entry data
			// Note: Property names must match Notion database exactly
			// - "00 Task Order" is the relation to Task Order (was Billing/Contractor Fees)
			// - "01 Refund" is the relation to Refund Request
			// - "02 Invoice Split" is the relation to Invoice Split
			// - "00 Work Details" is a formula that extracts proof of works from task orders
			entry := PayoutEntry{
				PageID:          page.ID,
				Name:            s.extractTitle(props, "Name"),
				Description:     s.extractRichText(props, "Description"),
				PersonPageID:    s.extractFirstRelationID(props, "Person"),
				Amount:          s.extractNumber(props, "Amount"),
				Currency:        s.extractSelect(props, "Currency"),
				Status:          s.extractStatus(props, "Status"),
				TaskOrderID:     s.extractFirstRelationID(props, "00 Task Order"),
				InvoiceSplitID:  s.extractFirstRelationID(props, "02 Invoice Split"),
				RefundRequestID: s.extractFirstRelationID(props, "01 Refund"),
				WorkDetails:     s.extractFormulaString(props, "00 Work Details"),
			}

			// Determine source type based on which relation is set
			entry.SourceType = s.determineSourceType(entry)

			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: parsed entry pageID=%s name=%s sourceType=%s amount=%.2f currency=%s",
				entry.PageID, entry.Name, entry.SourceType, entry.Amount, entry.Currency))

			payouts = append(payouts, entry)
		}

		if !resp.HasMore || resp.NextCursor == nil {
			break
		}

		query.StartCursor = *resp.NextCursor
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: fetching next page with cursor=%s", *resp.NextCursor))
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: total pending payouts found=%d", len(payouts)))

	// Fetch Invoice Split info in parallel for commission payouts
	s.logger.Debug("[DEBUG] contractor_payouts: starting parallel FetchInvoiceSplitInfo for commissions")
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := range payouts {
		if payouts[i].SourceType == PayoutSourceTypeCommission && payouts[i].InvoiceSplitID != "" {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				splitInfo, err := s.FetchInvoiceSplitInfo(ctx, payouts[idx].InvoiceSplitID)
				mu.Lock()
				defer mu.Unlock()
				if err != nil {
					s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: failed to fetch invoice split info for idx=%d: %v", idx, err))
				} else {
					payouts[idx].CommissionRole = splitInfo.Role
					payouts[idx].CommissionProject = splitInfo.Project
					s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: commission info idx=%d - Role=%s Project=%s",
						idx, payouts[idx].CommissionRole, payouts[idx].CommissionProject))
				}
			}(i)
		}
	}

	wg.Wait()
	s.logger.Debug("[DEBUG] contractor_payouts: parallel FetchInvoiceSplitInfo completed")

	return payouts, nil
}

// determineSourceType determines the source type based on which relation is set
func (s *ContractorPayoutsService) determineSourceType(entry PayoutEntry) PayoutSourceType {
	if entry.TaskOrderID != "" {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: sourceType=ServiceFee (TaskOrderID=%s)", entry.TaskOrderID))
		return PayoutSourceTypeServiceFee
	}
	if entry.InvoiceSplitID != "" {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: sourceType=Commission (InvoiceSplitID=%s)", entry.InvoiceSplitID))
		return PayoutSourceTypeCommission
	}
	if entry.RefundRequestID != "" {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: sourceType=Refund (RefundRequestID=%s)", entry.RefundRequestID))
		return PayoutSourceTypeRefund
	}
	s.logger.Debug("[DEBUG] contractor_payouts: sourceType=Other (no relation set)")
	return PayoutSourceTypeOther
}

// Helper functions for extracting properties

func (s *ContractorPayoutsService) extractFirstRelationID(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || len(prop.Relation) == 0 {
		return ""
	}
	return prop.Relation[0].ID
}

func (s *ContractorPayoutsService) extractNumber(props nt.DatabasePageProperties, propName string) float64 {
	prop, ok := props[propName]
	if !ok || prop.Number == nil {
		return 0
	}
	return *prop.Number
}

func (s *ContractorPayoutsService) extractSelect(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Select == nil {
		return ""
	}
	return prop.Select.Name
}

func (s *ContractorPayoutsService) extractStatus(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Status == nil {
		return ""
	}
	return prop.Status.Name
}

func (s *ContractorPayoutsService) extractTitle(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || len(prop.Title) == 0 {
		return ""
	}
	var result string
	for _, rt := range prop.Title {
		result += rt.PlainText
	}
	return result
}

func (s *ContractorPayoutsService) extractRichText(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || len(prop.RichText) == 0 {
		return ""
	}
	var result string
	for _, rt := range prop.RichText {
		result += rt.PlainText
	}
	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: extracted rich text %s length=%d", propName, len(result)))
	return result
}

func (s *ContractorPayoutsService) extractFormulaString(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Formula == nil {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: formula property %s not found or nil", propName))
		return ""
	}
	if prop.Formula.String != nil {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: extracted formula %s value length=%d", propName, len(*prop.Formula.String)))
		return *prop.Formula.String
	}
	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: formula property %s has no string value", propName))
	return ""
}

// CheckPayoutExistsByContractorFee checks if a payout already exists for a given task order (was contractor fee)
// Returns (exists bool, existingPayoutPageID string, error)
func (s *ContractorPayoutsService) CheckPayoutExistsByContractorFee(ctx context.Context, taskOrderPageID string) (bool, string, error) {
	payoutsDBID := s.cfg.Notion.Databases.ContractorPayouts
	if payoutsDBID == "" {
		return false, "", errors.New("contractor payouts database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: checking if payout exists for taskOrder=%s", taskOrderPageID))

	// Query Payouts by "00 Task Order" relation (was "Billing")
	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			Property: "00 Task Order",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Relation: &nt.RelationDatabaseQueryFilter{
					Contains: taskOrderPageID,
				},
			},
		},
		PageSize: 1,
	}

	resp, err := s.client.QueryDatabase(ctx, payoutsDBID, query)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_payouts: failed to check payout existence for taskOrder=%s", taskOrderPageID))
		return false, "", fmt.Errorf("failed to check payout existence: %w", err)
	}

	if len(resp.Results) > 0 {
		existingPayoutID := resp.Results[0].ID
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: payout already exists: %s for taskOrder: %s", existingPayoutID, taskOrderPageID))
		return true, existingPayoutID, nil
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: no payout exists for taskOrder: %s", taskOrderPageID))
	return false, "", nil
}

// CreatePayoutInput contains the input data for creating a new payout
type CreatePayoutInput struct {
	Name             string  // Title/Name of the payout
	ContractorPageID string  // Person relation
	TaskOrderID      string  // "00 Task Order" relation (was ContractorFeeID/Billing)
	ServiceRateID    string  // "00 Service Rate" relation
	Amount           float64 // Payment amount
	Currency         string  // Currency (e.g., "VND", "USD")
	Date             string  // Date in YYYY-MM-DD format
	Description      string  // Optional description/notes
}

// CreatePayout creates a new payout entry in the Contractor Payouts database
// Returns the created page ID
func (s *ContractorPayoutsService) CreatePayout(ctx context.Context, input CreatePayoutInput) (string, error) {
	payoutsDBID := s.cfg.Notion.Databases.ContractorPayouts
	if payoutsDBID == "" {
		return "", errors.New("contractor payouts database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: creating payout name=%s contractor=%s taskOrder=%s amount=%.2f",
		input.Name, input.ContractorPageID, input.TaskOrderID, input.Amount))

	// Build properties for the new payout
	// Note: Type, Month, Direction are now formulas or removed - do not write them
	props := nt.DatabasePageProperties{
		// Title: Payout name
		"Name": nt.DatabasePageProperty{
			Title: []nt.RichText{
				{Text: &nt.Text{Content: input.Name}},
			},
		},
		// Amount
		"Amount": nt.DatabasePageProperty{
			Number: &input.Amount,
		},
		// Status: Pending
		"Status": nt.DatabasePageProperty{
			Status: &nt.SelectOptions{
				Name: "Pending",
			},
		},
		// Person relation (Contractor)
		"Person": nt.DatabasePageProperty{
			Relation: []nt.Relation{
				{ID: input.ContractorPageID},
			},
		},
		// "00 Task Order" relation (was Billing/Contractor Fees)
		"00 Task Order": nt.DatabasePageProperty{
			Relation: []nt.Relation{
				{ID: input.TaskOrderID},
			},
		},
	}

	// Add "00 Service Rate" relation if provided
	if input.ServiceRateID != "" {
		props["00 Service Rate"] = nt.DatabasePageProperty{
			Relation: []nt.Relation{
				{ID: input.ServiceRateID},
			},
		}
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: set 00 Service Rate=%s", input.ServiceRateID))
	}

	// Add Currency if provided
	if input.Currency != "" {
		props["Currency"] = nt.DatabasePageProperty{
			Select: &nt.SelectOptions{
				Name: input.Currency,
			},
		}
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: set currency=%s", input.Currency))
	}

	// Add Date if provided
	if input.Date != "" {
		dateObj, err := nt.ParseDateTime(input.Date)
		if err == nil {
			props["Date"] = nt.DatabasePageProperty{
				Date: &nt.Date{
					Start: dateObj,
				},
			}
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: set date=%s", input.Date))
		} else {
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: failed to parse date=%s: %v", input.Date, err))
		}
	}

	// Add Description if provided
	if input.Description != "" {
		props["Description"] = nt.DatabasePageProperty{
			RichText: []nt.RichText{
				{Text: &nt.Text{Content: input.Description}},
			},
		}
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: set description=%s", input.Description))
	}

	params := nt.CreatePageParams{
		ParentType:             nt.ParentTypeDatabase,
		ParentID:               payoutsDBID,
		DatabasePageProperties: &props,
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: creating page in database=%s", payoutsDBID))

	page, err := s.client.CreatePage(ctx, params)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_payouts: failed to create payout: %v", err))
		return "", fmt.Errorf("failed to create payout: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: created payout pageID=%s", page.ID))

	return page.ID, nil
}

// CreateRefundPayoutInput contains the input data for creating a refund payout
type CreateRefundPayoutInput struct {
	Name             string  // Title/Name of the payout
	ContractorPageID string  // Person relation
	RefundRequestID  string  // "01 Refund" relation
	Amount           float64 // Payment amount
	Currency         string  // Currency (e.g., "VND", "USD")
	Date             string  // Date in YYYY-MM-DD format
}

// CheckPayoutExistsByRefundRequest checks if a payout already exists for a given refund request
// Returns (exists bool, existingPayoutPageID string, error)
func (s *ContractorPayoutsService) CheckPayoutExistsByRefundRequest(ctx context.Context, refundRequestPageID string) (bool, string, error) {
	payoutsDBID := s.cfg.Notion.Databases.ContractorPayouts
	if payoutsDBID == "" {
		return false, "", errors.New("contractor payouts database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: checking if payout exists for refund request=%s", refundRequestPageID))

	// Query Payouts by "01 Refund" relation (was "Refund")
	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			Property: "01 Refund",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Relation: &nt.RelationDatabaseQueryFilter{
					Contains: refundRequestPageID,
				},
			},
		},
		PageSize: 1,
	}

	resp, err := s.client.QueryDatabase(ctx, payoutsDBID, query)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_payouts: failed to check payout existence for refund request=%s", refundRequestPageID))
		return false, "", fmt.Errorf("failed to check payout existence: %w", err)
	}

	if len(resp.Results) > 0 {
		existingPayoutID := resp.Results[0].ID
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: payout already exists: %s for refund request: %s", existingPayoutID, refundRequestPageID))
		return true, existingPayoutID, nil
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: no payout exists for refund request: %s", refundRequestPageID))
	return false, "", nil
}

// CreateRefundPayout creates a new refund payout entry in the Contractor Payouts database
// Returns the created page ID
func (s *ContractorPayoutsService) CreateRefundPayout(ctx context.Context, input CreateRefundPayoutInput) (string, error) {
	payoutsDBID := s.cfg.Notion.Databases.ContractorPayouts
	if payoutsDBID == "" {
		return "", errors.New("contractor payouts database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: creating refund payout name=%s contractor=%s refundRequest=%s amount=%.2f",
		input.Name, input.ContractorPageID, input.RefundRequestID, input.Amount))

	// Build properties for the new payout
	// Note: Type, Month, Direction are now formulas or removed - do not write them
	props := nt.DatabasePageProperties{
		// Title: Payout name
		"Name": nt.DatabasePageProperty{
			Title: []nt.RichText{
				{Text: &nt.Text{Content: input.Name}},
			},
		},
		// Amount
		"Amount": nt.DatabasePageProperty{
			Number: &input.Amount,
		},
		// Status: Pending
		"Status": nt.DatabasePageProperty{
			Status: &nt.SelectOptions{
				Name: "Pending",
			},
		},
		// Person relation (Contractor)
		"Person": nt.DatabasePageProperty{
			Relation: []nt.Relation{
				{ID: input.ContractorPageID},
			},
		},
		// "01 Refund" relation (was "Refund")
		"01 Refund": nt.DatabasePageProperty{
			Relation: []nt.Relation{
				{ID: input.RefundRequestID},
			},
		},
	}

	// Add Currency if provided
	if input.Currency != "" {
		props["Currency"] = nt.DatabasePageProperty{
			Select: &nt.SelectOptions{
				Name: input.Currency,
			},
		}
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: set currency=%s", input.Currency))
	}

	// Add Date if provided
	if input.Date != "" {
		dateObj, err := nt.ParseDateTime(input.Date)
		if err == nil {
			props["Date"] = nt.DatabasePageProperty{
				Date: &nt.Date{
					Start: dateObj,
				},
			}
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: set date=%s", input.Date))
		} else {
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: failed to parse date=%s: %v", input.Date, err))
		}
	}

	params := nt.CreatePageParams{
		ParentType:             nt.ParentTypeDatabase,
		ParentID:               payoutsDBID,
		DatabasePageProperties: &props,
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: creating refund payout page in database=%s", payoutsDBID))

	page, err := s.client.CreatePage(ctx, params)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_payouts: failed to create refund payout: %v", err))
		return "", fmt.Errorf("failed to create refund payout: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: created refund payout pageID=%s", page.ID))

	return page.ID, nil
}

// CheckPayoutExistsByInvoiceSplit checks if a payout already exists for a given invoice split
// Returns (exists bool, existingPayoutPageID string, error)
func (s *ContractorPayoutsService) CheckPayoutExistsByInvoiceSplit(ctx context.Context, invoiceSplitPageID string) (bool, string, error) {
	payoutsDBID := s.cfg.Notion.Databases.ContractorPayouts
	if payoutsDBID == "" {
		return false, "", errors.New("contractor payouts database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: checking if payout exists for invoice split=%s", invoiceSplitPageID))

	// Query Payouts by "02 Invoice Split" relation (was "Invoice Split")
	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			Property: "02 Invoice Split",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Relation: &nt.RelationDatabaseQueryFilter{
					Contains: invoiceSplitPageID,
				},
			},
		},
		PageSize: 1,
	}

	resp, err := s.client.QueryDatabase(ctx, payoutsDBID, query)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_payouts: failed to check payout existence for invoice split=%s", invoiceSplitPageID))
		return false, "", fmt.Errorf("failed to check payout existence: %w", err)
	}

	if len(resp.Results) > 0 {
		existingPayoutID := resp.Results[0].ID
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: payout already exists: %s for invoice split: %s", existingPayoutID, invoiceSplitPageID))
		return true, existingPayoutID, nil
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: no payout exists for invoice split: %s", invoiceSplitPageID))
	return false, "", nil
}

// CreateCommissionPayoutInput contains the input data for creating a commission payout
type CreateCommissionPayoutInput struct {
	Name             string  // Title/Name of the payout
	ContractorPageID string  // Person relation
	InvoiceSplitID   string  // Invoice Split relation
	Amount           float64 // Payment amount
	Currency         string  // Currency (e.g., "VND", "USD")
	Date             string  // Date in YYYY-MM-DD format
	Description      string  // Description/Notes (from Invoice Split Notes)
}

// CreateCommissionPayout creates a new commission payout entry in the Contractor Payouts database
// Returns the created page ID
func (s *ContractorPayoutsService) CreateCommissionPayout(ctx context.Context, input CreateCommissionPayoutInput) (string, error) {
	payoutsDBID := s.cfg.Notion.Databases.ContractorPayouts
	if payoutsDBID == "" {
		return "", errors.New("contractor payouts database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: creating commission payout name=%s contractor=%s invoiceSplit=%s amount=%.2f",
		input.Name, input.ContractorPageID, input.InvoiceSplitID, input.Amount))

	// Build properties for the new payout
	// Note: Type is now a formula (auto-calculated from relations), Direction removed from schema
	props := nt.DatabasePageProperties{
		// Title: Payout name
		"Name": nt.DatabasePageProperty{
			Title: []nt.RichText{
				{Text: &nt.Text{Content: input.Name}},
			},
		},
		// Amount
		"Amount": nt.DatabasePageProperty{
			Number: &input.Amount,
		},
		// Status: Pending
		"Status": nt.DatabasePageProperty{
			Status: &nt.SelectOptions{
				Name: "Pending",
			},
		},
		// Person relation (Contractor)
		"Person": nt.DatabasePageProperty{
			Relation: []nt.Relation{
				{ID: input.ContractorPageID},
			},
		},
		// 02 Invoice Split relation
		"02 Invoice Split": nt.DatabasePageProperty{
			Relation: []nt.Relation{
				{ID: input.InvoiceSplitID},
			},
		},
	}

	// Add Currency if provided
	if input.Currency != "" {
		props["Currency"] = nt.DatabasePageProperty{
			Select: &nt.SelectOptions{
				Name: input.Currency,
			},
		}
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: set currency=%s", input.Currency))
	}

	// Add Date if provided
	if input.Date != "" {
		dateObj, err := nt.ParseDateTime(input.Date)
		if err == nil {
			props["Date"] = nt.DatabasePageProperty{
				Date: &nt.Date{
					Start: dateObj,
				},
			}
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: set date=%s", input.Date))
		} else {
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: failed to parse date=%s: %v", input.Date, err))
		}
	}

	// Add Description if provided (from Invoice Split Notes)
	if input.Description != "" {
		props["Description"] = nt.DatabasePageProperty{
			RichText: []nt.RichText{
				{Text: &nt.Text{Content: input.Description}},
			},
		}
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: set description=%s", input.Description))
	}

	params := nt.CreatePageParams{
		ParentType:             nt.ParentTypeDatabase,
		ParentID:               payoutsDBID,
		DatabasePageProperties: &props,
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: creating commission payout page in database=%s", payoutsDBID))

	page, err := s.client.CreatePage(ctx, params)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_payouts: failed to create commission payout: %v", err))
		return "", fmt.Errorf("failed to create commission payout: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: created commission payout pageID=%s", page.ID))

	return page.ID, nil
}

// InvoiceSplitInfo contains role and project info from an Invoice Split record
type InvoiceSplitInfo struct {
	Role    string // From "Role" select (shortened: DL, AM, Sales, Ref)
	Project string // From "Code" formula (project code via Deployment)
}

// shortenRole converts full role names to short codes
func shortenRole(role string) string {
	switch role {
	case "Delivery Lead":
		return "DL"
	case "Account Manager":
		return "AM"
	case "Hiring Referral":
		return "Ref"
	case "Sales":
		return "Sales"
	default:
		return role
	}
}

// FetchInvoiceSplitInfo fetches Role and Code (project) from an Invoice Split record
func (s *ContractorPayoutsService) FetchInvoiceSplitInfo(ctx context.Context, invoiceSplitID string) (*InvoiceSplitInfo, error) {
	if invoiceSplitID == "" {
		return nil, errors.New("invoice split ID is empty")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: fetching invoice split info for pageID=%s", invoiceSplitID))

	page, err := s.client.FindPageByID(ctx, invoiceSplitID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_payouts: failed to fetch invoice split pageID=%s", invoiceSplitID))
		return nil, fmt.Errorf("failed to fetch invoice split: %w", err)
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Debug("[DEBUG] contractor_payouts: failed to cast invoice split page properties")
		return nil, errors.New("failed to cast invoice split page properties")
	}

	// Debug: log available properties
	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: invoice split %s available properties:", invoiceSplitID))
	for propName := range props {
		s.logger.Debug(fmt.Sprintf("[DEBUG]   - %s", propName))
	}

	// Extract Role from select and Code from formula (project code)
	info := &InvoiceSplitInfo{
		Role:    shortenRole(s.extractSelect(props, "Role")),
		Project: s.extractFormulaString(props, "Code"), // Use Code formula for project grouping
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: invoice split info - Role=%s Project=%s (from Code formula)", info.Role, info.Project))

	return info, nil
}

