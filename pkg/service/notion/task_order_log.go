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
func (s *TaskOrderLogService) QueryApprovedTimesheetsByMonth(ctx context.Context, month string, contractorDiscord string) ([]*TimesheetEntry, error) {
	timesheetDBID := s.cfg.Notion.Databases.Timesheet
	if timesheetDBID == "" {
		return nil, errors.New("timesheet database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("querying approved timesheets: month=%s contractor=%s", month, contractorDiscord))

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
				Name: "Draft",
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
					Name: "Approved",
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
