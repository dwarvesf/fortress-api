package notion

import (
	"context"
	"errors"
	"fmt"

	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

// ContractorFeesService handles contractor fees operations with Notion
type ContractorFeesService struct {
	client *nt.Client
	cfg    *config.Config
	logger logger.Logger
}

// ContractorFeesData represents contractor fees data from Notion
type ContractorFeesData struct {
	PageID           string
	TotalHoursWorked float64 // Rollup from Task Order Log
	HourlyRate       float64 // Rollup from Contractor Rate
	FixedFee         float64 // Rollup from Contractor Rate
	BillingType      string  // Rollup from Contractor Rate: "Monthly Fixed", "Hourly Rate", etc.
	ProofOfWorks     string  // Rollup from Task Order Log (rich text URLs)
	TotalAmount      float64 // Formula: calculated total
	Currency         string
}

// NewContractorFeesService creates a new Notion contractor fees service
func NewContractorFeesService(cfg *config.Config, logger logger.Logger) *ContractorFeesService {
	if cfg.Notion.Secret == "" {
		logger.Error(errors.New("notion secret not configured"), "notion secret is empty")
		return nil
	}

	logger.Debug("creating new ContractorFeesService")

	return &ContractorFeesService{
		client: nt.NewClient(cfg.Notion.Secret),
		cfg:    cfg,
		logger: logger,
	}
}

// GetContractorFeesByID fetches contractor fees data by page ID
func (s *ContractorFeesService) GetContractorFeesByID(ctx context.Context, feesPageID string) (*ContractorFeesData, error) {
	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_fees: fetching page=%s", feesPageID))

	page, err := s.client.FindPageByID(ctx, feesPageID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_fees: failed to fetch page=%s: %v", feesPageID, err))
		return nil, fmt.Errorf("failed to fetch contractor fees page: %w", err)
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Debug("[DEBUG] contractor_fees: failed to cast page properties")
		return nil, errors.New("failed to cast contractor fees page properties")
	}

	// Debug: Log available properties
	fmt.Printf("[DEBUG] contractor_fees: Available properties for page %s:\n", feesPageID)
	for propName := range props {
		fmt.Printf("[DEBUG]   - %s\n", propName)
	}

	// Extract contractor fees data
	data := &ContractorFeesData{
		PageID:           feesPageID,
		TotalHoursWorked: s.extractRollupNumber(props, "Total Hours Worked"),
		HourlyRate:       s.extractRollupNumber(props, "Hourly Rate"),
		FixedFee:         s.extractRollupNumber(props, "Fixed Fee"),
		BillingType:      s.extractRollupSelect(props, "Billing Type"),
		ProofOfWorks:     s.extractRollupRichText(props, "Proof of Works"),
		TotalAmount:      s.extractFormulaNumber(props, "Total Amount"),
		Currency:         s.extractRollupSelect(props, "Currency"),
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_fees: parsed data pageID=%s billingType=%s totalHours=%.2f hourlyRate=%.2f fixedFee=%.2f totalAmount=%.2f currency=%s",
		data.PageID, data.BillingType, data.TotalHoursWorked, data.HourlyRate, data.FixedFee, data.TotalAmount, data.Currency))
	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_fees: proofOfWorks=%s", data.ProofOfWorks))

	return data, nil
}

// Helper functions for extracting properties

func (s *ContractorFeesService) extractRollupNumber(props nt.DatabasePageProperties, propName string) float64 {
	prop, ok := props[propName]
	if !ok || prop.Rollup == nil {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_fees: rollup property %s not found or nil", propName))
		return 0
	}

	// Handle different rollup result types
	switch prop.Rollup.Type {
	case "number":
		if prop.Rollup.Number != nil {
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_fees: rollup %s value=%.2f", propName, *prop.Rollup.Number))
			return *prop.Rollup.Number
		}
	case "array":
		// For array rollups, sum up all numbers
		var sum float64
		for _, item := range prop.Rollup.Array {
			if item.Number != nil {
				sum += *item.Number
			}
		}
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_fees: rollup %s array sum=%.2f", propName, sum))
		return sum
	}

	return 0
}

