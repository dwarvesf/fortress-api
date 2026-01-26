package notion

import (
	"fmt"
	"net/http"
	"time"

	nt "github.com/dstotijn/go-notion"
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

// CloneMonthlyInvoicesRequest represents the query parameters for cloning invoices
type CloneMonthlyInvoicesRequest struct {
	Month     int      `form:"month"`
	Year      int      `form:"year"`
	ProjectID string   `form:"projectId"`
	Status    []string `form:"status"`
	DryRun    bool     `form:"dryRun"`
}

// CloneInvoiceDetail represents the result of cloning a single invoice
type CloneInvoiceDetail struct {
	InvoiceNumber   string `json:"invoiceNumber"`
	ProjectID       string `json:"projectId"`
	Status          string `json:"status"`
	NewPageID       string `json:"newPageId,omitempty"`
	Reason          string `json:"reason,omitempty"`
	LineItemsCloned int    `json:"lineItemsCloned,omitempty"`
}

// CloneMonthlyInvoicesResponse represents the response from the clone invoices endpoint
type CloneMonthlyInvoicesResponse struct {
	Cloned      int                  `json:"cloned"`
	Skipped     int                  `json:"skipped"`
	Errors      int                  `json:"errors"`
	DryRun      bool                 `json:"dryRun"`
	SourceMonth string               `json:"sourceMonth"`
	TargetMonth string               `json:"targetMonth"`
	Details     []CloneInvoiceDetail `json:"details"`
}

// CloneMonthlyInvoices godoc
// @Summary Clone client invoices from a source month to current month
// @Description Clones client invoices from a previous month to generate new invoices for the current month
// @Tags Cronjobs
// @Accept json
// @Produce json
// @Param month query int false "Source month (1-12, default: previous month)"
// @Param year query int false "Source year (default: current/previous year)"
// @Param projectId query string false "Optional Notion page ID to filter specific project"
// @Param status query []string false "Statuses to clone (default: ['Paid'])"
// @Param dryRun query bool false "If true, preview without creating (default: false)"
// @Security BearerAuth
// @Success 200 {object} view.Response{data=CloneMonthlyInvoicesResponse}
// @Failure 400 {object} view.Response
// @Failure 500 {object} view.Response
// @Router /cronjobs/clone-monthly-invoices [post]
func (h *handler) CloneMonthlyInvoices(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "Notion",
		"method":  "CloneMonthlyInvoices",
	})

	l.Info("starting CloneMonthlyInvoices cronjob")

	// Parse query parameters
	var req CloneMonthlyInvoicesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		l.Error(err, "failed to parse query parameters")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Set defaults
	now := time.Now()
	targetYear := now.Year()
	targetMonth := int(now.Month())
	targetIssueDate := now

	// Default source month is previous month
	sourceYear := targetYear
	sourceMonth := targetMonth - 1
	if sourceMonth < 1 {
		sourceMonth = 12
		sourceYear--
	}

	// Override with provided values
	if req.Year > 0 {
		sourceYear = req.Year
	}
	if req.Month > 0 {
		sourceMonth = req.Month
	}

	// Validate month
	if sourceMonth < 1 || sourceMonth > 12 {
		err := fmt.Errorf("invalid month: %d (must be 1-12)", sourceMonth)
		l.Error(err, "invalid month parameter")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Validate year
	if sourceYear < 2020 || sourceYear > 2100 {
		err := fmt.Errorf("invalid year: %d (must be 2020-2100)", sourceYear)
		l.Error(err, "invalid year parameter")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Default statuses
	statuses := req.Status
	if len(statuses) == 0 {
		statuses = []string{"Paid"}
	}

	l.Debug(fmt.Sprintf("parameters: sourceYear=%d, sourceMonth=%d, targetYear=%d, targetMonth=%d, statuses=%v, projectId=%s, dryRun=%v",
		sourceYear, sourceMonth, targetYear, targetMonth, statuses, req.ProjectID, req.DryRun))

	// Query invoices from source month
	invoices, err := h.service.Notion.QueryInvoicesByMonth(sourceYear, sourceMonth, statuses, req.ProjectID)
	if err != nil {
		l.Error(err, "failed to query invoices")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Info(fmt.Sprintf("found %d invoices to clone from %d-%02d", len(invoices), sourceYear, sourceMonth))

	// Process each invoice
	var cloned, skipped, errors int
	var details []CloneInvoiceDetail

	for _, invoice := range invoices {
		props, ok := invoice.Properties.(nt.DatabasePageProperties)
		if !ok {
			l.Debug(fmt.Sprintf("failed to cast invoice properties for page %s", invoice.ID))
			continue
		}

		// Extract invoice number for logging
		invoiceNumber := extractInvoiceNumberFromProps(props)

		// Extract project ID from relation
		projectPageID := extractProjectIDFromProps(props)
		if projectPageID == "" {
			l.Debug(fmt.Sprintf("skipping invoice %s: no project found", invoice.ID))
			details = append(details, CloneInvoiceDetail{
				InvoiceNumber: invoiceNumber,
				ProjectID:     "",
				Status:        "skipped",
				Reason:        "no project relation found",
			})
			skipped++
			continue
		}

		// Check if invoice already exists for target month (idempotency)
		exists, existingID, err := h.service.Notion.CheckInvoiceExistsForMonth(projectPageID, targetYear, targetMonth)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to check invoice existence for project %s", projectPageID))
			details = append(details, CloneInvoiceDetail{
				InvoiceNumber: invoiceNumber,
				ProjectID:     projectPageID,
				Status:        "error",
				Reason:        fmt.Sprintf("failed to check existence: %v", err),
			})
			errors++
			continue
		}

		if exists {
			l.Debug(fmt.Sprintf("invoice already exists for project %s in %d-%02d: %s", projectPageID, targetYear, targetMonth, existingID))
			details = append(details, CloneInvoiceDetail{
				InvoiceNumber: invoiceNumber,
				ProjectID:     projectPageID,
				Status:        "skipped",
				Reason:        "invoice already exists for target month",
				NewPageID:     existingID,
			})
			skipped++
			continue
		}

		// If dry run, just add to preview
		if req.DryRun {
			l.Debug(fmt.Sprintf("dry run: would clone invoice %s", invoice.ID))
			details = append(details, CloneInvoiceDetail{
				InvoiceNumber: invoiceNumber,
				ProjectID:     projectPageID,
				Status:        "would_clone",
			})
			cloned++
			continue
		}

		// Clone the invoice
		result, err := h.service.Notion.CloneInvoiceToNextMonth(invoice.ID, targetIssueDate)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to clone invoice %s", invoice.ID))
			details = append(details, CloneInvoiceDetail{
				InvoiceNumber: invoiceNumber,
				ProjectID:     projectPageID,
				Status:        "error",
				Reason:        fmt.Sprintf("failed to clone: %v", err),
			})
			errors++
			continue
		}

		l.Info(fmt.Sprintf("cloned invoice %s -> %s", invoice.ID, result.NewInvoicePageID))
		details = append(details, CloneInvoiceDetail{
			InvoiceNumber:   invoiceNumber,
			ProjectID:       projectPageID,
			Status:          "cloned",
			NewPageID:       result.NewInvoicePageID,
			LineItemsCloned: result.LineItemsCloned,
		})
		cloned++
	}

	// Build response
	response := CloneMonthlyInvoicesResponse{
		Cloned:      cloned,
		Skipped:     skipped,
		Errors:      errors,
		DryRun:      req.DryRun,
		SourceMonth: fmt.Sprintf("%d-%02d", sourceYear, sourceMonth),
		TargetMonth: fmt.Sprintf("%d-%02d", targetYear, targetMonth),
		Details:     details,
	}

	l.Info(fmt.Sprintf("clone complete: cloned=%d, skipped=%d, errors=%d, dryRun=%v",
		cloned, skipped, errors, req.DryRun))

	// Send Discord notification if not dry run and there were changes
	if !req.DryRun && (cloned > 0 || errors > 0) {
		h.sendCloneNotification(l, response)
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](response, nil, nil, nil, "ok"))
}

