package notion

import (
	"context"
	"errors"
	"fmt"

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
	PageID           string
	Name             string           // Title/Name of the payout
	PersonPageID     string           // From Person relation
	SourceType       PayoutSourceType // Determined by which relation is set
	Direction        PayoutDirection
	Amount           float64
	Currency         string
	Status           string
	ContractorFeesID string // From Contractor Fees relation
	InvoiceSplitID   string // From Invoice Split relation
	RefundRequestID  string // From Refund Request relation
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

	// Build filter: Person relation contains contractorPageID AND Status=Pending AND Direction=Outgoing
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
				{
					Property: "Direction",
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Select: &nt.SelectDatabaseQueryFilter{
							Equals: string(PayoutDirectionOutgoing),
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
			fmt.Printf("[DEBUG] contractor_payouts: Available properties for page %s:\n", page.ID)
			for propName := range props {
				fmt.Printf("[DEBUG]   - %s\n", propName)
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
			// - "Billing" is the relation to Contractor Fees
			// - "Refund" is the relation to Refund Request
			entry := PayoutEntry{
				PageID:           page.ID,
				Name:             s.extractTitle(props, "Name"),
				PersonPageID:     s.extractFirstRelationID(props, "Person"),
				Direction:        PayoutDirection(s.extractSelect(props, "Direction")),
				Amount:           s.extractNumber(props, "Amount"),
				Currency:         s.extractSelect(props, "Currency"),
				Status:           s.extractStatus(props, "Status"),
				ContractorFeesID: s.extractFirstRelationID(props, "Billing"),
				InvoiceSplitID:   s.extractFirstRelationID(props, "Invoice Split"),
				RefundRequestID:  s.extractFirstRelationID(props, "Refund"),
			}

			// Determine source type based on which relation is set
			entry.SourceType = s.determineSourceType(entry)

			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: parsed entry pageID=%s name=%s sourceType=%s direction=%s amount=%.2f currency=%s",
				entry.PageID, entry.Name, entry.SourceType, entry.Direction, entry.Amount, entry.Currency))

			payouts = append(payouts, entry)
		}

		if !resp.HasMore || resp.NextCursor == nil {
			break
		}

		query.StartCursor = *resp.NextCursor
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: fetching next page with cursor=%s", *resp.NextCursor))
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: total pending payouts found=%d", len(payouts)))

	return payouts, nil
}

// determineSourceType determines the source type based on which relation is set
func (s *ContractorPayoutsService) determineSourceType(entry PayoutEntry) PayoutSourceType {
	if entry.ContractorFeesID != "" {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: sourceType=ContractorPayroll (ContractorFeesID=%s)", entry.ContractorFeesID))
		return PayoutSourceTypeContractorPayroll
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

// CheckPayoutExistsByContractorFee checks if a payout already exists for a given contractor fee
// Returns (exists bool, existingPayoutPageID string, error)
func (s *ContractorPayoutsService) CheckPayoutExistsByContractorFee(ctx context.Context, contractorFeePageID string) (bool, string, error) {
	payoutsDBID := s.cfg.Notion.Databases.ContractorPayouts
	if payoutsDBID == "" {
		return false, "", errors.New("contractor payouts database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: checking if payout exists for fee=%s", contractorFeePageID))

	// Query Payouts by Billing relation
	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			Property: "Billing",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Relation: &nt.RelationDatabaseQueryFilter{
					Contains: contractorFeePageID,
				},
			},
		},
		PageSize: 1,
	}

	resp, err := s.client.QueryDatabase(ctx, payoutsDBID, query)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_payouts: failed to check payout existence for fee=%s", contractorFeePageID))
		return false, "", fmt.Errorf("failed to check payout existence: %w", err)
	}

	if len(resp.Results) > 0 {
		existingPayoutID := resp.Results[0].ID
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: payout already exists: %s for fee: %s", existingPayoutID, contractorFeePageID))
		return true, existingPayoutID, nil
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: no payout exists for fee: %s", contractorFeePageID))
	return false, "", nil
}

// CreatePayoutInput contains the input data for creating a new payout
type CreatePayoutInput struct {
	Name              string  // Title/Name of the payout
	ContractorPageID  string  // Person relation
	ContractorFeeID   string  // Billing relation (links to Contractor Fees)
	Amount            float64 // Payment amount
	Currency          string  // Currency (e.g., "VND", "USD")
	Month             string  // YYYY-MM format
	Date              string  // Date in YYYY-MM-DD format
	Type              string  // Payout type (e.g., "Contractor Payroll", "Commission", "Refund")
}

// CreatePayout creates a new payout entry in the Contractor Payouts database
// Returns the created page ID
func (s *ContractorPayoutsService) CreatePayout(ctx context.Context, input CreatePayoutInput) (string, error) {
	payoutsDBID := s.cfg.Notion.Databases.ContractorPayouts
	if payoutsDBID == "" {
		return "", errors.New("contractor payouts database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: creating payout name=%s contractor=%s fee=%s amount=%.2f month=%s",
		input.Name, input.ContractorPageID, input.ContractorFeeID, input.Amount, input.Month))

	// Build properties for the new payout
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
		// Month (rich text)
		"Month": nt.DatabasePageProperty{
			RichText: []nt.RichText{
				{Text: &nt.Text{Content: input.Month}},
			},
		},
		// Status: Pending
		"Status": nt.DatabasePageProperty{
			Status: &nt.SelectOptions{
				Name: "Pending",
			},
		},
		// Type: from input (default: Contractor Payroll)
		"Type": nt.DatabasePageProperty{
			Select: &nt.SelectOptions{
				Name: input.Type,
			},
		},
		// Direction: Outgoing (you pay)
		"Direction": nt.DatabasePageProperty{
			Select: &nt.SelectOptions{
				Name: string(PayoutDirectionOutgoing),
			},
		},
		// Person relation (Contractor)
		"Person": nt.DatabasePageProperty{
			Relation: []nt.Relation{
				{ID: input.ContractorPageID},
			},
		},
		// Billing relation (Contractor Fees)
		"Billing": nt.DatabasePageProperty{
			Relation: []nt.Relation{
				{ID: input.ContractorFeeID},
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
	RefundRequestID  string  // Refund Request relation
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

	// Query Payouts by Refund relation (actual property name in Notion is "Refund")
	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			Property: "Refund",
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
		// Type: Refund
		"Type": nt.DatabasePageProperty{
			Select: &nt.SelectOptions{
				Name: "Refund",
			},
		},
		// Direction: Outgoing
		"Direction": nt.DatabasePageProperty{
			Select: &nt.SelectOptions{
				Name: string(PayoutDirectionOutgoing),
			},
		},
		// Person relation (Contractor)
		"Person": nt.DatabasePageProperty{
			Relation: []nt.Relation{
				{ID: input.ContractorPageID},
			},
		},
		// Refund relation (actual property name in Notion is "Refund")
		"Refund": nt.DatabasePageProperty{
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