func (s *ContractorFeesService) extractRollupRichText(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Rollup == nil {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_fees: rollup property %s not found or nil", propName))
		return ""
	}

	var result string
	for _, item := range prop.Rollup.Array {
		if len(item.RichText) > 0 {
			for _, rt := range item.RichText {
				if result != "" {
					result += "\n"
				}
				result += rt.PlainText
			}
		}
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_fees: rollup %s richText length=%d", propName, len(result)))
	return result
}

func (s *ContractorFeesService) extractRollupSelect(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Rollup == nil {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_fees: rollup property %s not found or nil", propName))
		return ""
	}

	// For select rollups, get the first item
	for _, item := range prop.Rollup.Array {
		if item.Select != nil {
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_fees: rollup %s select=%s", propName, item.Select.Name))
			return item.Select.Name
		}
	}

	return ""
}

func (s *ContractorFeesService) extractFormulaNumber(props nt.DatabasePageProperties, propName string) float64 {
	prop, ok := props[propName]
	if !ok || prop.Formula == nil || prop.Formula.Number == nil {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_fees: formula property %s not found or nil", propName))
		return 0
	}
	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_fees: formula %s value=%.2f", propName, *prop.Formula.Number))
	return *prop.Formula.Number
}

// GetTaskOrderLogIDs extracts Task Order Log page IDs from Contractor Fees relation
func (s *ContractorFeesService) GetTaskOrderLogIDs(ctx context.Context, feesPageID string) ([]string, error) {
	s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_fees: fetching Task Order Log IDs for page=%s", feesPageID))

	page, err := s.client.FindPageByID(ctx, feesPageID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("[DEBUG] contractor_fees: failed to fetch page=%s", feesPageID))
		return nil, fmt.Errorf("failed to fetch contractor fees page: %w", err)
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Debug("[DEBUG] contractor_fees: failed to cast page properties")
		return nil, errors.New("failed to cast contractor fees page properties")
	}

	// Debug: Log available properties to find Task Order Log relation name
	fmt.Printf("[DEBUG] contractor_fees: Looking for Task Order Log relation in page %s:\n", feesPageID)
	for propName, prop := range props {
		if len(prop.Relation) > 0 {
			fmt.Printf("[DEBUG]   - %s (Relation, count=%d)\n", propName, len(prop.Relation))
		}
	}

	// Try common property names for Task Order Log relation
	relationNames := []string{"Task Order Log", "Orders", "Order", "Billing"}
	var ids []string

	for _, name := range relationNames {
		prop, ok := props[name]
		if ok && len(prop.Relation) > 0 {
			s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_fees: found relation %s with %d items", name, len(prop.Relation)))
			for _, rel := range prop.Relation {
				ids = append(ids, rel.ID)
				s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_fees: Task Order Log ID=%s", rel.ID))
			}
			break
		}
	}

	if len(ids) == 0 {
		s.logger.Debug("[DEBUG] contractor_fees: no Task Order Log relation found")
	}

	return ids, nil
}

// CheckFeeExistsByTaskOrder checks if a contractor fee entry exists for a given Task Order Log
// Returns (exists bool, existingFeePageID string, error)
func (s *ContractorFeesService) CheckFeeExistsByTaskOrder(ctx context.Context, taskOrderPageID string) (bool, string, error) {
	contractorFeesDBID := s.cfg.Notion.Databases.ContractorFees
	if contractorFeesDBID == "" {
		return false, "", errors.New("contractor fees database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("checking if fee exists for task order: %s", taskOrderPageID))

	// Query Contractor Fees by Task Order Log relation
	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			Property: "Task Order Log",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Relation: &nt.RelationDatabaseQueryFilter{
					Contains: taskOrderPageID,
				},
			},
		},
		PageSize: 1,
	}

	resp, err := s.client.QueryDatabase(ctx, contractorFeesDBID, query)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to check fee existence: taskOrderPageID=%s", taskOrderPageID))
		return false, "", fmt.Errorf("failed to check fee existence: %w", err)
	}

	if len(resp.Results) > 0 {
		existingFeeID := resp.Results[0].ID
		s.logger.Debug(fmt.Sprintf("fee already exists: %s for task order: %s", existingFeeID, taskOrderPageID))
		return true, existingFeeID, nil
	}

	s.logger.Debug(fmt.Sprintf("no fee exists for task order: %s", taskOrderPageID))
	return false, "", nil
}

// NewFeeData represents contractor fee data needed for payout creation
type NewFeeData struct {
	PageID           string  // Contractor Fee page ID
	ContractorPageID string  // From Contractor relation
	ContractorName   string  // From rollup
	TotalAmount      float64 // Formula: calculated total
	Month            string  // YYYY-MM format from Task Order Log
	Date             string  // YYYY-MM-DD from Task Order Log
}

