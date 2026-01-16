package notion

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
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
			"month":                 month,
			"orders_created":        0,
			"line_items_created":    0,
			"contractors_processed": 0,
			"details":               []any{},
		}, nil, nil, nil, "ok"))
		return
	}

	// Step 2: Group timesheets by Contractor → Project
	l.Debug("grouping timesheets by contractor and project")
	contractorGroups := groupTimesheetsByContractor(timesheets)
	l.Debug(fmt.Sprintf("grouped into %d contractors", len(contractorGroups)))

	// Step 3: Process each contractor
	var (
		ordersCreated        = 0
		lineItemsCreated     = 0
		lineItemsUpdated     = 0
		contractorsProcessed = 0
		details              = []map[string]any{}
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

		// Step 3b: Check if order already exists for contractor+month
		orderExists, orderID, err := taskOrderLogService.CheckOrderExistsByContractor(ctx, contractorID, month)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to check order existence for contractor: %s", contractorID))
			continue
		}

		if !orderExists {
			// Create new Order entry (without Deployment - ADR-002)
			l.Debug(fmt.Sprintf("creating new order for contractor: %s", contractorID))
			orderID, err = taskOrderLogService.CreateOrder(ctx, month)
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
				hours += ts.ApproxEffort
				if ts.Title != "" { // Title contains Key deliverables from query
					proofOfWorks = append(proofOfWorks, openrouter.ProofOfWorkEntry{
						Text:  ts.Title,
						Hours: ts.ApproxEffort,
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
				// Line item exists - check if update is needed (upsert logic)
				l.Debug(fmt.Sprintf("line item exists: %s, checking for changes", lineItemID))

				existingDetails, err := taskOrderLogService.GetLineItemDetails(ctx, lineItemID)
				if err != nil {
					l.Error(err, fmt.Sprintf("failed to get line item details: %s", lineItemID))
					continue
				}

				// Compare: hours changed OR timesheet IDs changed
				hoursChanged := existingDetails.Hours != hours
				timesheetsChanged := !equalStringSlices(existingDetails.TimesheetIDs, timesheetIDs)

				l.Debug(fmt.Sprintf("line item comparison: hoursChanged=%v (%.2f->%.2f) timesheetsChanged=%v (%d->%d)",
					hoursChanged, existingDetails.Hours, hours,
					timesheetsChanged, len(existingDetails.TimesheetIDs), len(timesheetIDs)))

				if hoursChanged || timesheetsChanged {
					l.Debug(fmt.Sprintf("line item changed, updating: hours %.2f->%.2f, timesheets %d->%d",
						existingDetails.Hours, hours,
						len(existingDetails.TimesheetIDs), len(timesheetIDs)))

					// Update line item
					err = taskOrderLogService.UpdateTimesheetLineItem(ctx, lineItemID, orderID, hours, summarizedPoW, timesheetIDs)
					if err != nil {
						l.Error(err, fmt.Sprintf("failed to update line item: %s", lineItemID))
						continue
					}

					lineItemsUpdated++
					l.Info(fmt.Sprintf("updated line item: %s for project: %s (%.1f hours)", lineItemID, projectID, hours))
				} else {
					l.Debug(fmt.Sprintf("line item unchanged, skipping update: %s", lineItemID))
				}
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

	l.Info(fmt.Sprintf("sync complete: orders_created=%d line_items_created=%d line_items_updated=%d contractors_processed=%d", ordersCreated, lineItemsCreated, lineItemsUpdated, contractorsProcessed))

	// Return response
	c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
		"month":                 month,
		"orders_created":        ordersCreated,
		"line_items_created":    lineItemsCreated,
		"line_items_updated":    lineItemsUpdated,
		"contractors_processed": contractorsProcessed,
		"details":               details,
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
// @Param test_email query string false "Override recipient email for testing purposes"
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
	testEmail := strings.TrimSpace(c.Query("test_email"))

	l.Info(fmt.Sprintf("sending task order confirmations: month=%s discord=%s test_email=%s", month, contractorDiscord, testEmail))

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

		// Use test email if provided
		emailToSend := teamEmail
		if testEmail != "" {
			l.Info(fmt.Sprintf("overriding recipient email for %s: %s -> %s", name, teamEmail, testEmail))
			emailToSend = testEmail
		}

		detail["contractor"] = name
		detail["discord"] = discord
		detail["email"] = emailToSend
		if testEmail != "" {
			detail["original_email"] = teamEmail
		}

		// Step 5b: Extract client info from deployments
		var clients []model.TaskOrderClient
		var clientStrings []string
		seenClients := make(map[string]bool)

		for _, deployment := range contractorDeployments {
			clientInfo, err := taskOrderLogService.GetClientInfo(ctx, deployment.ProjectPageID)
			if err != nil {
				l.Error(err, fmt.Sprintf("failed to get client for project %s", deployment.ProjectPageID))
				continue
			}
			if clientInfo != nil && clientInfo.Name != "" {
				// Check if deployment Type contains "Shadow"
				isShadow := false
				for _, t := range deployment.Type {
					if t == "Shadow" {
						isShadow = true
						l.Debug(fmt.Sprintf("deployment %s is Shadow type, will use Dwarves LLC", deployment.PageID))
						break
					}
				}

				// If Shadow deployment, use "Dwarves LLC" (USA)
				if isShadow {
					l.Debug(fmt.Sprintf("replacing client %s (%s) with Dwarves LLC (USA) for Shadow deployment", clientInfo.Name, clientInfo.Country))
					clientInfo.Name = "Dwarves LLC"
					clientInfo.Country = "USA"
				} else if strings.TrimSpace(clientInfo.Country) == "Vietnam" {
					// If client is in Vietnam, use "Dwarves LLC" (USA) instead
					l.Debug(fmt.Sprintf("replacing Vietnam client %s with Dwarves LLC (USA)", clientInfo.Name))
					clientInfo.Name = "Dwarves LLC"
					clientInfo.Country = "USA"
				}

				clientKey := fmt.Sprintf("%s:%s", clientInfo.Name, clientInfo.Country)
				if seenClients[clientKey] {
					continue
				}
				seenClients[clientKey] = true

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

		// Step 5c: Fetch contractor payday and calculate invoice due day
		payday, err := taskOrderLogService.GetContractorPayday(ctx, contractorID)
		if err != nil {
			// Error already logged in service layer, continue with default
			l.Debug(fmt.Sprintf("using default payday for contractor %s: %v", name, err))
			payday = 0
		}

		// Calculate invoice due day based on payday
		invoiceDueDay := "10th" // Default for Payday 1 or fallback
		if payday == 15 {
			invoiceDueDay = "25th"
		}

		l.Debug(fmt.Sprintf("contractor %s: payday=%d invoice_due_day=%s", name, payday, invoiceDueDay))

		// Step 5d: Create mock milestones (TODO: Replace with real data source)
		milestones := []string{
			"Client 1 – Feature Y demo expected mid-month",
			"Dwarves LLC – Sprint review scheduled end of month",
		}

		// Step 5e: Prepare email data
		emailData := &model.TaskOrderConfirmationEmail{
			ContractorName: name,
			TeamEmail:      emailToSend,
			Month:          month,
			Clients:        clients,
			InvoiceDueDay:  invoiceDueDay,
			Milestones:     milestones,
		}

		// Step 5d: Send email via Gmail using accounting refresh token
		err = googleMailService.SendTaskOrderConfirmationMail(emailData)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to send email to %s (%s)", name, emailToSend))
			detail["status"] = "failed"
			detail["error"] = err.Error()
			emailsFailed++
		} else {
			l.Info(fmt.Sprintf("sent email to %s (%s)", name, emailToSend))
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

// equalStringSlices compares two string slices for equality (order-independent)
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aMap := make(map[string]bool)
	for _, v := range a {
		aMap[v] = true
	}
	for _, v := range b {
		if !aMap[v] {
			return false
		}
	}
	return true
}

// contractorProcessResult holds the result of processing a single contractor
type contractorProcessResult struct {
	ContractorID         string
	OrderCreated         bool
	LineItemsCreated     int
	DeploymentsProcessed int
	Skipped              int
	Detail               map[string]any
	Error                error
}

// InitTaskOrderLogs godoc
// @Summary Initialize task order logs for all active deployments
// @Description Creates empty Task Order Log entries for all active deployments for a given month
// @Tags Cronjobs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param month query string false "Target month in YYYY-MM format (e.g., 2025-12). Defaults to current month if not provided"
// @Param contractor query string false "Discord username to filter by specific contractor (e.g., chinhld)"
// @Success 200 {object} view.Response
// @Failure 400 {object} view.Response
// @Failure 500 {object} view.Response
// @Router /cronjobs/init-task-order-logs [post]
func (h *handler) InitTaskOrderLogs(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "Notion",
		"method":  "InitTaskOrderLogs",
	})
	ctx := c.Request.Context()

	// Parse month parameter (optional, defaults to current month)
	month := c.Query("month")
	if month == "" {
		month = time.Now().Format("2006-01")
		l.Debug(fmt.Sprintf("no month parameter provided, using current month: %s", month))
	}

	// Validate month format (YYYY-MM)
	if !isValidMonthFormat(month) {
		l.Error(fmt.Errorf("invalid month format"), fmt.Sprintf("month=%s", month))
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, fmt.Errorf("invalid month format, expected YYYY-MM (e.g., 2025-12)"), nil, ""))
		return
	}

	// Parse contractor parameter (optional)
	contractorDiscord := strings.TrimSpace(c.Query("contractor"))

	l.Info(fmt.Sprintf("initializing task order logs: month=%s contractor=%s", month, contractorDiscord))
	l.Debug(fmt.Sprintf("starting initialization for month: %s contractor=%s", month, contractorDiscord))

	// Get services
	taskOrderLogService := h.service.Notion.TaskOrderLog
	if taskOrderLogService == nil {
		err := fmt.Errorf("task order log service not configured")
		l.Error(err, "task order log service is nil")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Step 1: Query all active deployments for the month
	l.Debug(fmt.Sprintf("querying active deployments for month: %s contractor=%s", month, contractorDiscord))
	deployments, err := taskOrderLogService.QueryActiveDeploymentsByMonth(ctx, month, contractorDiscord)
	if err != nil {
		l.Error(err, fmt.Sprintf("failed to query active deployments: month=%s", month))
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("found %d active deployments", len(deployments)))

	if len(deployments) == 0 {
		l.Info("no active deployments found, returning success with zero counts")
		c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
			"month":                 month,
			"orders_created":        0,
			"line_items_created":    0,
			"deployments_processed": 0,
			"skipped":               0,
			"details":               []any{},
		}, nil, nil, nil, "ok"))
		return
	}

	// Step 2: Group deployments by contractor
	l.Debug("grouping deployments by contractor")
	contractorDeployments := groupDeploymentsByContractor(deployments)
	l.Debug(fmt.Sprintf("found %d unique contractors", len(contractorDeployments)))

	// Step 3: Process contractors concurrently
	const maxConcurrency = 10 // Limit concurrent API calls to avoid rate limiting
	l.Debug(fmt.Sprintf("processing %d contractors with max concurrency: %d", len(contractorDeployments), maxConcurrency))

	// Create a channel for results
	resultChan := make(chan contractorProcessResult, len(contractorDeployments))

	// Create a semaphore to limit concurrency
	sem := make(chan struct{}, maxConcurrency)

	// Use WaitGroup to wait for all goroutines
	var wg sync.WaitGroup

	// Launch goroutines for each contractor
	for contractorID, contractorDeps := range contractorDeployments {
		wg.Add(1)
		go func(cID string, cDeps []*notion.DeploymentData) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			l.Debug(fmt.Sprintf("[CONCURRENT] starting processing contractor: %s with %d deployments", cID, len(cDeps)))

			result := h.processContractorInit(ctx, l, taskOrderLogService, cID, cDeps, month)
			resultChan <- result

			l.Debug(fmt.Sprintf("[CONCURRENT] finished processing contractor: %s", cID))
		}(contractorID, contractorDeps)
	}

	// Close result channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	ordersCreated := 0
	lineItemsCreated := 0
	deploymentsProcessed := 0
	skipped := 0
	var details []map[string]any

	for result := range resultChan {
		if result.Error != nil {
			l.Error(result.Error, fmt.Sprintf("error processing contractor: %s", result.ContractorID))
			continue
		}

		if result.OrderCreated {
			ordersCreated++
		}
		lineItemsCreated += result.LineItemsCreated
		deploymentsProcessed += result.DeploymentsProcessed
		skipped += result.Skipped
		details = append(details, result.Detail)
	}

	l.Info(fmt.Sprintf("initialization complete: orders_created=%d line_items_created=%d deployments_processed=%d skipped=%d",
		ordersCreated, lineItemsCreated, deploymentsProcessed, skipped))

	c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
		"month":                 month,
		"orders_created":        ordersCreated,
		"line_items_created":    lineItemsCreated,
		"deployments_processed": deploymentsProcessed,
		"skipped":               skipped,
		"details":               details,
	}, nil, nil, nil, "ok"))
}

