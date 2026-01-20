package notion

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
)

// TaskOrderLogService handles task order log operations with Notion
type TaskOrderLogService struct {
	client *nt.Client
	cfg    *config.Config
	logger logger.Logger
}

// NewTaskOrderLogService creates a new Notion task order log service
func NewTaskOrderLogService(cfg *config.Config, logger logger.Logger) *TaskOrderLogService {
	if cfg.Notion.Secret == "" {
		logger.Error(errors.New("notion secret not configured"), "notion secret is empty")
		return nil
	}

	logger.Debug("creating new TaskOrderLogService")

	return &TaskOrderLogService{
		client: nt.NewClient(cfg.Notion.Secret),
		cfg:    cfg,
		logger: logger,
	}
}

// QueryApprovedTimesheetsByMonth queries approved timesheets for a given month
func (s *TaskOrderLogService) QueryApprovedTimesheetsByMonth(ctx context.Context, month string, contractorDiscord string, projectName string) ([]*TimesheetEntry, error) {
	timesheetDBID := s.cfg.Notion.Databases.Timesheet
	if timesheetDBID == "" {
		return nil, errors.New("timesheet database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("querying approved timesheets: month=%s contractor=%s project=%s", month, contractorDiscord, projectName))

	// Build filter using date range instead of Month formula to avoid timezone issues
	// Parse month to get start and end dates
	monthStart, err := time.Parse("2006-01", month)
	if err != nil {
		return nil, fmt.Errorf("invalid month format: %w", err)
	}
	// Use OnOrAfter with first day and OnOrBefore with last day of month
	// This ensures we include exactly the days in the specified month
	startDate := monthStart                                  // First day of month (e.g., Nov 1)
	endDate := monthStart.AddDate(0, 1, 0).AddDate(0, 0, -1) // Last day of month (e.g., Nov 30)

	s.logger.Debug(fmt.Sprintf("filtering timesheets by date range: on_or_after %s, on_or_before %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")))

	filters := []nt.DatabaseQueryFilter{
		{
			Property: "Status",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Status: &nt.StatusDatabaseQueryFilter{
					Equals: "Approved",
				},
			},
		},
		{
			Property: "Date",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Date: &nt.DatePropertyFilter{
					OnOrAfter: &startDate,
				},
			},
		},
		{
			Property: "Date",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Date: &nt.DatePropertyFilter{
					OnOrBefore: &endDate,
				},
			},
		},
	}

	// Add contractor filter if specified
	if contractorDiscord != "" {
		filters = append(filters, nt.DatabaseQueryFilter{
			Property: "Discord",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Rollup: &nt.RollupDatabaseQueryFilter{
					Any: &nt.DatabaseQueryPropertyFilter{
						RichText: &nt.TextPropertyFilter{
							Contains: contractorDiscord,
						},
					},
				},
			},
		})
	}

	// Add project filter if specified (filters by title which contains project code)
	if projectName != "" {
		s.logger.Debug(fmt.Sprintf("adding project filter by title: %s", projectName))
		filters = append(filters, nt.DatabaseQueryFilter{
			Property: "(auto) Timesheet Entry",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Title: &nt.TextPropertyFilter{
					Contains: projectName,
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

	var timesheets []*TimesheetEntry

	// Query with pagination
	for {
		resp, err := s.client.QueryDatabase(ctx, timesheetDBID, query)
		if err != nil {
			s.logger.Error(err, fmt.Sprintf("failed to query timesheet database: month=%s", month))
			return nil, fmt.Errorf("failed to query timesheet database: %w", err)
		}

		for _, page := range resp.Results {
			props, ok := page.Properties.(nt.DatabasePageProperties)
			if !ok {
				continue
			}

			entry := &TimesheetEntry{
				PageID:           page.ID,
				Title:            s.extractTitle(props, "(auto) Entry"),
				ContractorPageID: s.extractFirstRelationID(props, "Contractor"),
				ProjectPageID:    s.extractFirstRelationID(props, "Project"),
				Date:             s.extractDateString(props, "Date"),
				ApproxEffort:     s.extractNumber(props, "Appx. effort"),
				Status:           s.extractStatus(props, "Status"),
			}

			// Extract proof of works
			if prop, ok := props["Key deliverables"]; ok && len(prop.RichText) > 0 {
				var pow string
				for _, rt := range prop.RichText {
					pow += rt.PlainText
				}
				// Store in Title field temporarily for later processing
				entry.Title = pow
			}

			timesheets = append(timesheets, entry)
		}

		if !resp.HasMore || resp.NextCursor == nil {
			break
		}

		query.StartCursor = *resp.NextCursor
	}

	s.logger.Debug(fmt.Sprintf("found %d approved timesheets for month %s", len(timesheets), month))
	return timesheets, nil
}

// GetDeploymentByContractor gets deployment ID for a contractor
func (s *TaskOrderLogService) GetDeploymentByContractor(ctx context.Context, contractorID string) (string, error) {
	deploymentDBID := s.cfg.Notion.Databases.DeploymentTracker
	if deploymentDBID == "" {
		return "", errors.New("deployment tracker database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("querying deployment for contractor: %s", contractorID))

	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			Property: "Contractor",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Relation: &nt.RelationDatabaseQueryFilter{
					Contains: contractorID,
				},
			},
		},
		PageSize: 1,
	}

	resp, err := s.client.QueryDatabase(ctx, deploymentDBID, query)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to query deployment database: contractor=%s", contractorID))
		return "", fmt.Errorf("failed to query deployment database: %w", err)
	}

	if len(resp.Results) == 0 {
		s.logger.Debug(fmt.Sprintf("no deployment found for contractor: %s", contractorID))
		return "", fmt.Errorf("no deployment found for contractor: %s", contractorID)
	}

	deploymentID := resp.Results[0].ID
	s.logger.Debug(fmt.Sprintf("found deployment: %s for contractor: %s", deploymentID, contractorID))
	return deploymentID, nil
}

// GetDeploymentByContractorAndProject gets deployment ID for a contractor and project
func (s *TaskOrderLogService) GetDeploymentByContractorAndProject(ctx context.Context, contractorID, projectID string) (string, error) {
	deploymentDBID := s.cfg.Notion.Databases.DeploymentTracker
	if deploymentDBID == "" {
		return "", errors.New("deployment tracker database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("querying deployment for contractor: %s project: %s", contractorID, projectID))

	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			And: []nt.DatabaseQueryFilter{
				{
					Property: "Contractor",
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Relation: &nt.RelationDatabaseQueryFilter{
							Contains: contractorID,
						},
					},
				},
				{
					Property: "Project",
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Relation: &nt.RelationDatabaseQueryFilter{
							Contains: projectID,
						},
					},
				},
			},
		},
		PageSize: 1,
	}

	resp, err := s.client.QueryDatabase(ctx, deploymentDBID, query)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to query deployment database: contractor=%s project=%s", contractorID, projectID))
		return "", fmt.Errorf("failed to query deployment database: %w", err)
	}

	if len(resp.Results) == 0 {
		s.logger.Debug(fmt.Sprintf("no deployment found for contractor: %s project: %s", contractorID, projectID))
		return "", fmt.Errorf("no deployment found for contractor: %s project: %s", contractorID, projectID)
	}

	deploymentID := resp.Results[0].ID
	s.logger.Debug(fmt.Sprintf("found deployment: %s for contractor: %s project: %s", deploymentID, contractorID, projectID))
	return deploymentID, nil
}

// CheckOrderExists checks if an order already exists for a deployment and month
func (s *TaskOrderLogService) CheckOrderExists(ctx context.Context, deploymentID, month string) (bool, string, error) {
	taskOrderLogDBID := s.cfg.Notion.Databases.TaskOrderLog
	if taskOrderLogDBID == "" {
		return false, "", errors.New("task order log database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("checking if order exists: deployment=%s month=%s", deploymentID, month))

	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			And: []nt.DatabaseQueryFilter{
				{
					Property: "Type",
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Select: &nt.SelectDatabaseQueryFilter{
							Equals: "Order",
						},
					},
				},
				{
					Property: "Month",
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Formula: &nt.FormulaDatabaseQueryFilter{
							String: &nt.TextPropertyFilter{
								Equals: month,
							},
						},
					},
				},
				{
					Property: "Project Deployment",
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Relation: &nt.RelationDatabaseQueryFilter{
							Contains: deploymentID,
						},
					},
				},
			},
		},
		PageSize: 1,
	}

	resp, err := s.client.QueryDatabase(ctx, taskOrderLogDBID, query)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to check order exists: deployment=%s month=%s", deploymentID, month))
		return false, "", fmt.Errorf("failed to check order exists: %w", err)
	}

	if len(resp.Results) > 0 {
		orderID := resp.Results[0].ID
		s.logger.Debug(fmt.Sprintf("order already exists: %s", orderID))
		return true, orderID, nil
	}

	s.logger.Debug("order does not exist")
	return false, "", nil
}

