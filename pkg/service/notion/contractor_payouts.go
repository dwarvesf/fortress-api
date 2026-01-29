package notion

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

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
	Name            string           // Title/Name of the payout (Auto Name formula)
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

	// ServiceRateID from "00 Service Rate" relation (for hourly rate detection)
	ServiceRateID string
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
// If month is provided (format: YYYY-MM), filters payouts by that month
func (s *ContractorPayoutsService) QueryPendingPayoutsByContractor(ctx context.Context, contractorPageID string, month string) ([]PayoutEntry, error) {
	payoutsDBID := s.cfg.Notion.Databases.ContractorPayouts
	if payoutsDBID == "" {
		return nil, errors.New("contractor payouts database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: querying pending payouts for contractor=%s month=%s", contractorPageID, month))

	// Build filter: Person relation contains contractorPageID AND Status=Pending
	// If month provided, also filter by Month formula field
	filters := []nt.DatabaseQueryFilter{
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
	}

	// Add month filter if provided
	if month != "" {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: adding month filter=%s", month))
		filters = append(filters, nt.DatabaseQueryFilter{
			Property: "Month",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Formula: &nt.FormulaDatabaseQueryFilter{
					String: &nt.TextPropertyFilter{
						Equals: month,
					},
				},
			},
		})
	}

	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			And: filters,
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
				ServiceRateID:   s.extractFirstRelationID(props, "00 Service Rate"),
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

// GetLatestPayoutDateByDiscord returns the most recent payout Date for a contractor identified by Discord username.
// It returns nil if no payout exists with a Date.
func (s *ContractorPayoutsService) GetLatestPayoutDateByDiscord(ctx context.Context, discord string) (*time.Time, error) {
	if discord == "" {
		return nil, errors.New("discord username is required")
	}

	payoutsDBID := s.cfg.Notion.Databases.ContractorPayouts
	if payoutsDBID == "" {
		return nil, errors.New("contractor payouts database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: querying latest payout date for discord=%s", discord))

	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			And: []nt.DatabaseQueryFilter{
				{
					Property: "Discord",
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Rollup: &nt.RollupDatabaseQueryFilter{
							Any: &nt.DatabaseQueryPropertyFilter{
								RichText: &nt.TextPropertyFilter{
									Contains: discord,
								},
							},
						},
					},
				},
				{
					Property: "Date",
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Date: &nt.DatePropertyFilter{
							IsNotEmpty: true,
						},
					},
				},
			},
		},
		Sorts: []nt.DatabaseQuerySort{
			{
				Property:  "Date",
				Direction: nt.SortDirDesc,
			},
		},
		PageSize: 5,
	}

	resp, err := s.client.QueryDatabase(ctx, payoutsDBID, query)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_payouts: failed to query latest payout date for discord=%s", discord))
		return nil, fmt.Errorf("failed to query contractor payouts database: %w", err)
	}

	var latest *time.Time
	for _, page := range resp.Results {
		props, ok := page.Properties.(nt.DatabasePageProperties)
		if !ok {
			continue
		}

		if date := s.extractDate(props, "Date"); date != nil {
			if latest == nil || date.After(*latest) {
				latest = date
			}
		}
	}

	if latest == nil {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: no payout date found for discord=%s", discord))
		return nil, nil
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: latest payout date found discord=%s date=%s", discord, latest.Format("2006-01-02")))
	return latest, nil
}

