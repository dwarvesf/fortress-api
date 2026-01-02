package notion

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/service/openrouter"
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
// @Param project query string false "Project code/name to filter by specific project (e.g., kafi, nghenhan)"
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

	// Parse project parameter (optional)
	projectID := strings.TrimSpace(c.Query("project"))

	l.Info(fmt.Sprintf("syncing task order logs: month=%s contractor=%s project=%s", month, contractorDiscord, projectID))

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
	timesheets, err := taskOrderLogService.QueryApprovedTimesheetsByMonth(ctx, month, contractorDiscord, projectID)
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
				hours        = 0.0
				proofOfWorks = []openrouter.ProofOfWorkEntry{}
				timesheetIDs = []string{}
			)

			for _, ts := range projectTimesheets {
				hours += ts.Hours
				if ts.Title != "" { // Title contains Proof of Works from query
					proofOfWorks = append(proofOfWorks, openrouter.ProofOfWorkEntry{
						Text:  ts.Title,
						Hours: ts.Hours,
					})
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
					var texts []string
					for _, pow := range proofOfWorks {
						texts = append(texts, pow.Text)
					}
					summarizedPoW = strings.Join(texts, "\n\n")
				} else if strings.TrimSpace(summarizedPoW) == "" {
					// OpenRouter returned empty summary, use original text
					l.Debug(fmt.Sprintf("OpenRouter returned empty summary, using original text for project: %s", projectID))
					var texts []string
					for _, pow := range proofOfWorks {
						texts = append(texts, pow.Text)
					}
					summarizedPoW = strings.Join(texts, "\n\n")
				}
			}

			// Check if line item already exists
			lineItemExists, lineItemID, err := taskOrderLogService.CheckLineItemExists(ctx, orderID, deploymentID)
			if err != nil {
				l.Error(err, fmt.Sprintf("failed to check line item existence: order=%s deployment=%s", orderID, deploymentID))
				continue
			}

			if !lineItemExists {
				// Create timesheet line item
				l.Debug(fmt.Sprintf("creating line item: project=%s hours=%.1f deployment=%s", projectID, hours, deploymentID))
				lineItemID, err = taskOrderLogService.CreateTimesheetLineItem(ctx, orderID, deploymentID, projectID, hours, summarizedPoW, timesheetIDs, month)
				if err != nil {
					l.Error(err, fmt.Sprintf("failed to create line item: project=%s", projectID))
					continue
				}

				lineItemsCreated++
				l.Info(fmt.Sprintf("created line item: %s for project: %s (%.1f hours)", lineItemID, projectID, hours))
			} else {
				l.Debug(fmt.Sprintf("line item already exists: %s for order: %s deployment: %s", lineItemID, orderID, deploymentID))
			}

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

