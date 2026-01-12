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

// ContractorRatesService handles contractor rates operations with Notion
type ContractorRatesService struct {
	client *nt.Client
	cfg    *config.Config
	logger logger.Logger
}

// ContractorRateData represents contractor rate data from Notion
type ContractorRateData struct {
	PageID           string
	ContractorPageID string
	ContractorName   string
	Discord          string
	BillingType      string  // "Monthly Fixed", "Hourly Rate", etc.
	MonthlyFixed     float64 // From formula
	HourlyRate       float64 // From number field
	GrossFixed       float64 // From formula
	Currency         string  // "VND", "USD"
	StartDate        *time.Time
	EndDate          *time.Time
	PayDay           int // Pay day of month (1-31)
}

// NewContractorRatesService creates a new Notion contractor rates service
func NewContractorRatesService(cfg *config.Config, logger logger.Logger) *ContractorRatesService {
	if cfg.Notion.Secret == "" {
		logger.Error(errors.New("notion secret not configured"), "notion secret is empty")
		return nil
	}

	logger.Debug("creating new ContractorRatesService")

	return &ContractorRatesService{
		client: nt.NewClient(cfg.Notion.Secret),
		cfg:    cfg,
		logger: logger,
	}
}

// QueryRatesByDiscordAndMonth queries contractor rates by Discord username and month
// Returns the active rate for the given month (Start Date <= month AND (End Date >= month OR End Date is empty))
func (s *ContractorRatesService) QueryRatesByDiscordAndMonth(ctx context.Context, discord, month string) (*ContractorRateData, error) {
	contractorRatesDBID := s.cfg.Notion.Databases.ContractorRates
	if contractorRatesDBID == "" {
		return nil, errors.New("contractor rates database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("querying contractor rates: discord=%s month=%s", discord, month))

	// Parse month to get date range
	monthTime, err := time.Parse("2006-01", month)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to parse month: %s", month))
		return nil, fmt.Errorf("invalid month format: %w", err)
	}

	// Get start of month for date filtering
	startOfMonth := time.Date(monthTime.Year(), monthTime.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, -1)

	s.logger.Debug(fmt.Sprintf("date range for filtering: start=%s end=%s", startOfMonth.Format("2006-01-02"), endOfMonth.Format("2006-01-02")))

	// Build filter for Discord username and active status
	// Filter: Discord contains discord AND Status = Active
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
					Property: "Status",
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Status: &nt.StatusDatabaseQueryFilter{
							Equals: "Active",
						},
					},
				},
			},
		},
		PageSize: 100,
	}

	var matchedRate *ContractorRateData

	// Query with pagination
	for {
		resp, err := s.client.QueryDatabase(ctx, contractorRatesDBID, query)
		if err != nil {
			s.logger.Error(err, fmt.Sprintf("failed to query contractor rates database: discord=%s", discord))
			return nil, fmt.Errorf("failed to query contractor rates database: %w", err)
		}

		s.logger.Debug(fmt.Sprintf("found %d contractor rates entries", len(resp.Results)))

		for _, page := range resp.Results {
			props, ok := page.Properties.(nt.DatabasePageProperties)
			if !ok {
				s.logger.Debug("failed to cast page properties")
				continue
			}

			// Debug: Log all property names
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_rates: Available properties for page %s:", page.ID))
			for propName := range props {
				s.logger.Debug(fmt.Sprintf("[DEBUG]   - %s", propName))
			}

			// Extract Start Date and End Date for filtering
			startDate := s.extractDate(props, "Start Date")
			endDate := s.extractDate(props, "End Date")

			s.logger.Debug(fmt.Sprintf("checking rate: pageID=%s startDate=%v endDate=%v", page.ID, startDate, endDate))

			// Check date range: Start Date <= month AND (End Date >= month OR End Date is empty)
			if startDate != nil && startDate.After(endOfMonth) {
				// Start date is after the month we're looking for
				s.logger.Debug(fmt.Sprintf("skipping rate: start date after month, startDate=%s endOfMonth=%s", startDate.Format("2006-01-02"), endOfMonth.Format("2006-01-02")))
				continue
			}

			if endDate != nil && endDate.Before(startOfMonth) {
				// End date is before the month we're looking for
				s.logger.Debug(fmt.Sprintf("skipping rate: end date before month, endDate=%s startOfMonth=%s", endDate.Format("2006-01-02"), startOfMonth.Format("2006-01-02")))
				continue
			}

			// Extract contractor page ID
			contractorPageID := s.extractFirstRelationID(props, "Contractor")
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_rates: contractorPageID=%s", contractorPageID))

			// Fetch contractor name from Contractor page
			contractorName := ""
			if contractorPageID != "" {
				contractorName = s.getContractorName(ctx, contractorPageID)
				s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_rates: fetched contractorName=%s", contractorName))
			}

			// Extract Payday (Select type with values like "01", "15")
			payDayStr := s.extractSelect(props, "Payday")
			payDay := 0
			if payDayStr != "" {
				_, _ = fmt.Sscanf(payDayStr, "%d", &payDay)
			}
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_rates: extracted Payday=%s -> payDay=%d", payDayStr, payDay))

			// Extract rate data
			rateData := &ContractorRateData{
				PageID:           page.ID,
				ContractorPageID: contractorPageID,
				ContractorName:   contractorName,
				Discord:          s.extractRollupRichText(props, "Discord"),
				BillingType:      s.extractSelect(props, "Billing Type"),
				MonthlyFixed:     s.extractFormulaNumber(props, "Monthly Fixed"),
				HourlyRate:       s.extractNumber(props, "Hourly Rate"),
				GrossFixed:       s.extractFormulaNumber(props, "Gross Fixed"),
				Currency:         s.extractSelect(props, "Currency"),
				StartDate:        startDate,
				EndDate:          endDate,
				PayDay:           payDay,
			}

			s.logger.Debug(fmt.Sprintf("found matching rate: pageID=%s contractor=%s billingType=%s currency=%s monthlyFixed=%.2f hourlyRate=%.2f",
				rateData.PageID, rateData.ContractorName, rateData.BillingType, rateData.Currency, rateData.MonthlyFixed, rateData.HourlyRate))

			matchedRate = rateData
			break // Take the first matching rate
		}

		if matchedRate != nil || !resp.HasMore || resp.NextCursor == nil {
			break
		}

		query.StartCursor = *resp.NextCursor
	}

	if matchedRate == nil {
		s.logger.Debug(fmt.Sprintf("no active contractor rate found for discord=%s month=%s", discord, month))
		return nil, fmt.Errorf("no active contractor rate found for discord=%s month=%s", discord, month)
	}

	return matchedRate, nil
}