// CheckOrderExistsByContractor checks if an order already exists for a contractor and month
// Since Order doesn't have Deployment, we check by finding any Timesheet line items for this contractor+month
// and return their Parent item (Order)
func (s *TaskOrderLogService) CheckOrderExistsByContractor(ctx context.Context, contractorID, month string) (bool, string, error) {
	taskOrderLogDBID := s.cfg.Notion.Databases.TaskOrderLog
	if taskOrderLogDBID == "" {
		return false, "", errors.New("task order log database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("checking if order exists for contractor: %s month: %s", contractorID, month))

	// Query for Timesheet entries (Type=Timesheet) for this month with this contractor
	// Contractor is via Deployment relation, so we filter by Deployment → Contractor
	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			And: []nt.DatabaseQueryFilter{
				{
					Property: "Type",
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Select: &nt.SelectDatabaseQueryFilter{
							Equals: "Timesheet",
						},
					},
				},
				{
					Property: "Month",
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Formula: &nt.FormulaDatabaseQueryFilter{
							String: &nt.TextPropertyFilter{
								Equals: month,
							},
						},
					},
				},
			},
		},
		PageSize: 100, // Get all timesheet entries for this month
	}

	resp, err := s.client.QueryDatabase(ctx, taskOrderLogDBID, query)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to query timesheet entries: contractor=%s month=%s", contractorID, month))
		return false, "", fmt.Errorf("failed to query timesheet entries: %w", err)
	}

	// Find a timesheet entry that belongs to this contractor
	// Then get its Parent item (Order)
	for _, page := range resp.Results {
		props, ok := page.Properties.(nt.DatabasePageProperties)
		if !ok {
			continue
		}

		// Check if this Timesheet's Deployment belongs to our contractor
		// We need to check via Contractor rollup from Deployment
		if rollup, ok := props["Contractor"]; ok && rollup.Rollup != nil {
			// Rollup array contains relation to contractor
			if len(rollup.Rollup.Array) > 0 {
				for _, item := range rollup.Rollup.Array {
					// Check if this is a relation type with our contractor ID
					if len(item.Relation) > 0 && item.Relation[0].ID == contractorID {
						// Found a timesheet for this contractor
						// Get its Parent item (Order)
						if parentProp, ok := props["Parent item"]; ok && len(parentProp.Relation) > 0 {
							orderID := parentProp.Relation[0].ID
							s.logger.Debug(fmt.Sprintf("order already exists: %s for contractor: %s", orderID, contractorID))
							return true, orderID, nil
						}
					}
				}
			}
		}
	}

	s.logger.Debug(fmt.Sprintf("order does not exist for contractor: %s month: %s", contractorID, month))
	return false, "", nil
}

// CreateOrder creates an Order entry in Task Order Log
// Note: Deployment field is not set for Order type records (ADR-002)
func (s *TaskOrderLogService) CreateOrder(ctx context.Context, month string) (string, error) {
	taskOrderLogDBID := s.cfg.Notion.Databases.TaskOrderLog
	if taskOrderLogDBID == "" {
		return "", errors.New("task order log database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("creating order: month=%s", month))

	// Calculate date from month (last day of month)
	targetDate := time.Now()
	parts := strings.Split(month, "-")
	if len(parts) == 2 {
		year, err1 := strconv.Atoi(parts[0])
		monthInt, err2 := strconv.Atoi(parts[1])
		if err1 == nil && err2 == nil {
			targetDate = timeutil.LastDayOfMonth(monthInt, year)
		} else {
			s.logger.Error(fmt.Errorf("failed to parse month string: %s", month), "using current date as fallback")
		}
	} else {
		s.logger.Error(fmt.Errorf("invalid month format: %s", month), "using current date as fallback")
	}

	properties := &nt.DatabasePageProperties{
		"Type": nt.DatabasePageProperty{
			Type: nt.DBPropTypeSelect,
			Select: &nt.SelectOptions{
				Name: "Order",
			},
		},
		"Status": nt.DatabasePageProperty{
			Type: nt.DBPropTypeSelect,
			Select: &nt.SelectOptions{
				Name: "Draft",
			},
		},
		"Date": nt.DatabasePageProperty{
			Type: nt.DBPropTypeDate,
			Date: &nt.Date{
				Start: nt.NewDateTime(targetDate, false),
			},
		},
	}

	s.logger.Debug(fmt.Sprintf("creating order with Status=Draft: month=%s", month))

	params := nt.CreatePageParams{
		ParentType:             nt.ParentTypeDatabase,
		ParentID:               taskOrderLogDBID,
		DatabasePageProperties: properties,
	}

	page, err := s.client.CreatePage(ctx, params)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to create order: month=%s", month))
		return "", fmt.Errorf("failed to create order: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("created order: %s with Status=Draft", page.ID))
	return page.ID, nil
}

// CheckLineItemExists checks if a line item already exists for an order and deployment
func (s *TaskOrderLogService) CheckLineItemExists(ctx context.Context, orderID, deploymentID string) (bool, string, error) {
	taskOrderLogDBID := s.cfg.Notion.Databases.TaskOrderLog
	if taskOrderLogDBID == "" {
		return false, "", errors.New("task order log database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("checking if line item exists: order=%s deployment=%s", orderID, deploymentID))

	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			And: []nt.DatabaseQueryFilter{
				{
					Property: "Type",
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Select: &nt.SelectDatabaseQueryFilter{
							Equals: "Timesheet",
						},
					},
				},
				{
					Property: "Parent item",
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Relation: &nt.RelationDatabaseQueryFilter{
							Contains: orderID,
						},
					},
				},
				{
					Property: "Project Deployment",
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Relation: &nt.RelationDatabaseQueryFilter{
							Contains: deploymentID,
						},
					},
				},
			},
		},
		PageSize: 1,
	}

	resp, err := s.client.QueryDatabase(ctx, taskOrderLogDBID, query)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to check line item exists: order=%s deployment=%s", orderID, deploymentID))
		return false, "", fmt.Errorf("failed to check line item exists: %w", err)
	}

	if len(resp.Results) > 0 {
		lineItemID := resp.Results[0].ID
		s.logger.Debug(fmt.Sprintf("line item already exists: %s", lineItemID))
		return true, lineItemID, nil
	}

	s.logger.Debug("line item does not exist")
	return false, "", nil
}