// processContractorInit processes a single contractor's task order initialization
func (h *handler) processContractorInit(
	ctx context.Context,
	l logger.Logger,
	taskOrderLogService *notion.TaskOrderLogService,
	contractorID string,
	contractorDeps []*notion.DeploymentData,
	month string,
) contractorProcessResult {
	result := contractorProcessResult{
		ContractorID: contractorID,
		Detail: map[string]any{
			"contractor_id":      contractorID,
			"deployments":        []string{},
			"line_items_created": 0,
		},
	}

	l.Debug(fmt.Sprintf("processing contractor: %s with %d deployments", contractorID, len(contractorDeps)))

	var deploymentIDs []string

	// Check if Order exists for this contractor and month
	orderExists, orderID, err := taskOrderLogService.CheckOrderExistsByContractor(ctx, contractorID, month)
	if err != nil {
		l.Error(err, fmt.Sprintf("failed to check order existence: contractor=%s month=%s", contractorID, month))
		result.Error = err
		return result
	}

	// Create Order if it doesn't exist
	if !orderExists {
		l.Debug(fmt.Sprintf("creating order for contractor: %s month=%s", contractorID, month))
		orderID, err = taskOrderLogService.CreateOrder(ctx, month)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to create order: contractor=%s month=%s", contractorID, month))
			result.Error = err
			return result
		}
		result.OrderCreated = true
		l.Debug(fmt.Sprintf("created order: %s for contractor: %s", orderID, contractorID))
	} else {
		l.Debug(fmt.Sprintf("order already exists: %s for contractor: %s", orderID, contractorID))
	}

	result.Detail["order_page_id"] = orderID

	// Process each deployment for this contractor
	for _, deployment := range contractorDeps {
		deploymentID := deployment.PageID
		l.Debug(fmt.Sprintf("processing deployment: %s for contractor: %s", deploymentID, contractorID))

		// Check if Line Item already exists for this order and deployment
		exists, existingLineItemID, err := taskOrderLogService.CheckLineItemExists(ctx, orderID, deploymentID)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to check line item existence: order=%s deployment=%s", orderID, deploymentID))
			continue
		}

		if exists {
			l.Debug(fmt.Sprintf("line item already exists: %s for deployment: %s", existingLineItemID, deploymentID))
			result.Skipped++
			continue
		}

		// Create empty Line Item
		l.Debug(fmt.Sprintf("creating empty line item: order=%s deployment=%s", orderID, deploymentID))
		lineItemID, err := taskOrderLogService.CreateEmptyTimesheetLineItem(ctx, orderID, deploymentID, month)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to create empty line item: order=%s deployment=%s", orderID, deploymentID))
			continue
		}

		l.Debug(fmt.Sprintf("created empty line item: %s for deployment: %s", lineItemID, deploymentID))
		result.LineItemsCreated++
		result.DeploymentsProcessed++
		deploymentIDs = append(deploymentIDs, deploymentID)
	}

	result.Detail["deployments"] = deploymentIDs
	result.Detail["line_items_created"] = len(deploymentIDs)

	// Generate and append confirmation content to Order page
	// Only generate if we created the order (new order)
	if result.OrderCreated || !orderExists {
		l.Debug(fmt.Sprintf("collecting client info for contractor: %s", contractorID))

		// Collect client info from all deployments
		var clients []model.TaskOrderClient
		seenClients := make(map[string]bool)

		for _, deployment := range contractorDeps {
			clientInfo, err := taskOrderLogService.GetClientInfo(ctx, deployment.ProjectPageID)
			if err != nil {
				l.Error(err, fmt.Sprintf("failed to get client for project %s", deployment.ProjectPageID))
				continue
			}
			if clientInfo != nil && clientInfo.Name != "" {
				// If client is in Vietnam, use "Dwarves LLC" (USA) instead
				if strings.TrimSpace(clientInfo.Country) == "Vietnam" {
					clientInfo.Name = "Dwarves LLC"
					clientInfo.Country = "USA"
				}

				clientKey := fmt.Sprintf("%s:%s", clientInfo.Name, clientInfo.Country)
				if seenClients[clientKey] {
					continue
				}
				seenClients[clientKey] = true

				clients = append(clients, model.TaskOrderClient{
					Name:    clientInfo.Name,
					Country: clientInfo.Country,
				})
			}
		}

		if len(clients) > 0 {
			// Get contractor name
			contractorName, _ := taskOrderLogService.GetContractorInfo(ctx, contractorID)
			if contractorName == "" {
				contractorName = "Contractor"
			}

			// Generate HTML confirmation content from template
			l.Debug(fmt.Sprintf("generating HTML confirmation content for contractor: %s with %d clients", contractorName, len(clients)))
			htmlContent := taskOrderLogService.GenerateConfirmationHTML(contractorName, month, clients)

			// Append HTML content as code block to Order page
			l.Debug(fmt.Sprintf("appending HTML code block to order: %s", orderID))
			if err := taskOrderLogService.AppendCodeBlockToPage(ctx, orderID, htmlContent); err != nil {
				l.Error(err, fmt.Sprintf("failed to append HTML code block to order: %s", orderID))
				result.Detail["content_error"] = err.Error()
			} else {
				l.Debug(fmt.Sprintf("successfully appended HTML code block to order: %s", orderID))
				result.Detail["content_generated"] = true
			}
		} else {
			l.Debug(fmt.Sprintf("no clients found for contractor: %s, skipping content generation", contractorID))
		}
	}

	return result
}