// QueryNewFees queries contractor fees with Payment Status=New
func (s *ContractorFeesService) QueryNewFees(ctx context.Context) ([]*NewFeeData, error) {
	contractorFeesDBID := s.cfg.Notion.Databases.ContractorFees
	if contractorFeesDBID == "" {
		return nil, errors.New("contractor fees database ID not configured")
	}

	s.logger.Debug("querying contractor fees with Payment Status=New")

	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			Property: "Payment Status",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Status: &nt.StatusDatabaseQueryFilter{
					Equals: "New",
				},
			},
		},
		PageSize: 100,
	}

	var fees []*NewFeeData

	for {
		resp, err := s.client.QueryDatabase(ctx, contractorFeesDBID, query)
		if err != nil {
			s.logger.Error(err, "failed to query new contractor fees")
			return nil, fmt.Errorf("failed to query new contractor fees: %w", err)
		}

		s.logger.Debug(fmt.Sprintf("found %d contractor fees with Payment Status=New", len(resp.Results)))

		for _, page := range resp.Results {
			props, ok := page.Properties.(nt.DatabasePageProperties)
			if !ok {
				s.logger.Debug("failed to cast page properties")
				continue
			}

			// Extract fee data
			// Get Task Order Log relation to fetch Contractor
			taskOrderLogIDs := s.extractRelationIDs(props, "Task Order Log")
			var contractorPageID, contractorName string
			if len(taskOrderLogIDs) > 0 {
				// Fetch contractor from first Task Order Log
				contractorPageID, contractorName = s.getContractorFromTaskOrderLog(ctx, taskOrderLogIDs[0])
				s.logger.Debug(fmt.Sprintf("fetched contractor from task order log: pageID=%s name=%s", contractorPageID, contractorName))
			}

			fee := &NewFeeData{
				PageID:           page.ID,
				ContractorPageID: contractorPageID,
				ContractorName:   contractorName,
				TotalAmount:      s.extractFormulaNumber(props, "Total Amount"),
				Month:            s.extractRollupText(props, "Month"),
				Date:             s.extractRollupDate(props, "Date"),
			}

			s.logger.Debug(fmt.Sprintf("parsed fee: pageID=%s contractor=%s amount=%.2f month=%s",
				fee.PageID, fee.ContractorName, fee.TotalAmount, fee.Month))

			fees = append(fees, fee)
		}

		if !resp.HasMore || resp.NextCursor == nil {
			break
		}

		query.StartCursor = *resp.NextCursor
	}

	s.logger.Debug(fmt.Sprintf("total new fees found: %d", len(fees)))
	return fees, nil
}

// UpdatePaymentStatus updates the Payment Status of a contractor fee
func (s *ContractorFeesService) UpdatePaymentStatus(ctx context.Context, feePageID, status string) error {
	s.logger.Debug(fmt.Sprintf("updating fee %s payment status to %s", feePageID, status))

	params := nt.UpdatePageParams{
		DatabasePageProperties: nt.DatabasePageProperties{
			"Payment Status": nt.DatabasePageProperty{
				Status: &nt.SelectOptions{
					Name: status,
				},
			},
		},
	}

	_, err := s.client.UpdatePage(ctx, feePageID, params)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to update payment status for fee %s", feePageID))
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("successfully updated fee %s payment status to %s", feePageID, status))
	return nil
}

func (s *ContractorFeesService) extractRelationIDs(props nt.DatabasePageProperties, propName string) []string {
	prop, ok := props[propName]
	if !ok || len(prop.Relation) == 0 {
		return nil
	}
	var ids []string
	for _, rel := range prop.Relation {
		ids = append(ids, rel.ID)
	}
	return ids
}

// getContractorFromTaskOrderLog fetches contractor page ID and name from a Task Order Log page
func (s *ContractorFeesService) getContractorFromTaskOrderLog(ctx context.Context, taskOrderLogPageID string) (pageID string, name string) {
	s.logger.Debug(fmt.Sprintf("fetching contractor from task order log: %s", taskOrderLogPageID))

	page, err := s.client.FindPageByID(ctx, taskOrderLogPageID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to fetch task order log page: %s", taskOrderLogPageID))
		return "", ""
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Debug("failed to cast task order log page properties")
		return "", ""
	}

	// Debug: log all relation properties in task order log
	s.logger.Debug(fmt.Sprintf("task order log %s available relation properties:", taskOrderLogPageID))
	for propName, prop := range props {
		if len(prop.Relation) > 0 {
			s.logger.Debug(fmt.Sprintf("  - %s (relation, count=%d)", propName, len(prop.Relation)))
		}
	}

	// Get Deployment relation (Task Order Log -> Deployment -> Contractor)
	deploymentProp, ok := props["Deployment"]
	if !ok || len(deploymentProp.Relation) == 0 {
		s.logger.Debug("Deployment relation not found in task order log")
		return "", ""
	}

	deploymentPageID := deploymentProp.Relation[0].ID
	s.logger.Debug(fmt.Sprintf("found Deployment page ID: %s", deploymentPageID))

	// Fetch contractor from Deployment page
	return s.getContractorFromDeployment(ctx, deploymentPageID)
}