// QueryPendingRefundCommissionBeforeDate queries pending Refund/Commission/Other payouts for a contractor
// where Date <= beforeDate. This is used to include older Refund/Commission/Other items in the current invoice.
func (s *ContractorPayoutsService) QueryPendingRefundCommissionBeforeDate(ctx context.Context, contractorPageID string, beforeDate string) ([]PayoutEntry, error) {
	payoutsDBID := s.cfg.Notion.Databases.ContractorPayouts
	if payoutsDBID == "" {
		return nil, errors.New("contractor payouts database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: querying pending Refund/Commission/Other payouts for contractor=%s beforeDate=%s", contractorPageID, beforeDate))

	// Parse beforeDate to time.Time for Notion API
	beforeDateTime, err := time.Parse("2006-01-02", beforeDate)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_payouts: failed to parse beforeDate=%s", beforeDate))
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Build filter: Person = contractorPageID AND Status = Pending AND Date < beforeDate
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
					Property: "Date",
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Date: &nt.DatePropertyFilter{
							OnOrBefore: &beforeDateTime,
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
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: executing Refund/Commission query on database=%s", payoutsDBID))

		resp, err := s.client.QueryDatabase(ctx, payoutsDBID, query)
		if err != nil {
			s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_payouts: failed to query Refund/Commission payouts: %v", err))
			return nil, fmt.Errorf("failed to query contractor payouts database: %w", err)
		}

		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: found %d payout entries before filtering by type", len(resp.Results)))

		for _, page := range resp.Results {
			props, ok := page.Properties.(nt.DatabasePageProperties)
			if !ok {
				s.logger.Debug("[DEBUG] contractor_payouts: failed to cast page properties")
				continue
			}

			// Extract payout entry data
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
				ServiceRateID:   s.extractFirstRelationID(props, "00 Service Rate"),
			}

			// Determine source type based on which relation is set
			entry.SourceType = s.determineSourceType(entry)

			// Only include Refund, Commission, or Other types (exclude Service Fee)
			if entry.SourceType == PayoutSourceTypeServiceFee {
				s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: skipping entry pageID=%s sourceType=%s (Service Fee excluded from before-date query)", entry.PageID, entry.SourceType))
				continue
			}

			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: including entry pageID=%s name=%s sourceType=%s amount=%.2f currency=%s",
				entry.PageID, entry.Name, entry.SourceType, entry.Amount, entry.Currency))

			payouts = append(payouts, entry)
		}

		if !resp.HasMore || resp.NextCursor == nil {
			break
		}

		query.StartCursor = *resp.NextCursor
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: fetching next page with cursor=%s", *resp.NextCursor))
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: total Refund/Commission payouts found before %s: %d", beforeDate, len(payouts)))

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
	s.logger.Debug("[DEBUG] contractor_payouts: parallel FetchInvoiceSplitInfo for Refund/Commission completed")

	return payouts, nil
}

// determineSourceType determines the source type based on which relation is set
func (s *ContractorPayoutsService) determineSourceType(entry PayoutEntry) PayoutSourceType {
	if entry.TaskOrderID != "" {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: sourceType=ServiceFee (TaskOrderID=%s)", entry.TaskOrderID))
		return PayoutSourceTypeServiceFee
	}

	// For InvoiceSplit items: check Description to determine type
	if entry.InvoiceSplitID != "" {
		// Add nil/empty check for safety
		if entry.Description != "" {
			desc := strings.ToLower(entry.Description)

			// Service Fee: Delivery Lead or Account Management roles
			// These keywords match the Notion formula logic
			if strings.Contains(desc, "delivery lead") || strings.Contains(desc, "account management") {
				s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: sourceType=ServiceFee (InvoiceSplitID=%s, keywords found in Description)", entry.InvoiceSplitID))
				return PayoutSourceTypeServiceFee
			}
		}

		// Otherwise: Commission
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: sourceType=Commission (InvoiceSplitID=%s, no keywords)", entry.InvoiceSplitID))
		return PayoutSourceTypeCommission
	}

	if entry.RefundRequestID != "" {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: sourceType=Refund (RefundRequestID=%s)", entry.RefundRequestID))
		return PayoutSourceTypeRefund
	}
	s.logger.Debug("[DEBUG] contractor_payouts: sourceType=Extra Payment (no relation set)")
	return PayoutSourceTypeExtraPayment
}

