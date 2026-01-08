package notion

import (
	"context"
	"errors"
	"fmt"

	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

// RefundRequestsService handles refund requests operations with Notion
type RefundRequestsService struct {
	client *nt.Client
	cfg    *config.Config
	logger logger.Logger
}

// ApprovedRefundData represents an approved refund request from Notion
type ApprovedRefundData struct {
	PageID           string
	RefundID         string  // Title: Refund ID
	Amount           float64 // Number
	Currency         string  // Select: VND, USD
	ContractorPageID string  // From Contractor relation
	ContractorName   string  // Rollup from Contractor
	Reason           string  // Select: Advance Return, Deduction Reversal, etc.
	Description      string  // Rich text
	DateRequested    string  // Date
}

// NewRefundRequestsService creates a new Notion refund requests service
func NewRefundRequestsService(cfg *config.Config, logger logger.Logger) *RefundRequestsService {
	if cfg.Notion.Secret == "" {
		logger.Error(errors.New("notion secret not configured"), "notion secret is empty")
		return nil
	}

	logger.Debug("creating new RefundRequestsService")

	return &RefundRequestsService{
		client: nt.NewClient(cfg.Notion.Secret),
		cfg:    cfg,
		logger: logger,
	}
}

// QueryApprovedRefunds queries all refund requests with Status=Approved
func (s *RefundRequestsService) QueryApprovedRefunds(ctx context.Context) ([]*ApprovedRefundData, error) {
	refundRequestsDBID := s.cfg.Notion.Databases.RefundRequest
	if refundRequestsDBID == "" {
		return nil, errors.New("refund requests database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] refund_requests: querying approved refunds from database=%s", refundRequestsDBID))

	// Build filter: Status=Approved
	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			Property: "Status",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Status: &nt.StatusDatabaseQueryFilter{
					Equals: "Approved",
				},
			},
		},
		PageSize: 100,
	}

	var refunds []*ApprovedRefundData

	// Query with pagination
	for {
		s.logger.Debug(fmt.Sprintf("[DEBUG] refund_requests: executing query on database=%s", refundRequestsDBID))

		resp, err := s.client.QueryDatabase(ctx, refundRequestsDBID, query)
		if err != nil {
			s.logger.Error(err, fmt.Sprintf("[DEBUG] refund_requests: failed to query database: %v", err))
			return nil, fmt.Errorf("failed to query refund requests database: %w", err)
		}

		s.logger.Debug(fmt.Sprintf("[DEBUG] refund_requests: found %d refund entries in this page", len(resp.Results)))

		for _, page := range resp.Results {
			props, ok := page.Properties.(nt.DatabasePageProperties)
			if !ok {
				s.logger.Debug("[DEBUG] refund_requests: failed to cast page properties")
				continue
			}

			// Extract refund data
			refund := &ApprovedRefundData{
				PageID:           page.ID,
				RefundID:         s.extractTitle(props, "Refund ID"),
				Amount:           s.extractNumber(props, "Amount"),
				Currency:         s.extractSelect(props, "Currency"),
				ContractorPageID: s.extractFirstRelationID(props, "Contractor"),
				ContractorName:   s.extractRollupText(props, "Person"),
				Reason:           s.extractSelect(props, "Reason"),
				Description:      s.extractRichText(props, "Description"),
				DateRequested:    s.extractDate(props, "Date Requested"),
			}

			s.logger.Debug(fmt.Sprintf("[DEBUG] refund_requests: parsed refund pageID=%s refundID=%s amount=%.2f currency=%s contractor=%s reason=%s",
				refund.PageID, refund.RefundID, refund.Amount, refund.Currency, refund.ContractorPageID, refund.Reason))

			refunds = append(refunds, refund)
		}

		if !resp.HasMore || resp.NextCursor == nil {
			break
		}

		query.StartCursor = *resp.NextCursor
		s.logger.Debug(fmt.Sprintf("[DEBUG] refund_requests: fetching next page with cursor=%s", *resp.NextCursor))
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] refund_requests: total approved refunds found=%d", len(refunds)))

	return refunds, nil
}

// Helper functions for extracting properties

func (s *RefundRequestsService) extractTitle(props nt.DatabasePageProperties, propName string) string {
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

func (s *RefundRequestsService) extractNumber(props nt.DatabasePageProperties, propName string) float64 {
	prop, ok := props[propName]
	if !ok || prop.Number == nil {
		return 0
	}
	return *prop.Number
}

func (s *RefundRequestsService) extractSelect(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Select == nil {
		return ""
	}
	return prop.Select.Name
}

func (s *RefundRequestsService) extractFirstRelationID(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || len(prop.Relation) == 0 {
		return ""
	}
	return prop.Relation[0].ID
}

func (s *RefundRequestsService) extractRichText(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || len(prop.RichText) == 0 {
		return ""
	}
	var result string
	for _, rt := range prop.RichText {
		result += rt.PlainText
	}
	return result
}

func (s *RefundRequestsService) extractDate(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Date == nil {
		return ""
	}
	return prop.Date.Start.Format("2006-01-02")
}

func (s *RefundRequestsService) extractRollupText(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Rollup == nil {
		return ""
	}

	// Handle array type rollup (most common for relation rollups)
	if prop.Rollup.Type == "array" && len(prop.Rollup.Array) > 0 {
		// Try to get title from first result
		firstItem := prop.Rollup.Array[0]
		if len(firstItem.Title) > 0 {
			var result string
			for _, rt := range firstItem.Title {
				result += rt.PlainText
			}
			return result
		}
		// Try rich text
		if len(firstItem.RichText) > 0 {
			var result string
			for _, rt := range firstItem.RichText {
				result += rt.PlainText
			}
			return result
		}
	}

	return ""
}

// UpdateRefundRequestStatus updates a refund request's Status to a new value
// Uses Status property type (not Select)
func (s *RefundRequestsService) UpdateRefundRequestStatus(ctx context.Context, pageID string, status string) error {
	if pageID == "" {
		return errors.New("refund request page ID is empty")
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] refund_requests: updating status pageID=%s status=%s", pageID, status))

	// Refund Request uses Status type (same as Contractor Payables and Contractor Payouts)
	// Note: This is different from Invoice Split which uses Select type
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
		s.logger.Error(err, fmt.Sprintf("[DEBUG] refund_requests: failed to update status pageID=%s: %v", pageID, err))
		return fmt.Errorf("failed to update refund request status: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] refund_requests: updated pageID=%s status=%s successfully", pageID, status))

	return nil
}