// SendTaskOrderConfirmation godoc
// @Summary Send monthly task order confirmation emails
// @Description Sends task order confirmation emails to contractors with active client assignments via Gmail
// @Tags Cronjobs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param month query string false "Target month in YYYY-MM format (default: current month)"
// @Param discord query string false "Discord username to filter specific contractor"
// @Success 200 {object} view.Response
// @Failure 400 {object} view.Response
// @Failure 500 {object} view.Response
// @Router /cronjobs/send-task-order-confirmation [post]
func (h *handler) SendTaskOrderConfirmation(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "Notion",
		"method":  "SendTaskOrderConfirmation",
	})
	ctx := c.Request.Context()

	// Step 1: Parse and validate parameters
	month := c.Query("month")
	if month == "" {
		now := time.Now()
		month = now.Format("2006-01")
	}

	// Validate month format (YYYY-MM)
	if !isValidMonthFormat(month) {
		l.Error(fmt.Errorf("invalid month format"), fmt.Sprintf("month=%s", month))
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil,
			fmt.Errorf("invalid month format, expected YYYY-MM (e.g., 2026-01)"), nil, ""))
		return
	}

	contractorDiscord := strings.TrimSpace(c.Query("discord"))
	l.Info(fmt.Sprintf("sending task order confirmations: month=%s discord=%s", month, contractorDiscord))

	// Step 2: Get services
	taskOrderLogService := h.service.Notion.TaskOrderLog
	if taskOrderLogService == nil {
		err := fmt.Errorf("task order log service not configured")
		l.Error(err, "service is nil")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	googleMailService := h.service.GoogleMail
	if googleMailService == nil {
		err := fmt.Errorf("google mail service not configured")
		l.Error(err, "service is nil")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Step 3: Query active deployments
	l.Debug(fmt.Sprintf("querying active deployments for month: %s", month))
	deployments, err := taskOrderLogService.QueryActiveDeploymentsByMonth(ctx, month, contractorDiscord)
	if err != nil {
		l.Error(err, "failed to query deployments")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("found %d active deployments", len(deployments)))

	if len(deployments) == 0 {
		l.Info("no active deployments found")
		c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
			"month":         month,
			"emails_sent":   0,
			"emails_failed": 0,
			"details":       []any{},
		}, nil, nil, nil, "ok"))
		return
	}

	// Step 4: Group deployments by contractor
	contractorGroups := groupDeploymentsByContractor(deployments)
	l.Debug(fmt.Sprintf("grouped into %d contractors", len(contractorGroups)))

	// Step 5: Process each contractor
	var (
		emailsSent   = 0
		emailsFailed = 0
		details      = []map[string]any{}
	)

	for contractorID, contractorDeployments := range contractorGroups {
		detail := map[string]any{
			"contractor": "",
			"discord":    "",
			"email":      "",
			"status":     "",
			"clients":    []string{},
		}

		// Step 5a: Get contractor info
		name, discord := taskOrderLogService.GetContractorInfo(ctx, contractorID)
		if name == "" {
			l.Warn(fmt.Sprintf("skipping contractor %s: no name found", contractorID))
			continue
		}

		// Get team email
		teamEmail := taskOrderLogService.GetContractorTeamEmail(ctx, contractorID)
		if teamEmail == "" {
			l.Warn(fmt.Sprintf("skipping contractor %s: no team email", name))
			detail["contractor"] = name
			detail["discord"] = discord
			detail["status"] = "failed"
			detail["error"] = "no team email found"
			emailsFailed++
			details = append(details, detail)
			continue
		}

		detail["contractor"] = name
		detail["discord"] = discord
		detail["email"] = teamEmail

		// Step 5b: Extract client info from deployments
		var clients []model.TaskOrderClient
		var clientStrings []string
		for _, deployment := range contractorDeployments {
			clientInfo, err := taskOrderLogService.GetClientInfo(ctx, deployment.ProjectPageID)
			if err != nil {
				l.Error(err, fmt.Sprintf("failed to get client for project %s", deployment.ProjectPageID))
				continue
			}
			if clientInfo != nil && clientInfo.Name != "" {
				clients = append(clients, model.TaskOrderClient{
					Name:    clientInfo.Name,
					Country: clientInfo.Country,
				})
				clientStr := clientInfo.Name
				if clientInfo.Country != "" {
					clientStr = fmt.Sprintf("%s (%s)", clientInfo.Name, clientInfo.Country)
				}
				clientStrings = append(clientStrings, clientStr)
			}
		}

		if len(clients) == 0 {
			l.Warn(fmt.Sprintf("skipping contractor %s: no clients found", name))
			detail["status"] = "failed"
			detail["error"] = "no clients found"
			emailsFailed++
			details = append(details, detail)
			continue
		}

		detail["clients"] = clientStrings

		// Step 5c: Prepare email data
		emailData := &model.TaskOrderConfirmationEmail{
			ContractorName: name,
			TeamEmail:      teamEmail,
			Month:          month,
			Clients:        clients,
		}

		// Step 5d: Send email via Gmail using accounting refresh token
		err = googleMailService.SendTaskOrderConfirmationMail(emailData)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to send email to %s (%s)", name, teamEmail))
			detail["status"] = "failed"
			detail["error"] = err.Error()
			emailsFailed++
		} else {
			l.Info(fmt.Sprintf("sent email to %s (%s)", name, teamEmail))
			detail["status"] = "sent"
			emailsSent++
		}

		details = append(details, detail)
	}

	// Step 6: Return response
	l.Info(fmt.Sprintf("email sending complete: sent=%d failed=%d", emailsSent, emailsFailed))

	c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
		"month":         month,
		"emails_sent":   emailsSent,
		"emails_failed": emailsFailed,
		"details":       details,
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

// groupDeploymentsByContractor groups deployments by contractor page ID
func groupDeploymentsByContractor(deployments []*notion.DeploymentData) map[string][]*notion.DeploymentData {
	groups := make(map[string][]*notion.DeploymentData)
	for _, deployment := range deployments {
		if deployment.ContractorPageID == "" {
			continue
		}
		groups[deployment.ContractorPageID] = append(groups[deployment.ContractorPageID], deployment)
	}
	return groups
}
