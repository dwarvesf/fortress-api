package notion

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
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

	// Build filter
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
			Property: "Month",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Formula: &nt.FormulaDatabaseQueryFilter{
					String: &nt.TextPropertyFilter{
						Equals: month,
					},
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
				Title:            s.extractTitle(props, "(auto) Timesheet Entry"),
				ContractorPageID: s.extractFirstRelationID(props, "Contractor"),
				ProjectPageID:    s.extractFirstRelationID(props, "Project"),
				Date:             s.extractDateString(props, "Date"),
				Hours:            s.extractNumber(props, "Hours"),
				Status:           s.extractStatus(props, "Status"),
			}

			// Extract proof of works
			if prop, ok := props["Proof of Works"]; ok && len(prop.RichText) > 0 {
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
					Property: "Deployment",
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
func (s *TaskOrderLogService) CreateOrder(ctx context.Context, deploymentID, month string) (string, error) {
	taskOrderLogDBID := s.cfg.Notion.Databases.TaskOrderLog
	if taskOrderLogDBID == "" {
		return "", errors.New("task order log database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("creating order: deployment=%s month=%s", deploymentID, month))

	// Use current date
	now := time.Now()

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
				Name: "Pending Approval",
			},
		},
		"Date": nt.DatabasePageProperty{
			Type: nt.DBPropTypeDate,
			Date: &nt.Date{
				Start: nt.NewDateTime(now, false),
			},
		},
	}

	// Add Deployment if provided
	if deploymentID != "" {
		(*properties)["Deployment"] = nt.DatabasePageProperty{
			Type: nt.DBPropTypeRelation,
			Relation: []nt.Relation{
				{ID: deploymentID},
			},
		}
	}

	params := nt.CreatePageParams{
		ParentType:             nt.ParentTypeDatabase,
		ParentID:               taskOrderLogDBID,
		DatabasePageProperties: properties,
	}

	page, err := s.client.CreatePage(ctx, params)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to create order: deployment=%s month=%s", deploymentID, month))
		return "", fmt.Errorf("failed to create order: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("created order: %s", page.ID))
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
					Property: "Deployment",
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

	// Use current date
	now := time.Now()

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
					Name: "Pending Approval",
				},
			},
			"Date": nt.DatabasePageProperty{
				Type: nt.DBPropTypeDate,
				Date: &nt.Date{
					Start: nt.NewDateTime(now, false),
				},
			},
			"Line Item Hours": nt.DatabasePageProperty{
				Type:   nt.DBPropTypeNumber,
				Number: &hours,
			},
			"Proof of Works": nt.DatabasePageProperty{
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
			"Deployment": nt.DatabasePageProperty{
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

// getProjectName fetches the project name from a Project page
func (s *TaskOrderLogService) getProjectName(ctx context.Context, pageID string) string {
	page, err := s.client.FindPageByID(ctx, pageID)
	if err != nil {
		fmt.Printf("[DEBUG] getProjectName: failed to fetch project page %s: %v\n", pageID, err)
		return ""
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		fmt.Printf("[DEBUG] getProjectName: failed to cast page properties for %s\n", pageID)
		return ""
	}

	// Try to get Name from Title property
	if prop, ok := props["Name"]; ok && len(prop.Title) > 0 {
		name := prop.Title[0].PlainText
		fmt.Printf("[DEBUG] getProjectName: found Name in Title: %s\n", name)
		return name
	}

	// Try Project Name property as fallback
	if prop, ok := props["Project Name"]; ok && len(prop.Title) > 0 {
		name := prop.Title[0].PlainText
		fmt.Printf("[DEBUG] getProjectName: found Project Name in Title: %s\n", name)
		return name
	}

	fmt.Printf("[DEBUG] getProjectName: no Name or Project Name property found for page %s\n", pageID)
	return ""
}

// OrderSubitem represents a line item (timesheet) in Task Order Log
type OrderSubitem struct {
	PageID      string
	ProjectName string  // From Project rollup
	ProjectID   string  // From Project relation
	Hours       float64 // From Line Item Hours
	ProofOfWork string  // From Proof of Works rich text
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
			fmt.Printf("[DEBUG] task_order_log: ===== Subitem page %s properties =====\n", page.ID)
			for propName, prop := range props {
				fmt.Printf("[DEBUG]   Property: %s\n", propName)
				if len(prop.Relation) > 0 {
					fmt.Printf("[DEBUG]     Type: Relation, Count: %d, First ID: %s\n", len(prop.Relation), prop.Relation[0].ID)
				}
				if prop.Rollup != nil {
					fmt.Printf("[DEBUG]     Type: Rollup, Array Length: %d\n", len(prop.Rollup.Array))
					for i, item := range prop.Rollup.Array {
						fmt.Printf("[DEBUG]       Rollup[%d]: Title=%d, RichText=%d, Relation=%d\n",
							i, len(item.Title), len(item.RichText), len(item.Relation))
						if len(item.Relation) > 0 {
							fmt.Printf("[DEBUG]         Relation[0] ID: %s\n", item.Relation[0].ID)
						}
					}
				}
				if len(prop.Title) > 0 {
					fmt.Printf("[DEBUG]     Type: Title, Value: %s\n", prop.Title[0].PlainText)
				}
				if len(prop.RichText) > 0 {
					fmt.Printf("[DEBUG]     Type: RichText, Value: %s\n", prop.RichText[0].PlainText)
				}
				if prop.Select != nil {
					fmt.Printf("[DEBUG]     Type: Select, Value: %s\n", prop.Select.Name)
				}
				if prop.Number != nil {
					fmt.Printf("[DEBUG]     Type: Number, Value: %f\n", *prop.Number)
				}
			}
			fmt.Printf("[DEBUG] task_order_log: ===== End of properties =====\n")

			// Extract project ID from Deployment relation (not Project rollup)
			projectID := s.extractFirstRelationID(props, "Deployment")
			fmt.Printf("[DEBUG] task_order_log: extracted projectID from Deployment=%s\n", projectID)

			// Fetch project name from Deployment/Project page
			projectName := ""
			if projectID != "" {
				projectName = s.getProjectName(ctx, projectID)
				fmt.Printf("[DEBUG] task_order_log: fetched projectName=%s for projectID=%s\n", projectName, projectID)
			}

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
	PageID           string    // Task Order Log page ID
	ContractorPageID string    // From Contractor rollup (property ID: q?kW)
	ContractorName   string    // From contractor page Full Name
	ContractorDiscord string   // From contractor page Discord property
	Date             time.Time // From Date property (property ID: Ri:O)
	FinalHoursWorked float64   // From Final Hours Worked formula (property ID: ;J>Y)
	ProofOfWorks     string    // From Proof of Works rich text (property ID: hlty)
}

// QueryApprovedOrders queries all Task Order Log entries with Type=Order and Status=Approved
// Returns approved orders ready to be processed for contractor fee creation
func (s *TaskOrderLogService) QueryApprovedOrders(ctx context.Context) ([]*ApprovedOrderData, error) {
	taskOrderLogDBID := s.cfg.Notion.Databases.TaskOrderLog
	if taskOrderLogDBID == "" {
		return nil, errors.New("task order log database ID not configured")
	}

	s.logger.Debug("querying approved orders: Type=Order, Status=Approved")

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
					Property: "Status",
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Select: &nt.SelectDatabaseQueryFilter{
							Equals: "Approved",
						},
					},
				},
			},
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

			// Fetch contractor name and discord if we have the page ID
			contractorDiscord := ""
			if contractorPageID != "" {
				contractorName, contractorDiscord = s.getContractorInfo(ctx, contractorPageID)
				s.logger.Debug(fmt.Sprintf("fetched contractor info: name=%s discord=%s", contractorName, contractorDiscord))
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

			// Extract Proof of Works (rich text)
			proofOfWorks := s.extractRichText(props, "Proof of Works")
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

// getContractorName fetches the Full Name from a Contractor page
func (s *TaskOrderLogService) getContractorName(ctx context.Context, pageID string) string {
	page, err := s.client.FindPageByID(ctx, pageID)
	if err != nil {
		s.logger.Debug(fmt.Sprintf("getContractorName: failed to fetch contractor page %s: %v", pageID, err))
		return ""
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Debug(fmt.Sprintf("getContractorName: failed to cast page properties for %s", pageID))
		return ""
	}

	// Try to get Full Name from Title property
	if prop, ok := props["Full Name"]; ok && len(prop.Title) > 0 {
		name := prop.Title[0].PlainText
		s.logger.Debug(fmt.Sprintf("getContractorName: found Full Name: %s", name))
		return name
	}

	// Try Name property as fallback
	if prop, ok := props["Name"]; ok && len(prop.Title) > 0 {
		name := prop.Title[0].PlainText
		s.logger.Debug(fmt.Sprintf("getContractorName: found Name: %s", name))
		return name
	}

	s.logger.Debug(fmt.Sprintf("getContractorName: no Full Name or Name property found for page %s", pageID))
	return ""
}

// getContractorInfo fetches both Full Name and Discord from a Contractor page
func (s *TaskOrderLogService) getContractorInfo(ctx context.Context, pageID string) (name string, discord string) {
	page, err := s.client.FindPageByID(ctx, pageID)
	if err != nil {
		s.logger.Debug(fmt.Sprintf("getContractorInfo: failed to fetch contractor page %s: %v", pageID, err))
		return "", ""
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Debug(fmt.Sprintf("getContractorInfo: failed to cast page properties for %s", pageID))
		return "", ""
	}

	// Get Full Name from Title property
	if prop, ok := props["Full Name"]; ok && len(prop.Title) > 0 {
		name = prop.Title[0].PlainText
		s.logger.Debug(fmt.Sprintf("getContractorInfo: found Full Name: %s", name))
	} else if prop, ok := props["Name"]; ok && len(prop.Title) > 0 {
		// Fallback to Name property
		name = prop.Title[0].PlainText
		s.logger.Debug(fmt.Sprintf("getContractorInfo: found Name: %s", name))
	}

	// Get Discord from rich text property
	if prop, ok := props["Discord"]; ok && len(prop.RichText) > 0 {
		discord = prop.RichText[0].PlainText
		s.logger.Debug(fmt.Sprintf("getContractorInfo: found Discord: %s", discord))
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
		result.WriteString(fmt.Sprintf("<b>%s:</b>\n", projectName))

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