// getContractorName fetches the Full Name from a Contractor page
func (s *ContractorRatesService) getContractorName(ctx context.Context, pageID string) string {
	page, err := s.client.FindPageByID(ctx, pageID)
	if err != nil {
		s.logger.Debug(fmt.Sprintf("[DEBUG] getContractorName: failed to fetch contractor page %s: %v", pageID, err))
		return ""
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Debug(fmt.Sprintf("[DEBUG] getContractorName: failed to cast page properties for %s", pageID))
		return ""
	}

	// Try to get Full Name from Title property
	if prop, ok := props["Full Name"]; ok && len(prop.Title) > 0 {
		name := prop.Title[0].PlainText
		s.logger.Debug(fmt.Sprintf("[DEBUG] getContractorName: found Full Name in Title: %s", name))
		return name
	}

	// Try Name property as fallback
	if prop, ok := props["Name"]; ok && len(prop.Title) > 0 {
		name := prop.Title[0].PlainText
		s.logger.Debug(fmt.Sprintf("[DEBUG] getContractorName: found Name in Title: %s", name))
		return name
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] getContractorName: no Full Name or Name property found for page %s", pageID))
	return ""
}

// Helper functions for extracting properties

func (s *ContractorRatesService) extractRollupRichText(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Rollup == nil {
		return ""
	}

	for _, item := range prop.Rollup.Array {
		if len(item.RichText) > 0 {
			return item.RichText[0].PlainText
		}
	}
	return ""
}

