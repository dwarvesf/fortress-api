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