// getContractorFromDeployment fetches contractor page ID and name from a Deployment page
func (s *ContractorFeesService) getContractorFromDeployment(ctx context.Context, deploymentPageID string) (pageID string, name string) {
	s.logger.Debug(fmt.Sprintf("fetching contractor from deployment: %s", deploymentPageID))

	page, err := s.client.FindPageByID(ctx, deploymentPageID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to fetch deployment page: %s", deploymentPageID))
		return "", ""
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Debug("failed to cast deployment page properties")
		return "", ""
	}

	// Debug: log all relation properties in deployment
	s.logger.Debug(fmt.Sprintf("deployment %s available relation properties:", deploymentPageID))
	for propName, prop := range props {
		if len(prop.Relation) > 0 {
			s.logger.Debug(fmt.Sprintf("  - %s (relation, count=%d)", propName, len(prop.Relation)))
		}
	}

	// Get Contractor relation from Deployment
	contractorProp, ok := props["Contractor"]
	if !ok || len(contractorProp.Relation) == 0 {
		s.logger.Debug("Contractor relation not found in deployment")
		return "", ""
	}

	contractorPageID := contractorProp.Relation[0].ID
	s.logger.Debug(fmt.Sprintf("found contractor page ID: %s", contractorPageID))

	// Fetch contractor name from contractor page
	contractorPage, err := s.client.FindPageByID(ctx, contractorPageID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to fetch contractor page: %s", contractorPageID))
		return contractorPageID, ""
	}

	contractorProps, ok := contractorPage.Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Debug("failed to cast contractor page properties")
		return contractorPageID, ""
	}

	// Extract Full Name from contractor page
	if nameProp, ok := contractorProps["Full Name"]; ok && len(nameProp.Title) > 0 {
		name = nameProp.Title[0].PlainText
		s.logger.Debug(fmt.Sprintf("found contractor name: %s", name))
	}

	return contractorPageID, name
}

func (s *ContractorFeesService) extractRollupText(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Rollup == nil {
		return ""
	}

	// For array rollups, concatenate all text values
	var result string
	for _, item := range prop.Rollup.Array {
		if len(item.RichText) > 0 {
			for _, rt := range item.RichText {
				if result != "" {
					result += ", "
				}
				result += rt.PlainText
			}
		}
		if len(item.Title) > 0 {
			for _, t := range item.Title {
				if result != "" {
					result += ", "
				}
				result += t.PlainText
			}
		}
	}

	return result
}

func (s *ContractorFeesService) extractRollupDate(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Rollup == nil {
		return ""
	}

	// For array rollups, get the first date
	for _, item := range prop.Rollup.Array {
		if item.Date != nil && !item.Date.Start.Time.IsZero() {
			return item.Date.Start.Time.Format("2006-01-02")
		}
	}

	return ""
}

// CreateContractorFee creates a new Contractor Fee entry in Notion
// Links to the Task Order Log and Contractor Rate, sets Payment Status to "New"
// Returns the created fee page ID
func (s *ContractorFeesService) CreateContractorFee(ctx context.Context, taskOrderPageID, contractorRatePageID string) (string, error) {
	contractorFeesDBID := s.cfg.Notion.Databases.ContractorFees
	if contractorFeesDBID == "" {
		return "", errors.New("contractor fees database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("creating contractor fee: taskOrderPageID=%s contractorRatePageID=%s", taskOrderPageID, contractorRatePageID))

	// Create the Contractor Fee page with relations and Payment Status
	params := nt.CreatePageParams{
		ParentType: nt.ParentTypeDatabase,
		ParentID:   contractorFeesDBID,
		DatabasePageProperties: &nt.DatabasePageProperties{
			"Task Order Log": nt.DatabasePageProperty{
				Type: nt.DBPropTypeRelation,
				Relation: []nt.Relation{
					{ID: taskOrderPageID},
				},
			},
			"Contractor Rate": nt.DatabasePageProperty{
				Type: nt.DBPropTypeRelation,
				Relation: []nt.Relation{
					{ID: contractorRatePageID},
				},
			},
			"Payment Status": nt.DatabasePageProperty{
				Type: nt.DBPropTypeStatus,
				Status: &nt.SelectOptions{
					Name: "New",
				},
			},
		},
	}

	page, err := s.client.CreatePage(ctx, params)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to create contractor fee: taskOrderPageID=%s", taskOrderPageID))
		return "", fmt.Errorf("failed to create contractor fee: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("successfully created contractor fee: %s for task order: %s", page.ID, taskOrderPageID))
	return page.ID, nil
}