func (s *ContractorRatesService) extractFormulaNumber(props nt.DatabasePageProperties, propName string) float64 {
	prop, ok := props[propName]
	if !ok || prop.Formula == nil || prop.Formula.Number == nil {
		return 0
	}
	return *prop.Formula.Number
}

func (s *ContractorRatesService) extractNumber(props nt.DatabasePageProperties, propName string) float64 {
	prop, ok := props[propName]
	if !ok || prop.Number == nil {
		return 0
	}
	return *prop.Number
}

func (s *ContractorRatesService) extractSelect(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Select == nil {
		return ""
	}
	return prop.Select.Name
}

func (s *ContractorRatesService) extractDate(props nt.DatabasePageProperties, propName string) *time.Time {
	prop, ok := props[propName]
	if !ok || prop.Date == nil {
		return nil
	}
	t := prop.Date.Start.Time
	return &t
}

func (s *ContractorRatesService) extractFirstRelationID(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || len(prop.Relation) == 0 {
		return ""
	}
	return prop.Relation[0].ID
}

// FindActiveRateByContractor finds the active contractor rate for a given contractor at a specific date
// Returns the matching rate or an error if not found
func (s *ContractorRatesService) FindActiveRateByContractor(ctx context.Context, contractorPageID string, orderDate time.Time) (*ContractorRateData, error) {
	contractorRatesDBID := s.cfg.Notion.Databases.ContractorRates
	if contractorRatesDBID == "" {
		return nil, errors.New("contractor rates database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("finding active rate: contractorPageID=%s orderDate=%s", contractorPageID, orderDate.Format("2006-01-02")))

	// Build filter for Contractor relation and Active status
	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
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
		},
		PageSize: 100,
	}

	var matchedRate *ContractorRateData

	// Query with pagination
	for {
		resp, err := s.client.QueryDatabase(ctx, contractorRatesDBID, query)
		if err != nil {
			s.logger.Error(err, fmt.Sprintf("failed to query contractor rates: contractorPageID=%s", contractorPageID))
			return nil, fmt.Errorf("failed to query contractor rates: %w", err)
		}

		s.logger.Debug(fmt.Sprintf("found %d contractor rates entries for contractor %s", len(resp.Results), contractorPageID))

		for _, page := range resp.Results {
			props, ok := page.Properties.(nt.DatabasePageProperties)
			if !ok {
				s.logger.Debug("failed to cast page properties")
				continue
			}

			// Extract Start Date and End Date for filtering
			startDate := s.extractDate(props, "Start Date")
			endDate := s.extractDate(props, "End Date")

			s.logger.Debug(fmt.Sprintf("checking rate: pageID=%s startDate=%v endDate=%v orderDate=%s",
				page.ID, startDate, endDate, orderDate.Format("2006-01-02")))

			// Check date range: Start Date <= orderDate AND (orderDate <= End Date OR End Date is nil)
			// If start date exists and is after order date, skip
			if startDate != nil && startDate.After(orderDate) {
				s.logger.Debug(fmt.Sprintf("skipping rate %s: start date %s is after order date %s",
					page.ID, startDate.Format("2006-01-02"), orderDate.Format("2006-01-02")))
				continue
			}

			// If end date exists and is before order date, skip
			if endDate != nil && endDate.Before(orderDate) {
				s.logger.Debug(fmt.Sprintf("skipping rate %s: end date %s is before order date %s",
					page.ID, endDate.Format("2006-01-02"), orderDate.Format("2006-01-02")))
				continue
			}

			// Rate is valid for this date
			s.logger.Debug(fmt.Sprintf("found valid rate: pageID=%s", page.ID))

			// Extract contractor page ID
			contractorID := s.extractFirstRelationID(props, "Contractor")

			// Fetch contractor name from Contractor page
			contractorName := ""
			if contractorID != "" {
				contractorName = s.getContractorName(ctx, contractorID)
			}

			// Extract Payday (Select type with values like "01", "15")
			payDayStr := s.extractSelect(props, "Payday")
			payDay := 0
			if payDayStr != "" {
				_, _ = fmt.Sscanf(payDayStr, "%d", &payDay)
			}
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_rates: extracted Payday=%s -> payDay=%d", payDayStr, payDay))
			s.logger.Debug(fmt.Sprintf("extracted payDay=%d", payDay))

			// Extract rate data
			matchedRate = &ContractorRateData{
				PageID:           page.ID,
				ContractorPageID: contractorID,
				ContractorName:   contractorName,
				Discord:          s.extractRollupRichText(props, "Discord"),
				BillingType:      s.extractSelect(props, "Billing Type"),
				MonthlyFixed:     s.extractFormulaNumber(props, "Monthly Fixed"),
				HourlyRate:       s.extractNumber(props, "Hourly Rate"),
				GrossFixed:       s.extractFormulaNumber(props, "Gross Fixed"),
				Currency:         s.extractSelect(props, "Currency"),
				StartDate:        startDate,
				EndDate:          endDate,
				PayDay:           payDay,
			}

			s.logger.Debug(fmt.Sprintf("matched rate: pageID=%s contractor=%s billingType=%s currency=%s hourlyRate=%.2f monthlyFixed=%.2f",
				matchedRate.PageID, matchedRate.ContractorName, matchedRate.BillingType, matchedRate.Currency, matchedRate.HourlyRate, matchedRate.MonthlyFixed))

			// Return first matching rate
			return matchedRate, nil
		}

		if !resp.HasMore || resp.NextCursor == nil {
			break
		}

		query.StartCursor = *resp.NextCursor
	}

	// No matching rate found
	s.logger.Debug(fmt.Sprintf("no active contractor rate found for contractor=%s date=%s", contractorPageID, orderDate.Format("2006-01-02")))
	return nil, fmt.Errorf("no active contractor rate found for contractor=%s date=%s", contractorPageID, orderDate.Format("2006-01-02"))
}