func (s *ContractorPayoutsService) extractDate(props nt.DatabasePageProperties, propName string) *time.Time {
	if prop, ok := props[propName]; ok && prop.Date != nil {
		return &prop.Date.Start.Time
	}

	return nil
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

// CheckPayoutsExistByContractorFees checks if payouts already exist for multiple task orders at once.
// Returns a map of taskOrderPageID -> existingPayoutPageID for all orders that have payouts.
// This is a batch operation that reduces N individual queries to fewer queries.
func (s *ContractorPayoutsService) CheckPayoutsExistByContractorFees(ctx context.Context, taskOrderPageIDs []string) (map[string]string, error) {
	if len(taskOrderPageIDs) == 0 {
		return make(map[string]string), nil
	}

	payoutsDBID := s.cfg.Notion.Databases.ContractorPayouts
	if payoutsDBID == "" {
		return nil, errors.New("contractor payouts database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("[BATCH_PAYOUT_CHECK] checking payout existence for %d task orders", len(taskOrderPageIDs)))

	// Create a set for quick lookup
	taskOrderSet := make(map[string]bool)
	for _, id := range taskOrderPageIDs {
		taskOrderSet[id] = true
	}

	// Query payouts that have "00 Task Order" relation set (non-empty)
	// We'll filter by our target IDs in memory
	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			Property: "00 Task Order",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Relation: &nt.RelationDatabaseQueryFilter{
					IsNotEmpty: true,
				},
			},
		},
		PageSize: 100,
	}

	results := make(map[string]string) // taskOrderID -> payoutPageID

	// Query with pagination
	for {
		resp, err := s.client.QueryDatabase(ctx, payoutsDBID, query)
		if err != nil {
			s.logger.Error(err, "[BATCH_PAYOUT_CHECK] failed to query contractor payouts database")
			return nil, fmt.Errorf("failed to check payout existence: %w", err)
		}

		s.logger.Debug(fmt.Sprintf("[BATCH_PAYOUT_CHECK] processing page with %d entries", len(resp.Results)))

		for _, page := range resp.Results {
			props, ok := page.Properties.(nt.DatabasePageProperties)
			if !ok {
				continue
			}

			// Extract task order ID from relation
			taskOrderID := s.extractFirstRelationID(props, "00 Task Order")
			if taskOrderID == "" || !taskOrderSet[taskOrderID] {
				continue // Not in our target set
			}

			// Only record the first payout found for each task order
			if _, exists := results[taskOrderID]; !exists {
				results[taskOrderID] = page.ID
				s.logger.Debug(fmt.Sprintf("[BATCH_PAYOUT_CHECK] found payout=%s for taskOrder=%s", page.ID, taskOrderID))
			}

			// Early exit if we found all task orders
			if len(results) == len(taskOrderPageIDs) {
				s.logger.Debug(fmt.Sprintf("[BATCH_PAYOUT_CHECK] found all %d task orders, stopping early", len(results)))
				goto done
			}
		}

		if !resp.HasMore || resp.NextCursor == nil {
			break
		}
		query.StartCursor = *resp.NextCursor
	}

done:
	s.logger.Debug(fmt.Sprintf("[BATCH_PAYOUT_CHECK] completed: found payouts for %d/%d task orders", len(results), len(taskOrderPageIDs)))
	return results, nil
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
	Description      string  // Description (from refund's Description Formatted)
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

// CheckPayoutsExistByRefundRequests checks if payouts already exist for multiple refund requests at once.
// Returns a map of refundRequestPageID -> existingPayoutPageID for all refunds that have payouts.
// This is a batch operation that reduces N individual queries to fewer queries.
func (s *ContractorPayoutsService) CheckPayoutsExistByRefundRequests(ctx context.Context, refundRequestPageIDs []string) (map[string]string, error) {
	if len(refundRequestPageIDs) == 0 {
		return make(map[string]string), nil
	}

	payoutsDBID := s.cfg.Notion.Databases.ContractorPayouts
	if payoutsDBID == "" {
		return nil, errors.New("contractor payouts database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("[BATCH_REFUND_CHECK] checking payout existence for %d refund requests", len(refundRequestPageIDs)))

	// Create a set for quick lookup
	refundSet := make(map[string]bool)
	for _, id := range refundRequestPageIDs {
		refundSet[id] = true
	}

	// Query payouts that have "01 Refund" relation set (non-empty)
	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			Property: "01 Refund",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Relation: &nt.RelationDatabaseQueryFilter{
					IsNotEmpty: true,
				},
			},
		},
		PageSize: 100,
	}

	results := make(map[string]string) // refundRequestID -> payoutPageID

	// Query with pagination
	for {
		resp, err := s.client.QueryDatabase(ctx, payoutsDBID, query)
		if err != nil {
			s.logger.Error(err, "[BATCH_REFUND_CHECK] failed to query contractor payouts database")
			return nil, fmt.Errorf("failed to check payout existence: %w", err)
		}

		s.logger.Debug(fmt.Sprintf("[BATCH_REFUND_CHECK] processing page with %d entries", len(resp.Results)))

		for _, page := range resp.Results {
			props, ok := page.Properties.(nt.DatabasePageProperties)
			if !ok {
				continue
			}

			// Extract refund request ID from relation
			refundRequestID := s.extractFirstRelationID(props, "01 Refund")
			if refundRequestID == "" || !refundSet[refundRequestID] {
				continue // Not in our target set
			}

			// Only record the first payout found for each refund
			if _, exists := results[refundRequestID]; !exists {
				results[refundRequestID] = page.ID
				s.logger.Debug(fmt.Sprintf("[BATCH_REFUND_CHECK] found payout=%s for refundRequest=%s", page.ID, refundRequestID))
			}

			// Early exit if we found all refund requests
			if len(results) == len(refundRequestPageIDs) {
				s.logger.Debug(fmt.Sprintf("[BATCH_REFUND_CHECK] found all %d refund requests, stopping early", len(results)))
				goto done
			}
		}

		if !resp.HasMore || resp.NextCursor == nil {
			break
		}
		query.StartCursor = *resp.NextCursor
	}

done:
	s.logger.Debug(fmt.Sprintf("[BATCH_REFUND_CHECK] completed: found payouts for %d/%d refund requests", len(results), len(refundRequestPageIDs)))
	return results, nil
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

// CheckPayoutsExistByInvoiceSplits checks if payouts already exist for multiple invoice splits at once.
// Returns a map of invoiceSplitPageID -> existingPayoutPageID for all splits that have payouts.
// This is a batch operation that reduces N individual queries to fewer queries.
func (s *ContractorPayoutsService) CheckPayoutsExistByInvoiceSplits(ctx context.Context, invoiceSplitPageIDs []string) (map[string]string, error) {
	if len(invoiceSplitPageIDs) == 0 {
		return make(map[string]string), nil
	}

	payoutsDBID := s.cfg.Notion.Databases.ContractorPayouts
	if payoutsDBID == "" {
		return nil, errors.New("contractor payouts database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("[BATCH_SPLIT_CHECK] checking payout existence for %d invoice splits", len(invoiceSplitPageIDs)))

	// Create a set for quick lookup
	splitSet := make(map[string]bool)
	for _, id := range invoiceSplitPageIDs {
		splitSet[id] = true
	}

	// Query payouts that have "02 Invoice Split" relation set (non-empty)
	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			Property: "02 Invoice Split",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Relation: &nt.RelationDatabaseQueryFilter{
					IsNotEmpty: true,
				},
			},
		},
		PageSize: 100,
	}

	results := make(map[string]string) // invoiceSplitID -> payoutPageID

	// Query with pagination
	for {
		resp, err := s.client.QueryDatabase(ctx, payoutsDBID, query)
		if err != nil {
			s.logger.Error(err, "[BATCH_SPLIT_CHECK] failed to query contractor payouts database")
			return nil, fmt.Errorf("failed to check payout existence: %w", err)
		}

		s.logger.Debug(fmt.Sprintf("[BATCH_SPLIT_CHECK] processing page with %d entries", len(resp.Results)))

		for _, page := range resp.Results {
			props, ok := page.Properties.(nt.DatabasePageProperties)
			if !ok {
				continue
			}

			// Extract invoice split ID from relation
			invoiceSplitID := s.extractFirstRelationID(props, "02 Invoice Split")
			if invoiceSplitID == "" || !splitSet[invoiceSplitID] {
				continue // Not in our target set
			}

			// Only record the first payout found for each split
			if _, exists := results[invoiceSplitID]; !exists {
				results[invoiceSplitID] = page.ID
				s.logger.Debug(fmt.Sprintf("[BATCH_SPLIT_CHECK] found payout=%s for invoiceSplit=%s", page.ID, invoiceSplitID))
			}

			// Early exit if we found all splits
			if len(results) == len(invoiceSplitPageIDs) {
				s.logger.Debug(fmt.Sprintf("[BATCH_SPLIT_CHECK] found all %d invoice splits, stopping early", len(results)))
				goto done
			}
		}

		if !resp.HasMore || resp.NextCursor == nil {
			break
		}
		query.StartCursor = *resp.NextCursor
	}

done:
	s.logger.Debug(fmt.Sprintf("[BATCH_SPLIT_CHECK] completed: found payouts for %d/%d invoice splits", len(results), len(invoiceSplitPageIDs)))
	return results, nil
}