// CreateTimesheetLineItem creates a Timesheet sub-item in Task Order Log
func (s *TaskOrderLogService) CreateTimesheetLineItem(ctx context.Context, orderID, deploymentID, projectID string, hours float64, proofOfWorks string, timesheetIDs []string, month string) (string, error) {
	taskOrderLogDBID := s.cfg.Notion.Databases.TaskOrderLog
	if taskOrderLogDBID == "" {
		return "", errors.New("task order log database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("creating timesheet line item: order=%s project=%s hours=%.1f", orderID, projectID, hours))

	// Calculate date from month (last day of month)
	targetDate := time.Now()
	parts := strings.Split(month, "-")
	if len(parts) == 2 {
		year, err1 := strconv.Atoi(parts[0])
		monthInt, err2 := strconv.Atoi(parts[1])
		if err1 == nil && err2 == nil {
			targetDate = timeutil.LastDayOfMonth(monthInt, year)
		} else {
			s.logger.Error(fmt.Errorf("failed to parse month string: %s", month), "using current date as fallback")
		}
	} else {
		s.logger.Error(fmt.Errorf("invalid month format: %s", month), "using current date as fallback")
	}

	// Build timesheet relations
	timesheetRelations := make([]nt.Relation, len(timesheetIDs))
	for i, id := range timesheetIDs {
		timesheetRelations[i] = nt.Relation{ID: id}
	}

	params := nt.CreatePageParams{
		ParentType: nt.ParentTypeDatabase,
		ParentID:   taskOrderLogDBID,
		DatabasePageProperties: &nt.DatabasePageProperties{
			"Type": nt.DatabasePageProperty{
				Type: nt.DBPropTypeSelect,
				Select: &nt.SelectOptions{
					Name: "Timesheet",
				},
			},
			"Status": nt.DatabasePageProperty{
				Type: nt.DBPropTypeSelect,
				Select: &nt.SelectOptions{
					Name: "Pending Feedback",
				},
			},
			"Date": nt.DatabasePageProperty{
				Type: nt.DBPropTypeDate,
				Date: &nt.Date{
					Start: nt.NewDateTime(targetDate, false),
				},
			},
			"Line Item Hours": nt.DatabasePageProperty{
				Type:   nt.DBPropTypeNumber,
				Number: &hours,
			},
			"Key deliverables": nt.DatabasePageProperty{
				Type: nt.DBPropTypeRichText,
				RichText: []nt.RichText{
					{
						Type: nt.RichTextTypeText,
						Text: &nt.Text{
							Content: proofOfWorks,
						},
					},
				},
			},
			"Project Deployment": nt.DatabasePageProperty{
				Type: nt.DBPropTypeRelation,
				Relation: []nt.Relation{
					{ID: deploymentID},
				},
			},
			"Timesheet": nt.DatabasePageProperty{
				Type:     nt.DBPropTypeRelation,
				Relation: timesheetRelations,
			},
		},
	}

	page, err := s.client.CreatePage(ctx, params)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to create timesheet line item: order=%s", orderID))
		return "", fmt.Errorf("failed to create timesheet line item: %w", err)
	}

	lineItemID := page.ID
	s.logger.Debug(fmt.Sprintf("created timesheet line item: %s", lineItemID))

	// Update Order's Sub-item relation to link the line item
	s.logger.Debug(fmt.Sprintf("updating order sub-item relation: order=%s lineItem=%s", orderID, lineItemID))
	err = s.addSubItemToOrder(ctx, orderID, lineItemID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to add sub-item to order: order=%s lineItem=%s", orderID, lineItemID))
		// Don't fail the whole operation, just log the error
	}

	return lineItemID, nil
}

// CreateEmptyTimesheetLineItem creates an empty Timesheet sub-item for initialization
// This is used when pre-creating Line Items at month start before any timesheets exist
func (s *TaskOrderLogService) CreateEmptyTimesheetLineItem(ctx context.Context, orderID, deploymentID string, month string) (string, error) {
	taskOrderLogDBID := s.cfg.Notion.Databases.TaskOrderLog
	if taskOrderLogDBID == "" {
		return "", errors.New("task order log database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("creating empty timesheet line item: order=%s deployment=%s month=%s", orderID, deploymentID, month))

	// Calculate date from month (last day of month)
	targetDate := time.Now()
	parts := strings.Split(month, "-")
	if len(parts) == 2 {
		year, err1 := strconv.Atoi(parts[0])
		monthInt, err2 := strconv.Atoi(parts[1])
		if err1 == nil && err2 == nil {
			targetDate = timeutil.LastDayOfMonth(monthInt, year)
		} else {
			s.logger.Error(fmt.Errorf("failed to parse month string: %s", month), "using current date as fallback")
		}
	} else {
		s.logger.Error(fmt.Errorf("invalid month format: %s", month), "using current date as fallback")
	}

	s.logger.Debug(fmt.Sprintf("using target date: %s for month: %s", targetDate.Format("2006-01-02"), month))

	// Initialize hours to 0
	hours := float64(0)

	s.logger.Debug(fmt.Sprintf("creating empty timesheet line item with Status=Draft: order=%s deployment=%s", orderID, deploymentID))

	params := nt.CreatePageParams{
		ParentType: nt.ParentTypeDatabase,
		ParentID:   taskOrderLogDBID,
		DatabasePageProperties: &nt.DatabasePageProperties{
			"Type": nt.DatabasePageProperty{
				Type: nt.DBPropTypeSelect,
				Select: &nt.SelectOptions{
					Name: "Timesheet",
				},
			},
			"Status": nt.DatabasePageProperty{
				Type: nt.DBPropTypeSelect,
				Select: &nt.SelectOptions{
					Name: "Draft",
				},
			},
			"Date": nt.DatabasePageProperty{
				Type: nt.DBPropTypeDate,
				Date: &nt.Date{
					Start: nt.NewDateTime(targetDate, false),
				},
			},
			"Line Item Hours": nt.DatabasePageProperty{
				Type:   nt.DBPropTypeNumber,
				Number: &hours,
			},
			"Project Deployment": nt.DatabasePageProperty{
				Type: nt.DBPropTypeRelation,
				Relation: []nt.Relation{
					{ID: deploymentID},
				},
			},
			// Note: Timesheet relation is intentionally omitted for initialization
			// It will be populated later by the sync endpoint when timesheets are approved
		},
	}

	page, err := s.client.CreatePage(ctx, params)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to create empty timesheet line item: order=%s deployment=%s", orderID, deploymentID))
		return "", fmt.Errorf("failed to create empty timesheet line item: %w", err)
	}

	lineItemID := page.ID
	s.logger.Debug(fmt.Sprintf("created empty timesheet line item: %s", lineItemID))

	// Update Order's Sub-item relation to link the line item
	s.logger.Debug(fmt.Sprintf("updating order sub-item relation: order=%s lineItem=%s", orderID, lineItemID))
	err = s.addSubItemToOrder(ctx, orderID, lineItemID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to add sub-item to order: order=%s lineItem=%s", orderID, lineItemID))
		// Don't fail the whole operation, just log the error
	}

	s.logger.Debug(fmt.Sprintf("successfully created empty line item: lineItemID=%s for order=%s deployment=%s", lineItemID, orderID, deploymentID))
	return lineItemID, nil
}

// UpdateTimesheetLineItem updates existing line item with new data and resets status to "Pending Feedback"
func (s *TaskOrderLogService) UpdateTimesheetLineItem(ctx context.Context, lineItemID, orderID string, hours float64, proofOfWorks string, timesheetIDs []string) error {
	s.logger.Debug(fmt.Sprintf("updating line item: lineItemID=%s orderID=%s hours=%.2f timesheets=%d", lineItemID, orderID, hours, len(timesheetIDs)))

	// Build timesheet relations
	timesheetRelations := make([]nt.Relation, len(timesheetIDs))
	for i, id := range timesheetIDs {
		timesheetRelations[i] = nt.Relation{ID: id}
	}

	// Update line item properties
	updateParams := nt.UpdatePageParams{
		DatabasePageProperties: nt.DatabasePageProperties{
			"Line Item Hours": nt.DatabasePageProperty{
				Type:   nt.DBPropTypeNumber,
				Number: &hours,
			},
			"Key deliverables": nt.DatabasePageProperty{
				Type: nt.DBPropTypeRichText,
				RichText: []nt.RichText{
					{
						Type: nt.RichTextTypeText,
						Text: &nt.Text{
							Content: proofOfWorks,
						},
					},
				},
			},
			"Timesheet": nt.DatabasePageProperty{
				Type:     nt.DBPropTypeRelation,
				Relation: timesheetRelations,
			},
			"Status": nt.DatabasePageProperty{
				Type: nt.DBPropTypeSelect,
				Select: &nt.SelectOptions{
					Name: "Pending Feedback",
				},
			},
		},
	}

	_, err := s.client.UpdatePage(ctx, lineItemID, updateParams)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to update line item: %s", lineItemID))
		return fmt.Errorf("failed to update line item: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("successfully updated line item: %s", lineItemID))

	// Also update parent Order status to "Pending Feedback"
	s.logger.Debug(fmt.Sprintf("updating parent order status: orderID=%s", orderID))
	err = s.UpdateOrderStatus(ctx, orderID, "Pending Feedback")
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to update order status after line item update: orderID=%s", orderID))
		// Don't fail the whole operation, just log the error
	}

	return nil
}

// GetLineItemDetails fetches existing line item data for comparison during upsert
func (s *TaskOrderLogService) GetLineItemDetails(ctx context.Context, lineItemID string) (*LineItemDetails, error) {
	s.logger.Debug(fmt.Sprintf("fetching line item details: lineItemID=%s", lineItemID))

	page, err := s.client.FindPageByID(ctx, lineItemID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to fetch line item page: %s", lineItemID))
		return nil, fmt.Errorf("failed to fetch line item page: %w", err)
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		return nil, fmt.Errorf("failed to cast line item page properties: %s", lineItemID)
	}

	details := &LineItemDetails{
		PageID: lineItemID,
	}

	// Extract Line Item Hours (number property)
	if prop, ok := props["Line Item Hours"]; ok && prop.Number != nil {
		details.Hours = *prop.Number
		s.logger.Debug(fmt.Sprintf("extracted hours: %.2f", details.Hours))
	}

	// Extract Timesheet relation IDs
	if prop, ok := props["Timesheet"]; ok && len(prop.Relation) > 0 {
		for _, rel := range prop.Relation {
			details.TimesheetIDs = append(details.TimesheetIDs, rel.ID)
		}
		s.logger.Debug(fmt.Sprintf("extracted %d timesheet IDs", len(details.TimesheetIDs)))
	}

	// Extract Status (select property)
	if prop, ok := props["Status"]; ok && prop.Select != nil {
		details.Status = prop.Select.Name
		s.logger.Debug(fmt.Sprintf("extracted status: %s", details.Status))
	}

	s.logger.Debug(fmt.Sprintf("line item details: hours=%.2f timesheets=%d status=%s", details.Hours, len(details.TimesheetIDs), details.Status))
	return details, nil
}

// addSubItemToOrder adds a line item to the Order's Sub-item relation
func (s *TaskOrderLogService) addSubItemToOrder(ctx context.Context, orderID, lineItemID string) error {
	// Get current Order page to read existing Sub-item relations
	order, err := s.client.FindPageByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to fetch order: %w", err)
	}

	// Get existing Sub-item relations
	var existingSubItems []nt.Relation
	if prop, ok := order.Properties.(nt.DatabasePageProperties)["Sub-item"]; ok {
		existingSubItems = prop.Relation
	}

	// Append new line item
	newSubItems := append(existingSubItems, nt.Relation{ID: lineItemID})

	// Update Order page with new Sub-item relation
	updateParams := nt.UpdatePageParams{
		DatabasePageProperties: nt.DatabasePageProperties{
			"Sub-item": nt.DatabasePageProperty{
				Type:     nt.DBPropTypeRelation,
				Relation: newSubItems,
			},
		},
	}

	_, err = s.client.UpdatePage(ctx, orderID, updateParams)
	if err != nil {
		return fmt.Errorf("failed to update order sub-items: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("successfully updated order %s with sub-item %s", orderID, lineItemID))
	return nil
}

// Helper methods reused from TimesheetService

func (s *TaskOrderLogService) extractTitle(props nt.DatabasePageProperties, propName string) string {
	if prop, ok := props[propName]; ok && len(prop.Title) > 0 {
		var parts []string
		for _, rt := range prop.Title {
			parts = append(parts, rt.PlainText)
		}
		return strings.Join(parts, "")
	}
	return ""
}

func (s *TaskOrderLogService) extractStatus(props nt.DatabasePageProperties, propName string) string {
	if prop, ok := props[propName]; ok && prop.Status != nil {
		return prop.Status.Name
	}
	return ""
}

func (s *TaskOrderLogService) extractFirstRelationID(props nt.DatabasePageProperties, propName string) string {
	if prop, ok := props[propName]; ok && len(prop.Relation) > 0 {
		return prop.Relation[0].ID
	}
	return ""
}

func (s *TaskOrderLogService) extractMultiSelectNames(props nt.DatabasePageProperties, propName string) []string {
	var names []string
	if prop, ok := props[propName]; ok && len(prop.MultiSelect) > 0 {
		for _, opt := range prop.MultiSelect {
			names = append(names, opt.Name)
		}
		s.logger.Debug(fmt.Sprintf("extracted multi-select %s: %v", propName, names))
	}
	return names
}

func (s *TaskOrderLogService) extractDateString(props nt.DatabasePageProperties, propName string) string {
	if prop, ok := props[propName]; ok && prop.Date != nil {
		return prop.Date.Start.String()
	}
	return ""
}

func (s *TaskOrderLogService) extractNumber(props nt.DatabasePageProperties, propName string) float64 {
	if prop, ok := props[propName]; ok && prop.Number != nil {
		return *prop.Number
	}
	return 0
}

func (s *TaskOrderLogService) extractRichText(props nt.DatabasePageProperties, propName string) string {
	if prop, ok := props[propName]; ok && len(prop.RichText) > 0 {
		var parts []string
		for _, rt := range prop.RichText {
			parts = append(parts, rt.PlainText)
		}
		return strings.Join(parts, "")
	}
	return ""
}

// extractSelect extracts the value from a Select property
// Returns empty string if property not found or not a Select type
func (s *TaskOrderLogService) extractSelect(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Select == nil {
		return ""
	}
	return prop.Select.Name
}

// LineItemDetails holds line item data for comparison during upsert
type LineItemDetails struct {
	PageID       string
	Hours        float64
	TimesheetIDs []string
	Status       string
}

// OrderSubitem represents a line item (timesheet) in Task Order Log
type OrderSubitem struct {
	PageID      string
	ProjectName string  // From Project rollup
	ProjectID   string  // From Project relation
	Hours       float64 // From Line Item Hours
	ProofOfWork string  // From Key deliverables rich text
}

// QueryOrderSubitems queries timesheet line items (subitems) for a given order
func (s *TaskOrderLogService) QueryOrderSubitems(ctx context.Context, orderPageID string) ([]*OrderSubitem, error) {
	taskOrderLogDBID := s.cfg.Notion.Databases.TaskOrderLog
	if taskOrderLogDBID == "" {
		return nil, errors.New("task order log database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("querying order subitems: orderPageID=%s", orderPageID))

	// Filter by Type="Timesheet" and Parent item=orderPageID
	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			And: []nt.DatabaseQueryFilter{
				{
					Property: "Type",
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Select: &nt.SelectDatabaseQueryFilter{
							Equals: "Timesheet",
						},
					},
				},
				{
					Property: "Parent item",
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Relation: &nt.RelationDatabaseQueryFilter{
							Contains: orderPageID,
						},
					},
				},
			},
		},
		PageSize: 100,
	}

	var subitems []*OrderSubitem

	// Query with pagination
	for {
		resp, err := s.client.QueryDatabase(ctx, taskOrderLogDBID, query)
		if err != nil {
			s.logger.Error(err, fmt.Sprintf("failed to query order subitems: orderPageID=%s", orderPageID))
			return nil, fmt.Errorf("failed to query order subitems: %w", err)
		}

		s.logger.Debug(fmt.Sprintf("found %d subitems in current page", len(resp.Results)))

		for _, page := range resp.Results {
			props, ok := page.Properties.(nt.DatabasePageProperties)
			if !ok {
				s.logger.Debug("failed to cast page properties")
				continue
			}

			// Debug: Log ALL available properties on this subitem page
			s.logger.Debug(fmt.Sprintf("[DEBUG] task_order_log: ===== Subitem page %s properties =====", page.ID))
			for propName, prop := range props {
				s.logger.Debug(fmt.Sprintf("[DEBUG]   Property: %s", propName))
				if len(prop.Relation) > 0 {
					s.logger.Debug(fmt.Sprintf("[DEBUG]     Type: Relation, Count: %d, First ID: %s", len(prop.Relation), prop.Relation[0].ID))
				}
				if prop.Rollup != nil {
					s.logger.Debug(fmt.Sprintf("[DEBUG]     Type: Rollup, Array Length: %d", len(prop.Rollup.Array)))
					for i, item := range prop.Rollup.Array {
						s.logger.Debug(fmt.Sprintf("[DEBUG]       Rollup[%d]: Title=%d, RichText=%d, Relation=%d", i, len(item.Title), len(item.RichText), len(item.Relation)))
						if len(item.Relation) > 0 {
							s.logger.Debug(fmt.Sprintf("[DEBUG]         Relation[0] ID: %s", item.Relation[0].ID))
						}
					}
				}
				if len(prop.Title) > 0 {
					s.logger.Debug(fmt.Sprintf("[DEBUG]     Type: Title, Value: %s", prop.Title[0].PlainText))
				}
				if len(prop.RichText) > 0 {
					s.logger.Debug(fmt.Sprintf("[DEBUG]     Type: RichText, Value: %s", prop.RichText[0].PlainText))
				}
				if prop.Select != nil {
					s.logger.Debug(fmt.Sprintf("[DEBUG]     Type: Select, Value: %s", prop.Select.Name))
				}
				if prop.Number != nil {
					s.logger.Debug(fmt.Sprintf("[DEBUG]     Type: Number, Value: %f", *prop.Number))
				}
			}
			s.logger.Debug("[DEBUG] task_order_log: ===== End of properties =====")

			// Extract project name: Subitem -> Project Deployment -> Project
			projectName := ""
			projectID := ""
			deploymentID := s.extractFirstRelationID(props, "Project Deployment")
			s.logger.Debug(fmt.Sprintf("[DEBUG] task_order_log: extracted deploymentID=%s", deploymentID))
			if deploymentID != "" {
				// Fetch Deployment page to get Project relation
				deploymentPage, err := s.client.FindPageByID(ctx, deploymentID)
				if err != nil {
					s.logger.Debug(fmt.Sprintf("[DEBUG] task_order_log: failed to fetch deployment page %s: %v", deploymentID, err))
				} else {
					deploymentProps, ok := deploymentPage.Properties.(nt.DatabasePageProperties)
					if ok {
						// Get Project relation from Deployment page
						projectID = s.extractFirstRelationID(deploymentProps, "Project")
						s.logger.Debug(fmt.Sprintf("[DEBUG] task_order_log: extracted projectID from Deployment.Project=%s", projectID))
						if projectID != "" {
							// Fetch Project page to get name
							projectPage, err := s.client.FindPageByID(ctx, projectID)
							if err != nil {
								s.logger.Debug(fmt.Sprintf("[DEBUG] task_order_log: failed to fetch project page %s: %v", projectID, err))
							} else {
								projectProps, ok := projectPage.Properties.(nt.DatabasePageProperties)
								if ok {
									// Try Project column first, then Name
									if prop, ok := projectProps["Project"]; ok && len(prop.Title) > 0 {
										projectName = prop.Title[0].PlainText
										s.logger.Debug(fmt.Sprintf("[DEBUG] task_order_log: extracted projectName from Project: %s", projectName))
									} else if prop, ok := projectProps["Name"]; ok && len(prop.Title) > 0 {
										projectName = prop.Title[0].PlainText
										s.logger.Debug(fmt.Sprintf("[DEBUG] task_order_log: extracted projectName from Name: %s", projectName))
									}
								}
							}
						}
					}
				}
			}
			s.logger.Debug(fmt.Sprintf("[DEBUG] task_order_log: final projectName=%s projectID=%s", projectName, projectID))

			subitem := &OrderSubitem{
				PageID:      page.ID,
				ProjectName: projectName,
				ProjectID:   projectID,
				Hours:       s.extractNumber(props, "Line Item Hours"),
				ProofOfWork: s.extractRichText(props, "Proof of Works"),
			}

			s.logger.Debug(fmt.Sprintf("found subitem: pageID=%s project=%s projectID=%s hours=%.2f", subitem.PageID, subitem.ProjectName, subitem.ProjectID, subitem.Hours))

			subitems = append(subitems, subitem)
		}

		if !resp.HasMore || resp.NextCursor == nil {
			break
		}

		query.StartCursor = *resp.NextCursor
	}

	s.logger.Debug(fmt.Sprintf("total subitems found: %d for order: %s", len(subitems), orderPageID))
	return subitems, nil
}

// ApprovedOrderData represents an approved Task Order Log entry (Type=Order, Status=Approved)
type ApprovedOrderData struct {
	PageID            string    // Task Order Log page ID
	ContractorPageID  string    // From Contractor rollup (property ID: q?kW)
	ContractorName    string    // From contractor page Full Name
	ContractorDiscord string    // From contractor page Discord property
	Date              time.Time // From Date property (property ID: Ri:O)
	FinalHoursWorked  float64   // From Final Hours Worked formula (property ID: ;J>Y)
	ProofOfWorks      string    // From Key deliverables rich text (property ID: hlty)
}

// DeploymentData represents an active deployment from Deployment Tracker
type DeploymentData struct {
	PageID           string   // Deployment page ID
	ContractorPageID string   // From Contractor relation
	ProjectPageID    string   // From Project relation
	Status           string   // Deployment status
	Type             []string // Deployment types from Type multi-select (Official, Part-time, Shadow, Not started)
}

// ClientInfo represents client information from Project relation
type ClientInfo struct {
	Name    string // Client name
	Country string // Client headquarters/country
}

// QueryApprovedOrders queries all Task Order Log entries with Type=Order and Status=Approved
// If month is provided (format: YYYY-MM), filters by that month
// Returns approved orders ready to be processed for contractor fee creation
func (s *TaskOrderLogService) QueryApprovedOrders(ctx context.Context, month string) ([]*ApprovedOrderData, error) {
	taskOrderLogDBID := s.cfg.Notion.Databases.TaskOrderLog
	if taskOrderLogDBID == "" {
		return nil, errors.New("task order log database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("querying approved orders: Type=Order, Status=Approved, month=%s", month))

	// Build filter: Type=Order AND Status=Approved
	filters := []nt.DatabaseQueryFilter{
		{
			Property: "Type",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Select: &nt.SelectDatabaseQueryFilter{
					Equals: "Order",
				},
			},
		},
		{
			Property: "Status",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Select: &nt.SelectDatabaseQueryFilter{
					Equals: "Approved",
				},
			},
		},
	}

	// Add month filter if provided
	if month != "" {
		s.logger.Debug(fmt.Sprintf("adding month filter: %s", month))
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

	var approvedOrders []*ApprovedOrderData

	// Query with pagination
	for {
		resp, err := s.client.QueryDatabase(ctx, taskOrderLogDBID, query)
		if err != nil {
			s.logger.Error(err, "failed to query approved orders")
			return nil, fmt.Errorf("failed to query approved orders: %w", err)
		}

		s.logger.Debug(fmt.Sprintf("found %d approved orders in current page", len(resp.Results)))

		for _, page := range resp.Results {
			props, ok := page.Properties.(nt.DatabasePageProperties)
			if !ok {
				s.logger.Debug(fmt.Sprintf("failed to cast page properties for page %s", page.ID))
				continue
			}

			// Debug: Log all properties for inspection
			s.logger.Debug(fmt.Sprintf("processing approved order page: %s", page.ID))

			// Extract contractor from Contractor rollup (property ID: q?kW)
			contractorPageID := ""
			contractorName := ""
			if rollup, ok := props["Contractor"]; ok && rollup.Rollup != nil {
				s.logger.Debug(fmt.Sprintf("contractor rollup array length: %d", len(rollup.Rollup.Array)))
				if len(rollup.Rollup.Array) > 0 {
					// Rollup contains relation items
					for _, item := range rollup.Rollup.Array {
						if len(item.Relation) > 0 {
							contractorPageID = item.Relation[0].ID
							s.logger.Debug(fmt.Sprintf("extracted contractor page ID from rollup: %s", contractorPageID))
							break
						}
					}
				}
			}

			// If contractor is empty, try to get from Sub-item (for Order type)
			if contractorPageID == "" {
				s.logger.Debug("contractor rollup empty, trying to get from Sub-item relation")
				if subitemRel, ok := props["Sub-item"]; ok && len(subitemRel.Relation) > 0 {
					firstSubitemID := subitemRel.Relation[0].ID
					s.logger.Debug(fmt.Sprintf("found first sub-item: %s", firstSubitemID))

					// Fetch the sub-item page to get its Contractor rollup
					subitemPage, err := s.client.FindPageByID(ctx, firstSubitemID)
					if err != nil {
						s.logger.Debug(fmt.Sprintf("failed to fetch sub-item page %s: %v", firstSubitemID, err))
					} else {
						subitemProps, ok := subitemPage.Properties.(nt.DatabasePageProperties)
						if ok {
							// Get Contractor from sub-item's rollup
							if subRollup, ok := subitemProps["Contractor"]; ok && subRollup.Rollup != nil {
								s.logger.Debug(fmt.Sprintf("sub-item contractor rollup array length: %d", len(subRollup.Rollup.Array)))
								for _, item := range subRollup.Rollup.Array {
									if len(item.Relation) > 0 {
										contractorPageID = item.Relation[0].ID
										s.logger.Debug(fmt.Sprintf("extracted contractor page ID from sub-item rollup: %s", contractorPageID))
										break
									}
								}
							}
						}
					}
				}
			}

			// Extract discord from Name property (format: "ORD :: discord :: YYYYMM")
			contractorDiscord := ""
			if titleProp, ok := props["Name"]; ok && len(titleProp.Title) > 0 {
				fullName := ""
				for _, t := range titleProp.Title {
					fullName += t.PlainText
				}
				s.logger.Debug(fmt.Sprintf("order name: %s", fullName))
				// Parse "ORD :: discord :: 202512" format
				parts := strings.Split(fullName, " :: ")
				if len(parts) >= 2 {
					contractorDiscord = strings.TrimSpace(parts[1])
					s.logger.Debug(fmt.Sprintf("extracted discord from name: %s", contractorDiscord))
				}
			}

			// Fetch contractor name if we have the page ID
			if contractorPageID != "" {
				contractorName, _ = s.GetContractorInfo(ctx, contractorPageID)
				s.logger.Debug(fmt.Sprintf("fetched contractor name: %s", contractorName))
			}

			// Extract date
			var orderDate time.Time
			if dateProp, ok := props["Date"]; ok && dateProp.Date != nil {
				orderDate = dateProp.Date.Start.Time
				s.logger.Debug(fmt.Sprintf("extracted date: %s", orderDate.Format("2006-01-02")))
			}

			// Extract Final Hours Worked (formula)
			finalHoursWorked := 0.0
			if prop, ok := props["Final Hours Worked"]; ok && prop.Formula != nil && prop.Formula.Number != nil {
				finalHoursWorked = *prop.Formula.Number
				s.logger.Debug(fmt.Sprintf("extracted final hours worked: %.2f", finalHoursWorked))
			}

			// Extract Key deliverables (rich text)
			proofOfWorks := s.extractRichText(props, "Key deliverables")
			s.logger.Debug(fmt.Sprintf("extracted proof of works length: %d", len(proofOfWorks)))

			order := &ApprovedOrderData{
				PageID:            page.ID,
				ContractorPageID:  contractorPageID,
				ContractorName:    contractorName,
				ContractorDiscord: contractorDiscord,
				Date:              orderDate,
				FinalHoursWorked:  finalHoursWorked,
				ProofOfWorks:      proofOfWorks,
			}

			s.logger.Debug(fmt.Sprintf("parsed approved order: pageID=%s contractor=%s date=%s hours=%.2f",
				order.PageID, order.ContractorName, order.Date.Format("2006-01-02"), order.FinalHoursWorked))

			approvedOrders = append(approvedOrders, order)
		}

		if !resp.HasMore || resp.NextCursor == nil {
			break
		}

		query.StartCursor = *resp.NextCursor
	}

	s.logger.Debug(fmt.Sprintf("total approved orders found: %d", len(approvedOrders)))
	return approvedOrders, nil
}

// UpdateOrderStatus updates the status field of a Task Order Log entry
func (s *TaskOrderLogService) UpdateOrderStatus(ctx context.Context, orderPageID, newStatus string) error {
	s.logger.Debug(fmt.Sprintf("updating order status: pageID=%s newStatus=%s", orderPageID, newStatus))

	updateParams := nt.UpdatePageParams{
		DatabasePageProperties: nt.DatabasePageProperties{
			"Status": nt.DatabasePageProperty{
				Type: nt.DBPropTypeSelect,
				Select: &nt.SelectOptions{
					Name: newStatus,
				},
			},
		},
	}

	_, err := s.client.UpdatePage(ctx, orderPageID, updateParams)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to update order status: pageID=%s", orderPageID))
		return fmt.Errorf("failed to update order status: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("successfully updated order %s status to %s", orderPageID, newStatus))
	return nil
}

// UpdateOrderAndSubitemsStatus updates the status of an order and all its subitems
func (s *TaskOrderLogService) UpdateOrderAndSubitemsStatus(ctx context.Context, orderPageID, newStatus string) error {
	s.logger.Debug(fmt.Sprintf("updating order and subitems status: orderPageID=%s newStatus=%s", orderPageID, newStatus))

	// Update the order itself
	err := s.UpdateOrderStatus(ctx, orderPageID, newStatus)
	if err != nil {
		return err
	}

	// Query and update all subitems
	subitems, err := s.QueryOrderSubitems(ctx, orderPageID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to query subitems for order %s", orderPageID))
		return fmt.Errorf("failed to query subitems: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("updating %d subitems to status %s", len(subitems), newStatus))

	for _, subitem := range subitems {
		s.logger.Debug(fmt.Sprintf("updating subitem status: pageID=%s newStatus=%s", subitem.PageID, newStatus))
		err := s.UpdateOrderStatus(ctx, subitem.PageID, newStatus)
		if err != nil {
			s.logger.Error(err, fmt.Sprintf("failed to update subitem status: %s", subitem.PageID))
			// Continue with other subitems even if one fails
		}
	}

	s.logger.Debug(fmt.Sprintf("successfully updated order %s and %d subitems to status %s", orderPageID, len(subitems), newStatus))
	return nil
}

// GetContractorInfo fetches both Full Name and Discord from a Contractor page
func (s *TaskOrderLogService) GetContractorInfo(ctx context.Context, pageID string) (name string, discord string) {
	page, err := s.client.FindPageByID(ctx, pageID)
	if err != nil {
		s.logger.Debug(fmt.Sprintf("GetContractorInfo: failed to fetch contractor page %s: %v", pageID, err))
		return "", ""
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Debug(fmt.Sprintf("GetContractorInfo: failed to cast page properties for %s", pageID))
		return "", ""
	}

	// Get Full Name from Title property
	if prop, ok := props["Full Name"]; ok && len(prop.Title) > 0 {
		name = prop.Title[0].PlainText
		s.logger.Debug(fmt.Sprintf("GetContractorInfo: found Full Name: %s", name))
	} else if prop, ok := props["Name"]; ok && len(prop.Title) > 0 {
		// Fallback to Name property
		name = prop.Title[0].PlainText
		s.logger.Debug(fmt.Sprintf("GetContractorInfo: found Name: %s", name))
	}

	// Get Discord from rich text property
	if prop, ok := props["Discord"]; ok && len(prop.RichText) > 0 {
		discord = prop.RichText[0].PlainText
		s.logger.Debug(fmt.Sprintf("GetContractorInfo: found Discord: %s", discord))
	}

	return name, discord
}

// FormatProofOfWorksByProject formats subitems grouped by project name with bold project headers
// Format:
// <b>Project Name 1:</b>
// ProofOfWork1
// ProofOfWork2
//
// <b>Project Name 2:</b>
// ProofOfWork3
func (s *TaskOrderLogService) FormatProofOfWorksByProject(ctx context.Context, orderPageIDs []string) (string, error) {
	s.logger.Debug(fmt.Sprintf("[DEBUG] task_order_log: formatting proof of works for %d orders", len(orderPageIDs)))

	// Collect all subitems from all orders
	var allSubitems []*OrderSubitem
	for _, orderID := range orderPageIDs {
		subitems, err := s.QueryOrderSubitems(ctx, orderID)
		if err != nil {
			s.logger.Error(err, fmt.Sprintf("[DEBUG] task_order_log: failed to query subitems for order=%s", orderID))
			continue
		}
		allSubitems = append(allSubitems, subitems...)
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] task_order_log: collected %d total subitems", len(allSubitems)))

	if len(allSubitems) == 0 {
		return "", nil
	}

	// Group by project name
	projectMap := make(map[string][]string) // projectName -> []proofOfWorks
	projectOrder := []string{}              // maintain order of first appearance

	for _, subitem := range allSubitems {
		projectName := subitem.ProjectName
		if projectName == "" {
			projectName = "Other"
		}

		// Track order of first appearance
		if _, exists := projectMap[projectName]; !exists {
			projectOrder = append(projectOrder, projectName)
		}

		if subitem.ProofOfWork != "" {
			projectMap[projectName] = append(projectMap[projectName], subitem.ProofOfWork)
		}
	}

	s.logger.Debug(fmt.Sprintf("[DEBUG] task_order_log: grouped into %d projects", len(projectOrder)))

	// Build formatted string
	var result strings.Builder
	for i, projectName := range projectOrder {
		if i > 0 {
			result.WriteString("\n\n")
		}
		result.WriteString(fmt.Sprintf("<b>%s</b>\n", projectName))

		proofs := projectMap[projectName]
		for j, proof := range proofs {
			if j > 0 {
				result.WriteString("\n")
			}
			result.WriteString(proof)
		}

		s.logger.Debug(fmt.Sprintf("[DEBUG] task_order_log: project=%s proofs=%d", projectName, len(proofs)))
	}

	formatted := result.String()
	s.logger.Debug(fmt.Sprintf("[DEBUG] task_order_log: formatted proof of works length=%d", len(formatted)))

	return formatted, nil
}

// QueryActiveDeploymentsByMonth queries active deployments from Deployment Tracker
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - month: Target month in YYYY-MM format (used for logging)
//   - contractorDiscord: Optional Discord username filter (empty string = all contractors)
//
// Returns:
//   - []*DeploymentData: Slice of active deployments
//   - error: Error if query fails
func (s *TaskOrderLogService) QueryActiveDeploymentsByMonth(ctx context.Context, month string, contractorDiscord string) ([]*DeploymentData, error) {
	deploymentDBID := s.cfg.Notion.Databases.DeploymentTracker
	if deploymentDBID == "" {
		return nil, errors.New("deployment tracker database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("querying active deployments: month=%s discord=%s", month, contractorDiscord))

	// Build filters
	filters := []nt.DatabaseQueryFilter{
		{
			Property: "Deployment Status",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Status: &nt.StatusDatabaseQueryFilter{
					Equals: "Active",
				},
			},
		},
	}

	// Add Discord filter if provided
	if contractorDiscord != "" {
		filters = append(filters, nt.DatabaseQueryFilter{
			Property: "Discord",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Rollup: &nt.RollupDatabaseQueryFilter{
					Any: &nt.DatabaseQueryPropertyFilter{
						RichText: &nt.TextPropertyFilter{
							Contains: contractorDiscord,
						},
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

	var deployments []*DeploymentData

	// Query with pagination
	for {
		resp, err := s.client.QueryDatabase(ctx, deploymentDBID, query)
		if err != nil {
			s.logger.Error(err, fmt.Sprintf("failed to query deployment tracker: month=%s", month))
			return nil, fmt.Errorf("failed to query deployment tracker: %w", err)
		}

		for _, page := range resp.Results {
			props, ok := page.Properties.(nt.DatabasePageProperties)
			if !ok {
				continue
			}

			deployment := &DeploymentData{
				PageID:           page.ID,
				ContractorPageID: s.extractFirstRelationID(props, "Contractor"),
				ProjectPageID:    s.extractFirstRelationID(props, "Project"),
				Status:           s.extractStatus(props, "Deployment Status"),
				Type:             s.extractMultiSelectNames(props, "Type"),
			}

			deployments = append(deployments, deployment)
		}

		if !resp.HasMore || resp.NextCursor == nil {
			break
		}

		query.StartCursor = *resp.NextCursor
	}

	s.logger.Debug(fmt.Sprintf("found %d active deployments for month %s", len(deployments), month))
	return deployments, nil
}

// GetClientInfo fetches client information from a project page
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - projectPageID: Project page ID
//
// Returns:
//   - *ClientInfo: Client name and country, nil if not found
//   - error: Error if fetch fails
func (s *TaskOrderLogService) GetClientInfo(ctx context.Context, projectPageID string) (*ClientInfo, error) {
	// Fetch project page
	projectPage, err := s.client.FindPageByID(ctx, projectPageID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to fetch project page: %s", projectPageID))
		return nil, fmt.Errorf("failed to fetch project page: %w", err)
	}

	props, ok := projectPage.Properties.(nt.DatabasePageProperties)
	if !ok {
		return nil, errors.New("failed to cast project page properties")
	}

	// Extract client relation
	clientPageID := s.extractFirstRelationID(props, "Client")
	if clientPageID == "" {
		// Fallback to "Clients"
		clientPageID = s.extractFirstRelationID(props, "Clients")
	}

	if clientPageID == "" {
		s.logger.Debug(fmt.Sprintf("no client relation found for project: %s", projectPageID))
		return nil, nil
	}

	// Fetch client page
	clientPage, err := s.client.FindPageByID(ctx, clientPageID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to fetch client page: %s", clientPageID))
		return nil, fmt.Errorf("failed to fetch client page: %w", err)
	}

	clientProps, ok := clientPage.Properties.(nt.DatabasePageProperties)
	if !ok {
		return nil, errors.New("failed to cast client page properties")
	}

	// Extract client name and country
	name := s.extractTitle(clientProps, "Client Name")
	country := s.extractRichText(clientProps, "Country")

	if name == "" {
		s.logger.Debug(fmt.Sprintf("client name not found for client page: %s", clientPageID))
		return nil, nil
	}

	return &ClientInfo{
		Name:    name,
		Country: country,
	}, nil
}

// GetContractorTeamEmail fetches team email from contractor page
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - contractorPageID: Contractor page ID
//
// Returns:
//   - string: Team email address, empty if not found
func (s *TaskOrderLogService) GetContractorTeamEmail(ctx context.Context, contractorPageID string) string {
	page, err := s.client.FindPageByID(ctx, contractorPageID)
	if err != nil {
		s.logger.Debug(fmt.Sprintf("failed to fetch contractor page: %s: %v", contractorPageID, err))
		return ""
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Debug(fmt.Sprintf("failed to cast contractor page properties: %s", contractorPageID))
		return ""
	}

	// Extract Team Email property (email type)
	if prop, ok := props["Team Email"]; ok && prop.Email != nil {
		return *prop.Email
	}

	s.logger.Debug(fmt.Sprintf("team email not found for contractor: %s", contractorPageID))
	return ""
}

// GetContractorPersonalEmail fetches the Personal Email from a Contractor page
func (s *TaskOrderLogService) GetContractorPersonalEmail(ctx context.Context, contractorPageID string) string {
	page, err := s.client.FindPageByID(ctx, contractorPageID)
	if err != nil {
		s.logger.Debug(fmt.Sprintf("failed to fetch contractor page: %s: %v", contractorPageID, err))
		return ""
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Debug(fmt.Sprintf("failed to cast contractor page properties: %s", contractorPageID))
		return ""
	}

	// Extract Personal Email property (email type)
	if prop, ok := props["Personal Email"]; ok && prop.Email != nil {
		s.logger.Debug(fmt.Sprintf("found personal email for contractor %s: %s", contractorPageID, *prop.Email))
		return *prop.Email
	}

	s.logger.Debug(fmt.Sprintf("personal email not found for contractor: %s", contractorPageID))
	return ""
}

// GetContractorPayday fetches the Payday value from Contractor Rates database
// Returns:
//   - 1 if Payday = "01"
//   - 15 if Payday = "15"
//   - 0 if not found or invalid (caller should use default)
//   - nil error on success, error object only for system failures
func (s *TaskOrderLogService) GetContractorPayday(ctx context.Context, contractorPageID string) (int, error) {
	// Validate database configuration
	contractorRatesDBID := s.cfg.Notion.Databases.ContractorRates
	if contractorRatesDBID == "" {
		s.logger.Debug("contractor rates database ID not configured")
		return 0, nil // Graceful fallback
	}

	s.logger.Debug(fmt.Sprintf("querying contractor rates database for contractor: %s", contractorPageID))

	// Build query to find Active Service Rate for contractor
	filter := nt.DatabaseQueryFilter{
		Property: "Contractor",
		DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
			Relation: &nt.RelationDatabaseQueryFilter{
				Contains: contractorPageID,
			},
		},
	}

	statusFilter := nt.DatabaseQueryFilter{
		Property: "Status",
		DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
			Status: &nt.StatusDatabaseQueryFilter{
				Equals: "Active",
			},
		},
	}

	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			And: []nt.DatabaseQueryFilter{filter, statusFilter},
		},
		PageSize: 1, // We only need one active record
	}

	// Execute query
	resp, err := s.client.QueryDatabase(ctx, contractorRatesDBID, query)
	if err != nil {
		s.logger.Debug(fmt.Sprintf("failed to query contractor rates database for contractor %s: %v", contractorPageID, err))
		return 0, nil // Graceful fallback - don't block email sending
	}

	// Check if any active service rate found
	if len(resp.Results) == 0 {
		s.logger.Debug(fmt.Sprintf("no active contractor rate found for contractor: %s", contractorPageID))
		return 0, nil
	}

	// Extract properties from first result
	props, ok := resp.Results[0].Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Debug(fmt.Sprintf("failed to cast contractor rate properties for contractor: %s", contractorPageID))
		return 0, nil
	}

	// Extract Payday field
	paydayStr := s.extractSelect(props, "Payday")
	if paydayStr == "" {
		s.logger.Debug(fmt.Sprintf("payday field is empty for contractor: %s", contractorPageID))
		return 0, nil
	}

	// Parse and validate Payday value
	switch paydayStr {
	case "01":
		s.logger.Debug(fmt.Sprintf("contractor payday found: contractor=%s payday=1", contractorPageID))
		return 1, nil
	case "15":
		s.logger.Debug(fmt.Sprintf("contractor payday found: contractor=%s payday=15", contractorPageID))
		return 15, nil
	default:
		s.logger.Debug(fmt.Sprintf("invalid payday value for contractor %s: %s", contractorPageID, paydayStr))
		return 0, nil
	}
}

// AppendBlocksToPage appends text content as paragraph blocks to a Notion page
func (s *TaskOrderLogService) AppendBlocksToPage(ctx context.Context, pageID string, content string) error {
	s.logger.Debug(fmt.Sprintf("appending blocks to page: pageID=%s contentLength=%d", pageID, len(content)))

	if content == "" {
		s.logger.Debug("empty content, skipping append")
		return nil
	}

	// Split content by newlines and create paragraph blocks
	lines := strings.Split(content, "\n")
	var blocks []nt.Block

	for _, line := range lines {
		// Skip empty lines but add empty paragraph for spacing
		block := nt.ParagraphBlock{
			RichText: []nt.RichText{
				{
					Type: nt.RichTextTypeText,
					Text: &nt.Text{
						Content: line,
					},
				},
			},
		}
		blocks = append(blocks, block)
	}

	s.logger.Debug(fmt.Sprintf("appending %d paragraph blocks to page: %s", len(blocks), pageID))

	_, err := s.client.AppendBlockChildren(ctx, pageID, blocks)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to append blocks to page: %s", pageID))
		return fmt.Errorf("failed to append blocks to page: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("successfully appended blocks to page: %s", pageID))
	return nil
}

// GenerateConfirmationContent generates plain text confirmation content for a contractor
func (s *TaskOrderLogService) GenerateConfirmationContent(contractorName, month string, clients []model.TaskOrderClient) string {
	s.logger.Debug(fmt.Sprintf("generating confirmation content: contractor=%s month=%s clients=%d", contractorName, month, len(clients)))

	// Parse month to get formatted values
	t, err := time.Parse("2006-01", month)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to parse month: %s", month))
		t = time.Now()
	}

	formattedMonth := t.Format("January 2006")
	lastDay := time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, time.UTC)
	periodEndDay := fmt.Sprintf("%02d", lastDay.Day())
	monthName := t.Format("January")
	year := t.Format("2006")

	// Get contractor last name for greeting
	parts := strings.Fields(contractorName)
	contractorLastName := contractorName
	if len(parts) > 0 {
		contractorLastName = parts[len(parts)-1]
	}

	// Build clients list
	var clientsText string
	for i, client := range clients {
		clientStr := client.Name
		if client.Country != "" {
			clientStr = fmt.Sprintf("%s – headquartered in %s", client.Name, client.Country)
		}
		if i > 0 {
			clientsText += "\n"
		}
		clientsText += "- " + clientStr
	}

	// Build content
	content := fmt.Sprintf(`Hi %s,

This email outlines your planned assignments and work order for: %s.

Period: 01 – %s %s, %s

Active clients & locations:
%s

All tasks and deliverables will be tracked in Notion/Jira as usual.

Please reply "Confirmed – %s" to acknowledge this work order and confirm your availability.

Thanks,

Dwarves LLC`,
		contractorLastName,
		formattedMonth,
		periodEndDay, monthName, year,
		clientsText,
		formattedMonth,
	)

	s.logger.Debug(fmt.Sprintf("generated confirmation content for %s: %d chars", contractorName, len(content)))
	return content
}

// GenerateConfirmationHTML generates HTML confirmation content matching template format
func (s *TaskOrderLogService) GenerateConfirmationHTML(contractorName, month string, clients []model.TaskOrderClient) string {
	s.logger.Debug(fmt.Sprintf("generating HTML confirmation: contractor=%s month=%s clients=%d", contractorName, month, len(clients)))

	// Parse month to get formatted values
	t, err := time.Parse("2006-01", month)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to parse month: %s", month))
		t = time.Now()
	}

	formattedMonth := t.Format("January 2006")

	// Get contractor last name for greeting
	parts := strings.Fields(contractorName)
	contractorLastName := contractorName
	if len(parts) > 0 {
		contractorLastName = parts[len(parts)-1]
	}

	// Default invoice due day (can be overridden if payday info is available)
	invoiceDueDay := "10th"

	// Build milestones list from clients as HTML
	var milestonesHTML string
	for _, client := range clients {
		milestone := fmt.Sprintf("Continuing work with %s", client.Name)
		if client.Country != "" {
			milestone = fmt.Sprintf("Continuing work with %s (based in %s)", client.Name, client.Country)
		}
		milestonesHTML += fmt.Sprintf("        <li>%s</li>\n", milestone)
	}

	// Load signature from template file
	signature := s.loadTaskOrderSignature()

	// Build HTML content matching new template format
	html := fmt.Sprintf(`<div>
    <p>Hi %s,</p>

    <p>Hope you're having a great start to %s!</p>

    <p>Just a quick note:</p>

    <p>Your regular monthly invoice for %s services is due by <b>%s</b>. As usual, please use the standard template and send to <a href="mailto:invoices@d.foundation">invoices@d.foundation</a>.</p>

    <p>Upcoming client milestones (for awareness):</p>
    <ul>
%s    </ul>

    <p>You're continuing to do excellent work on the embedded team – clients are very happy with your contributions.</p>

    <p>If anything comes up or you need support, just ping me anytime.</p>

    <p>Best,</p>

    <div><br></div>-- <br>
%s
</div>`,
		contractorLastName,
		formattedMonth,
		formattedMonth,
		invoiceDueDay,
		milestonesHTML,
		signature,
	)

	s.logger.Debug(fmt.Sprintf("generated HTML confirmation for %s: %d chars", contractorName, len(html)))
	return html
}

// loadTaskOrderSignature loads and renders the signature from template file
func (s *TaskOrderLogService) loadTaskOrderSignature() string {
	templatePath := s.cfg.Invoice.TemplatePath
	if s.cfg.Env == "local" || templatePath == "" {
		pwd, err := os.Getwd()
		if err != nil {
			pwd = os.Getenv("GOPATH") + "/src/github.com/dwarvesf/fortress-api"
		}
		templatePath = filepath.Join(pwd, "pkg/templates")
	}

	signaturePath := filepath.Join(templatePath, "signature.tpl")
	s.logger.Debug(fmt.Sprintf("loading task order signature from: %s", signaturePath))

	// Parse and execute template with signatureName, signatureTitle, and signatureNameSuffix
	tmpl, err := template.New("signature.tpl").Funcs(template.FuncMap{
		"signatureName": func() string {
			return "Team Dwarves"
		},
		"signatureTitle": func() string {
			return "People Operations"
		},
		"signatureNameSuffix": func() string {
			return "" // No dot for task order emails
		},
	}).ParseFiles(signaturePath)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to parse signature template: %s", signaturePath))
		return ""
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to execute signature template: %s", signaturePath))
		return ""
	}

	s.logger.Debug(fmt.Sprintf("loaded task order signature: %d bytes", buf.Len()))
	return buf.String()
}

// AppendCodeBlockToPage appends HTML content as a code block to a Notion page
func (s *TaskOrderLogService) AppendCodeBlockToPage(ctx context.Context, pageID string, content string) error {
	s.logger.Debug(fmt.Sprintf("appending code block to page: pageID=%s contentLength=%d", pageID, len(content)))

	if content == "" {
		s.logger.Debug("empty content, skipping append")
		return nil
	}

	// Split content into chunks of 2000 characters (Notion's limit for rich_text content)
	const maxChunkSize = 2000
	var richTexts []nt.RichText

	for i := 0; i < len(content); i += maxChunkSize {
		end := i + maxChunkSize
		if end > len(content) {
			end = len(content)
		}
		chunk := content[i:end]
		richTexts = append(richTexts, nt.RichText{
			Type: nt.RichTextTypeText,
			Text: &nt.Text{
				Content: chunk,
			},
		})
	}

	s.logger.Debug(fmt.Sprintf("split content into %d chunks", len(richTexts)))

	// Create a code block with HTML language
	lang := "html"
	codeBlock := nt.CodeBlock{
		RichText: richTexts,
		Language: &lang,
	}

	blocks := []nt.Block{codeBlock}

	s.logger.Debug(fmt.Sprintf("appending code block to page: %s", pageID))

	_, err := s.client.AppendBlockChildren(ctx, pageID, blocks)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to append code block to page: %s", pageID))
		return fmt.Errorf("failed to append code block to page: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("successfully appended code block to page: %s", pageID))
	return nil
}

// GetOrderPageContent reads page body content from Order page
// Returns concatenated text from paragraph blocks
func (s *TaskOrderLogService) GetOrderPageContent(ctx context.Context, pageID string) (string, error) {
	s.logger.Debug(fmt.Sprintf("getting order page content: pageID=%s", pageID))

	// Fetch block children from page
	resp, err := s.client.FindBlockChildrenByID(ctx, pageID, nil)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to fetch block children: pageID=%s", pageID))
		return "", fmt.Errorf("failed to fetch block children: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("found %d blocks in page: %s", len(resp.Results), pageID))

	// First, check for Code blocks (HTML content)
	for _, block := range resp.Results {
		if v, ok := block.(*nt.CodeBlock); ok {
			var codeText []string
			for _, text := range v.RichText {
				codeText = append(codeText, text.PlainText)
			}
			content := strings.Join(codeText, "")
			s.logger.Debug(fmt.Sprintf("found code block with HTML content: %d chars", len(content)))
			return content, nil
		}
	}

	// Fallback: extract text from paragraph and list blocks
	var lines []string
	for _, block := range resp.Results {
		switch v := block.(type) {
		case *nt.ParagraphBlock:
			var lineText []string
			for _, text := range v.RichText {
				lineText = append(lineText, text.PlainText)
			}
			lines = append(lines, strings.Join(lineText, ""))
		case *nt.BulletedListItemBlock:
			var lineText []string
			for _, text := range v.RichText {
				lineText = append(lineText, text.PlainText)
			}
			lines = append(lines, "- "+strings.Join(lineText, ""))
		case *nt.NumberedListItemBlock:
			var lineText []string
			for _, text := range v.RichText {
				lineText = append(lineText, text.PlainText)
			}
			lines = append(lines, "- "+strings.Join(lineText, ""))
		}
	}

	content := strings.Join(lines, "\n")
	s.logger.Debug(fmt.Sprintf("extracted content from page %s: %d lines, %d chars", pageID, len(lines), len(content)))

	return content, nil
}

// GetContractorFromOrder gets contractor info from Order via Sub-items → Deployment → Contractor chain
// Returns contractor page ID, team email, and full name
func (s *TaskOrderLogService) GetContractorFromOrder(ctx context.Context, orderID string) (contractorID, email, name string, err error) {
	s.logger.Debug(fmt.Sprintf("getting contractor from order: orderID=%s", orderID))

	// Step 1: Query subitems (Line Items) for this order
	subitems, err := s.QueryOrderSubitems(ctx, orderID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to query subitems: orderID=%s", orderID))
		return "", "", "", fmt.Errorf("failed to query subitems: %w", err)
	}

	if len(subitems) == 0 {
		s.logger.Debug(fmt.Sprintf("no subitems found for order: %s", orderID))
		return "", "", "", fmt.Errorf("no subitems found for order: %s", orderID)
	}

	s.logger.Debug(fmt.Sprintf("found %d subitems for order %s", len(subitems), orderID))

	// Step 2: Get Deployment from first Line Item
	firstSubitem := subitems[0]
	subitemPage, err := s.client.FindPageByID(ctx, firstSubitem.PageID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to fetch subitem page: %s", firstSubitem.PageID))
		return "", "", "", fmt.Errorf("failed to fetch subitem page: %w", err)
	}

	subitemProps, ok := subitemPage.Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Debug(fmt.Sprintf("failed to cast subitem page properties: %s", firstSubitem.PageID))
		return "", "", "", fmt.Errorf("failed to cast subitem page properties")
	}

	deploymentID := s.extractFirstRelationID(subitemProps, "Project Deployment")
	if deploymentID == "" {
		s.logger.Debug(fmt.Sprintf("no deployment found for subitem: %s", firstSubitem.PageID))
		return "", "", "", fmt.Errorf("no deployment found for subitem: %s", firstSubitem.PageID)
	}

	s.logger.Debug(fmt.Sprintf("found deployment %s from subitem %s", deploymentID, firstSubitem.PageID))

	// Step 3: Get Contractor from Deployment
	deploymentPage, err := s.client.FindPageByID(ctx, deploymentID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to fetch deployment page: %s", deploymentID))
		return "", "", "", fmt.Errorf("failed to fetch deployment page: %w", err)
	}

	deploymentProps, ok := deploymentPage.Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Debug(fmt.Sprintf("failed to cast deployment page properties: %s", deploymentID))
		return "", "", "", fmt.Errorf("failed to cast deployment page properties")
	}

	contractorID = s.extractFirstRelationID(deploymentProps, "Contractor")
	if contractorID == "" {
		s.logger.Debug(fmt.Sprintf("no contractor found for deployment: %s", deploymentID))
		return "", "", "", fmt.Errorf("no contractor found for deployment: %s", deploymentID)
	}

	s.logger.Debug(fmt.Sprintf("found contractor %s from deployment %s", contractorID, deploymentID))

	// Step 4: Get contractor personal email and name
	email = s.GetContractorPersonalEmail(ctx, contractorID)
	name, _ = s.GetContractorInfo(ctx, contractorID)

	s.logger.Debug(fmt.Sprintf("contractor info: id=%s email=%s name=%s", contractorID, email, name))

	return contractorID, email, name, nil
}

// FetchTaskOrderHoursByPageID fetches the Final Hours Worked from a Task Order Log page.
// Used for hourly rate invoice line item display.
func (s *TaskOrderLogService) FetchTaskOrderHoursByPageID(ctx context.Context, pageID string) (float64, error) {
	s.logger.Debug(fmt.Sprintf("[HOURLY_RATE] fetching task order hours: pageID=%s", pageID))

	// Step 1: Fetch the page by ID using Notion client
	page, err := s.client.FindPageByID(ctx, pageID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to fetch task order page: %s", pageID))
		return 0, fmt.Errorf("failed to fetch task order page: %w", err)
	}

	// Step 2: Cast page properties to database properties
	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		return 0, fmt.Errorf("failed to cast page properties for task order: %s", pageID)
	}

	// Step 3: Extract Final Hours Worked (formula field)
	// Returns 0.0 if field not found or empty (graceful degradation)
	var hours float64
	if prop, ok := props["Final Hours Worked"]; ok && prop.Formula != nil && prop.Formula.Number != nil {
		hours = *prop.Formula.Number
		s.logger.Debug(fmt.Sprintf("[HOURLY_RATE] fetched hours: %.2f", hours))
	} else {
		// Field not found or empty - log but don't error
		// This allows graceful degradation (display with 0 hours)
		s.logger.Debug(fmt.Sprintf("[HOURLY_RATE] Final Hours Worked not found or empty for pageID=%s, returning 0", pageID))
		return 0, nil
	}

	return hours, nil
}