// FetchContractorRateByPageID fetches a single Contractor Rate by its page ID.
// Used for hourly rate detection in invoice generation.
func (s *ContractorRatesService) FetchContractorRateByPageID(ctx context.Context, pageID string) (*ContractorRateData, error) {
	s.logger.Debug(fmt.Sprintf("[HOURLY_RATE] fetching contractor rate: pageID=%s", pageID))

	// Step 1: Fetch the page by ID using Notion client
	page, err := s.client.FindPageByID(ctx, pageID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to fetch contractor rate page: %s", pageID))
		return nil, fmt.Errorf("failed to fetch contractor rate page: %w", err)
	}

	// Step 2: Cast page properties to database properties
	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		return nil, fmt.Errorf("failed to cast page properties for contractor rate: %s", pageID)
	}

	// Step 3: Extract contractor page ID from relation
	contractorPageID := s.extractFirstRelationID(props, "Contractor")

	// Step 4: Fetch contractor name if contractor page ID available
	contractorName := ""
	if contractorPageID != "" {
		contractorName = s.getContractorName(ctx, contractorPageID)
	}

	// Step 5: Extract all rate data fields
	rateData := &ContractorRateData{
		PageID:           page.ID,
		ContractorPageID: contractorPageID,
		ContractorName:   contractorName,
		Discord:          s.extractRollupRichText(props, "Discord"),
		BillingType:      s.extractSelect(props, "Billing Type"),
		MonthlyFixed:     s.extractFormulaNumber(props, "Monthly Fixed"),
		HourlyRate:       s.extractNumber(props, "Hourly Rate"),
		GrossFixed:       s.extractFormulaNumber(props, "Gross Fixed"),
		Currency:         s.extractSelect(props, "Currency"),
		StartDate:        s.extractDate(props, "Start Date"),
		EndDate:          s.extractDate(props, "End Date"),
	}

	// Step 6: Log extracted data for debugging
	s.logger.Debug(fmt.Sprintf("[HOURLY_RATE] fetched rate: billingType=%s hourlyRate=%.2f currency=%s",
		rateData.BillingType, rateData.HourlyRate, rateData.Currency))

	return rateData, nil
}