// CreateCommissionPayoutInput contains the input data for creating a commission payout
type CreateCommissionPayoutInput struct {
	Name             string  // Title/Name of the payout
	ContractorPageID string  // Person relation
	InvoiceSplitID   string  // Invoice Split relation
	Amount           float64 // Payment amount
	Currency         string  // Currency (e.g., "VND", "USD")
	Date             string  // Date in YYYY-MM-DD format
	Description      string  // Description (from Invoice Split Description formula field)
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

	// Add Description if provided (from Invoice Split Description formula field)
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

// PayoutWithRelations contains payout data with related record IDs
// Used by payout commit to determine which related records need status updates
type PayoutWithRelations struct {
	PageID          string
	Status          string
	InvoiceSplitID  string // From "02 Invoice Split" relation (may be empty)
	RefundRequestID string // From "01 Refund" relation (may be empty)
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

// GetPayoutWithRelations fetches a payout record with its related Invoice Split and Refund Request IDs
// Used by payout commit to determine which related records need status updates
func (s *ContractorPayoutsService) GetPayoutWithRelations(ctx context.Context, payoutPageID string) (*PayoutWithRelations, error) {
	if payoutPageID == "" {
		return nil, errors.New("payout page ID is empty")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: fetching payout with relations pageID=%s", payoutPageID))

	page, err := s.client.FindPageByID(ctx, payoutPageID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_payouts: failed to fetch payout pageID=%s: %v", payoutPageID, err))
		return nil, fmt.Errorf("failed to fetch payout: %w", err)
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Debug("[DEBUG] contractor_payouts: failed to cast payout page properties")
		return nil, errors.New("failed to cast payout page properties")
	}

	result := &PayoutWithRelations{
		PageID:          payoutPageID,
		Status:          s.extractStatus(props, "Status"),
		InvoiceSplitID:  s.extractFirstRelationID(props, "02 Invoice Split"),
		RefundRequestID: s.extractFirstRelationID(props, "01 Refund"),
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: payout pageID=%s status=%s invoiceSplit=%s refund=%s",
		payoutPageID, result.Status, result.InvoiceSplitID, result.RefundRequestID))

	return result, nil
}

// UpdatePayoutStatus updates a payout's Status to a new value
// Uses Status property type (not Select)
func (s *ContractorPayoutsService) UpdatePayoutStatus(ctx context.Context, pageID string, status string) error {
	if pageID == "" {
		return errors.New("payout page ID is empty")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: updating payout pageID=%s status=%s", pageID, status))

	params := nt.UpdatePageParams{
		DatabasePageProperties: nt.DatabasePageProperties{
			"Status": nt.DatabasePageProperty{
				Status: &nt.SelectOptions{
					Name: status,
				},
			},
		},
	}

	_, err := s.client.UpdatePage(ctx, pageID, params)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_payouts: failed to update payout status pageID=%s: %v", pageID, err))
		return fmt.Errorf("failed to update payout status: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: updated payout pageID=%s status=%s successfully", pageID, status))

	return nil
}

// PayoutFieldUpdates contains fields that can be updated on a payout
// Extensible struct - only non-nil fields are updated
type PayoutFieldUpdates struct {
	Description *string  // Update if not nil
	Amount      *float64 // Future: Update if not nil
}

// UpdatePayoutFields updates specified fields on a payout
// Only updates fields that have non-nil values in the updates struct
func (s *ContractorPayoutsService) UpdatePayoutFields(ctx context.Context, pageID string, updates PayoutFieldUpdates) error {
	if pageID == "" {
		return errors.New("payout page ID is empty")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: updating payout fields pageID=%s", pageID))

	props := nt.DatabasePageProperties{}

	// Add Description if provided
	if updates.Description != nil {
		props["Description"] = nt.DatabasePageProperty{
			RichText: []nt.RichText{
				{Text: &nt.Text{Content: *updates.Description}},
			},
		}
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: setting description=%s", *updates.Description))
	}

	// Add Amount if provided (for future use)
	if updates.Amount != nil {
		props["Amount"] = nt.DatabasePageProperty{
			Number: updates.Amount,
		}
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: setting amount=%.2f", *updates.Amount))
	}

	// If no fields to update, return early
	if len(props) == 0 {
		s.logger.Debug("[DEBUG] contractor_payouts: no fields to update")
		return nil
	}

	params := nt.UpdatePageParams{
		DatabasePageProperties: props,
	}

	_, err := s.client.UpdatePage(ctx, pageID, params)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_payouts: failed to update payout fields pageID=%s: %v", pageID, err))
		return fmt.Errorf("failed to update payout fields: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: updated payout fields pageID=%s successfully", pageID))

	return nil
}

// QueryPayoutsWithInvoiceSplit queries payouts that have Invoice Split relation
// and are of type Commission or Extra Payment
func (s *ContractorPayoutsService) QueryPayoutsWithInvoiceSplit(ctx context.Context) ([]PayoutEntry, error) {
	payoutsDBID := s.cfg.Notion.Databases.ContractorPayouts
	if payoutsDBID == "" {
		return nil, errors.New("contractor payouts database ID not configured")
	}

	s.logger.Debug("[DEBUG] contractor_payouts: querying payouts with Invoice Split relation (Commission/Extra Payment)")

	// Build filter:
	// "02 Invoice Split" is not empty (has relation)
	// We can't directly filter for "relation is not empty" in Notion API,
	// but we can use "is_not_empty" filter
	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			Property: "02 Invoice Split",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Relation: &nt.RelationDatabaseQueryFilter{
					IsNotEmpty: true,
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

			// Extract payout entry data
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
				ServiceRateID:   s.extractFirstRelationID(props, "00 Service Rate"),
			}

			// Determine source type based on which relation is set
			entry.SourceType = s.determineSourceType(entry)

			// Only include Commission or Extra Payment types
			if entry.SourceType != PayoutSourceTypeCommission && entry.SourceType != PayoutSourceTypeExtraPayment {
				s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: skipping pageID=%s sourceType=%s (not Commission/Extra Payment)", entry.PageID, entry.SourceType))
				continue
			}

			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: including entry pageID=%s name=%s sourceType=%s invoiceSplitID=%s",
				entry.PageID, entry.Name, entry.SourceType, entry.InvoiceSplitID))

			payouts = append(payouts, entry)
		}

		if !resp.HasMore || resp.NextCursor == nil {
			break
		}

		query.StartCursor = *resp.NextCursor
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: fetching next page with cursor=%s", *resp.NextCursor))
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: total payouts with Invoice Split found=%d", len(payouts)))

	return payouts, nil
}

