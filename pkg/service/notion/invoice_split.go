package notion

import (
	"context"
	"errors"
	"fmt"

	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

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