// extractInvoiceNumberFromProps extracts the Legacy Number from invoice properties
func extractInvoiceNumberFromProps(props nt.DatabasePageProperties) string {
	if legacyNumber, ok := props["Legacy Number"]; ok && legacyNumber.RichText != nil {
		if len(legacyNumber.RichText) > 0 {
			return legacyNumber.RichText[0].PlainText
		}
	}
	return ""
}

// extractProjectIDFromProps extracts the Project relation ID from invoice properties
func extractProjectIDFromProps(props nt.DatabasePageProperties) string {
	if project, ok := props["Project"]; ok && project.Relation != nil {
		if len(project.Relation) > 0 {
			return project.Relation[0].ID
		}
	}
	return ""
}

// sendCloneNotification sends a Discord notification about the clone results
func (h *handler) sendCloneNotification(l logger.Logger, response CloneMonthlyInvoicesResponse) {
	l.Debug("sending Discord notification for clone results")

	// Build notification message
	message := fmt.Sprintf("**Invoice Cloning Complete**\n"+
		"- Source: %s\n"+
		"- Target: %s\n"+
		"- Cloned: %d invoices\n"+
		"- Skipped: %d (already exist)\n"+
		"- Errors: %d",
		response.SourceMonth,
		response.TargetMonth,
		response.Cloned,
		response.Skipped,
		response.Errors,
	)

	// Get the audit log webhook URL from config (use AuditLog as fallback)
	webhookURL := h.config.Discord.Webhooks.AuditLog
	if webhookURL == "" {
		l.Debug("no audit log webhook URL configured, skipping notification")
		return
	}

	// Send via Discord service
	discordMsg := model.DiscordMessage{
		Content: message,
	}
	_, err := h.service.Discord.SendMessage(discordMsg, webhookURL)
	if err != nil {
		l.Error(err, "failed to send Discord notification")
	}
}