// GetContractorPositions fetches the positions of a contractor from the Notion Contractors database.
// The Position field is a multi-select property containing roles like "Frontend", "Backend", "Product Designer", etc.
// Returns a slice of position names.
func (s *ContractorPayoutsService) GetContractorPositions(ctx context.Context, contractorPageID string) ([]string, error) {
	if contractorPageID == "" {
		return nil, nil
	}

	// Fetch contractor page by ID
	page, err := s.client.FindPageByID(ctx, contractorPageID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to fetch contractor page: %s", contractorPageID))
		return nil, err
	}

	// Extract properties from page
	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Debug(fmt.Sprintf("failed to cast contractor page properties for pageID=%s", contractorPageID))
		return nil, nil
	}

	// Extract Position multi-select property
	positionProp, ok := props["Position"]
	if !ok || positionProp.Type != nt.DBPropTypeMultiSelect {
		s.logger.Debug(fmt.Sprintf("Position property not found or not multi-select for contractor pageID=%s", contractorPageID))
		return nil, nil
	}

	var positions []string
	for _, opt := range positionProp.MultiSelect {
		positions = append(positions, opt.Name)
	}

	s.logger.Debug(fmt.Sprintf("found %d positions for contractor pageID=%s: %v", len(positions), contractorPageID, positions))

	return positions, nil
}

