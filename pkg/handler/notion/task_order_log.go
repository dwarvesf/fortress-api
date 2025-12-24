package notion

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

// SyncTaskOrderLogs godoc
// @Summary Sync task order logs from approved timesheets
// @Description Creates Task Order Log entries from approved Timesheet entries for a given month
// @Tags Cronjobs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param month query string true "Target month in YYYY-MM format (e.g., 2025-12)"
// @Param contractor query string false "Discord username to filter by specific contractor (e.g., chinhld)"
// @Success 200 {object} view.Response
// @Failure 400 {object} view.Response
// @Failure 500 {object} view.Response
// @Router /cronjobs/sync-task-order-logs [post]
func (h *handler) SyncTaskOrderLogs(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "Notion",
		"method":  "SyncTaskOrderLogs",
	})
	ctx := c.Request.Context()

	// Parse month parameter (required)
	month := c.Query("month")
	if month == "" {
		l.Error(fmt.Errorf("month parameter required"), "month query param is missing")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, fmt.Errorf("month parameter is required"), nil, ""))
		return
	}

	// Validate month format (YYYY-MM)
	if !isValidMonthFormat(month) {
		l.Error(fmt.Errorf("invalid month format"), fmt.Sprintf("month=%s", month))
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, fmt.Errorf("invalid month format, expected YYYY-MM (e.g., 2025-12)"), nil, ""))
		return
	}

	// Parse contractor parameter (optional)
	contractorDiscord := strings.TrimSpace(c.Query("contractor"))

	l.Info(fmt.Sprintf("syncing task order logs: month=%s contractor=%s", month, contractorDiscord))

	// Get services
	taskOrderLogService := h.service.Notion.TaskOrderLog
	if taskOrderLogService == nil {
		err := fmt.Errorf("task order log service not configured")
		l.Error(err, "task order log service is nil")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	openRouterService := h.service.OpenRouter
	if openRouterService == nil {
		err := fmt.Errorf("openrouter service not configured")
		l.Error(err, "openrouter service is nil")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Step 1: Query approved timesheets for the month
	l.Debug(fmt.Sprintf("querying approved timesheets for month: %s", month))
	timesheets, err := taskOrderLogService.QueryApprovedTimesheetsByMonth(ctx, month, contractorDiscord)
	if err != nil {
		l.Error(err, fmt.Sprintf("failed to query timesheets: month=%s", month))
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("found %d approved timesheets", len(timesheets)))

	if len(timesheets) == 0 {
		l.Info("no approved timesheets found, returning success with zero counts")
		c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
			"month":                  month,
			"orders_created":         0,
			"line_items_created":     0,
			"contractors_processed":  0,
			"details":                []any{},
		}, nil, nil, nil, "ok"))
		return
	}

	// Step 2: Group timesheets by Contractor → Project
	l.Debug("grouping timesheets by contractor and project")
	contractorGroups := groupTimesheetsByContractor(timesheets)
	l.Debug(fmt.Sprintf("grouped into %d contractors", len(contractorGroups)))

	// Step 3: Process each contractor
	var (
		ordersCreated      = 0
		lineItemsCreated   = 0
		contractorsProcessed = 0
		details            = []map[string]any{}
	)

	for contractorID, contractorTimesheets := range contractorGroups {
		contractorDetails := map[string]any{
			"contractor_id": contractorID,
			"projects":      []map[string]any{},
		}

		l.Debug(fmt.Sprintf("processing contractor: %s (%d timesheets)", contractorID, len(contractorTimesheets)))

		// Step 3a: Group timesheets by project
		projectGroups := groupTimesheetsByProject(contractorTimesheets)
		l.Debug(fmt.Sprintf("contractor %s has %d projects", contractorID, len(projectGroups)))

		// Step 3b: Get first project's deployment for Order
		var firstDeploymentID string
		for projectID := range projectGroups {
			deploymentID, err := taskOrderLogService.GetDeploymentByContractorAndProject(ctx, contractorID, projectID)
			if err != nil {
				l.Error(err, fmt.Sprintf("skipping project %s for contractor %s: no deployment found", projectID, contractorID))
				continue
			}
			firstDeploymentID = deploymentID
			l.Debug(fmt.Sprintf("using deployment %s from first project %s for order", firstDeploymentID, projectID))
			break
		}

		if firstDeploymentID == "" {
			l.Error(fmt.Errorf("no deployment found"), fmt.Sprintf("skipping contractor %s: no deployment found for any project", contractorID))
			continue
		}

		// Step 3c: Check if order already exists for contractor+month
		orderExists, orderID, err := taskOrderLogService.CheckOrderExistsByContractor(ctx, contractorID, month)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to check order existence for contractor: %s", contractorID))
			continue
		}

		if !orderExists {
			// Create new Order entry with first deployment
			l.Debug(fmt.Sprintf("creating new order for contractor: %s with deployment: %s", contractorID, firstDeploymentID))
			orderID, err = taskOrderLogService.CreateOrder(ctx, firstDeploymentID, month)
			if err != nil {
				l.Error(err, fmt.Sprintf("failed to create order for contractor: %s", contractorID))
				continue
			}
			ordersCreated++
			l.Info(fmt.Sprintf("created order: %s for contractor: %s", orderID, contractorID))
		} else {
			l.Debug(fmt.Sprintf("order already exists: %s for contractor: %s", orderID, contractorID))
		}

		contractorDetails["order_page_id"] = orderID

		totalHours := 0.0

		// Step 3d: Create line items for each project
		for projectID, projectTimesheets := range projectGroups {
			l.Debug(fmt.Sprintf("processing project: %s (%d timesheets)", projectID, len(projectTimesheets)))

			// Get deployment for this contractor+project
			deploymentID, err := taskOrderLogService.GetDeploymentByContractorAndProject(ctx, contractorID, projectID)
			if err != nil {
				l.Error(err, fmt.Sprintf("skipping project %s for contractor %s: no deployment found", projectID, contractorID))
				continue
			}

			l.Debug(fmt.Sprintf("found deployment: %s for contractor: %s project: %s", deploymentID, contractorID, projectID))

			// Aggregate hours
			var (
				hours         = 0.0
				proofOfWorks  = []string{}
				timesheetIDs  = []string{}
			)

			for _, ts := range projectTimesheets {
				hours += ts.Hours
				if ts.Title != "" { // Title contains Proof of Works from query
					proofOfWorks = append(proofOfWorks, ts.Title)
				}
				timesheetIDs = append(timesheetIDs, ts.PageID)
			}

			totalHours += hours

			l.Debug(fmt.Sprintf("project %s: %.1f hours, %d proof of works", projectID, hours, len(proofOfWorks)))

			// Summarize proof of works using LLM
			var summarizedPoW string
			if len(proofOfWorks) > 0 {
				l.Debug(fmt.Sprintf("summarizing %d proof of works for project: %s", len(proofOfWorks), projectID))
				summarizedPoW, err = openRouterService.SummarizeProofOfWorks(ctx, proofOfWorks)
				if err != nil {
					l.Error(err, fmt.Sprintf("failed to summarize proof of works, using original text: project=%s", projectID))
					// Fallback: use concatenated original text
					summarizedPoW = strings.Join(proofOfWorks, "\n\n")
				} else if strings.TrimSpace(summarizedPoW) == "" {
					// OpenRouter returned empty summary, use original text
					l.Debug(fmt.Sprintf("OpenRouter returned empty summary, using original text for project: %s", projectID))
					summarizedPoW = strings.Join(proofOfWorks, "\n\n")
				}
			}

			// Create timesheet line item
			l.Debug(fmt.Sprintf("creating line item: project=%s hours=%.1f deployment=%s", projectID, hours, deploymentID))
			lineItemID, err := taskOrderLogService.CreateTimesheetLineItem(ctx, orderID, deploymentID, projectID, hours, summarizedPoW, timesheetIDs, month)
			if err != nil {
				l.Error(err, fmt.Sprintf("failed to create line item: project=%s", projectID))
				continue
			}

			lineItemsCreated++
			l.Info(fmt.Sprintf("created line item: %s for project: %s (%.1f hours)", lineItemID, projectID, hours))

			// Add to project details
			projectDetails := map[string]any{
				"project_id":            projectID,
				"line_item_page_id":     lineItemID,
				"hours":                 hours,
				"timesheets_aggregated": len(timesheetIDs),
			}
			contractorDetails["projects"] = append(contractorDetails["projects"].([]map[string]any), projectDetails)
		}

		contractorDetails["total_hours"] = totalHours
		contractorsProcessed++
		details = append(details, contractorDetails)
	}

	l.Info(fmt.Sprintf("sync complete: orders_created=%d line_items_created=%d contractors_processed=%d", ordersCreated, lineItemsCreated, contractorsProcessed))

	// Return response
	c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
		"month":                  month,
		"orders_created":         ordersCreated,
		"line_items_created":     lineItemsCreated,
		"contractors_processed":  contractorsProcessed,
		"details":                details,
	}, nil, nil, nil, "ok"))
}

// Helper functions

// isValidMonthFormat validates month format as YYYY-MM
func isValidMonthFormat(month string) bool {
	if len(month) != 7 {
		return false
	}
	if month[4] != '-' {
		return false
	}
	// Basic validation - just check format
	return true
}

// groupTimesheetsByContractor groups timesheets by contractor page ID
func groupTimesheetsByContractor(timesheets []*notion.TimesheetEntry) map[string][]*notion.TimesheetEntry {
	groups := make(map[string][]*notion.TimesheetEntry)
	for _, ts := range timesheets {
		if ts.ContractorPageID == "" {
			continue
		}
		groups[ts.ContractorPageID] = append(groups[ts.ContractorPageID], ts)
	}
	return groups
}

// groupTimesheetsByProject groups timesheets by project page ID
func groupTimesheetsByProject(timesheets []*notion.TimesheetEntry) map[string][]*notion.TimesheetEntry {
	groups := make(map[string][]*notion.TimesheetEntry)
	for _, ts := range timesheets {
		if ts.ProjectPageID == "" {
			continue
		}
		groups[ts.ProjectPageID] = append(groups[ts.ProjectPageID], ts)
	}
	return groups
}