// GetContractorPositionsBatch fetches positions for multiple contractors at once.
// Returns a map of contractorPageID -> []positions.
// This is a batch operation that runs queries in parallel to reduce total time.
func (s *ContractorPayoutsService) GetContractorPositionsBatch(ctx context.Context, contractorPageIDs []string) map[string][]string {
	if len(contractorPageIDs) == 0 {
		return make(map[string][]string)
	}

	s.logger.Debug(fmt.Sprintf("[BATCH_POSITIONS] fetching positions for %d contractors in parallel", len(contractorPageIDs)))

	// Deduplicate contractor IDs
	seen := make(map[string]bool)
	var uniqueIDs []string
	for _, id := range contractorPageIDs {
		if id != "" && !seen[id] {
			seen[id] = true
			uniqueIDs = append(uniqueIDs, id)
		}
	}

	results := make(map[string][]string)
	var mu sync.Mutex
	var wg sync.WaitGroup
	const maxConcurrent = 5
	sem := make(chan struct{}, maxConcurrent)

	for _, contractorID := range uniqueIDs {
		wg.Add(1)
		go func(cID string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			positions, err := s.GetContractorPositions(ctx, cID)
			mu.Lock()
			if err != nil {
				s.logger.Debug(fmt.Sprintf("[BATCH_POSITIONS] failed to fetch positions for contractor=%s: %v", cID, err))
				results[cID] = nil
			} else {
				results[cID] = positions
			}
			mu.Unlock()
		}(contractorID)
	}

	wg.Wait()
	s.logger.Debug(fmt.Sprintf("[BATCH_POSITIONS] completed: fetched positions for %d contractors", len(results)))
	return results
}
