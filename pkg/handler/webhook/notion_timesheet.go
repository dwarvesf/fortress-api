package webhook

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

// NotionTimesheetWebhookPayload represents the webhook payload from Notion
type NotionTimesheetWebhookPayload struct {
	// Verification fields (for endpoint verification challenge)
	VerificationToken string `json:"verification_token"` // Verification token from Notion
	Challenge         string `json:"challenge"`          // Challenge string to respond with

	// Event fields
	Source *NotionTimesheetWebhookSource `json:"source"` // Automation source info
	Data   *NotionTimesheetWebhookData   `json:"data"`   // The page data
}

// NotionTimesheetWebhookSource represents the automation source
type NotionTimesheetWebhookSource struct {
	Type         string `json:"type"`          // "automation"
	AutomationID string `json:"automation_id"` // Automation ID
}

// NotionTimesheetWebhookData represents the page data in the webhook payload
type NotionTimesheetWebhookData struct {
	Object string `json:"object"` // "page"
	ID     string `json:"id"`     // Page ID
}

// HandleNotionTimesheet handles timesheet entry webhook events from Notion
// Automatically fills missing Contractor and Discord fields based on Created By user
func (h *handler) HandleNotionTimesheet(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "webhook",
		"method":  "HandleNotionTimesheet",
	})

	l.Debug("received notion timesheet webhook request")

	// Read body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		l.Error(err, "failed to read notion timesheet webhook body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("received webhook body: %s", string(body)))

	// Parse payload
	var payload NotionTimesheetWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		l.Error(err, "failed to parse notion timesheet webhook payload")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Handle verification challenge
	if payload.VerificationToken != "" {
		l.Debug(fmt.Sprintf("responding to notion webhook verification: token=%s challenge=%s", payload.VerificationToken, payload.Challenge))
		if payload.Challenge != "" {
			c.JSON(http.StatusOK, gin.H{"challenge": payload.Challenge})
		} else {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		}
		return
	}

	// Verify webhook signature if verification token is configured
	verificationToken := h.config.Notion.Secret
	if verificationToken != "" {
		signature := c.GetHeader("X-Notion-Signature")
		l.Debug(fmt.Sprintf("signature header: %s", signature))
		if signature != "" {
			if !h.verifyNotionWebhookSignature(body, signature, verificationToken) {
				l.Error(errors.New("invalid signature"), "webhook signature verification failed")
				c.JSON(http.StatusUnauthorized, view.CreateResponse[any](nil, nil, errors.New("invalid signature"), nil, ""))
				return
			}
			l.Debug("webhook signature verified successfully")
		}
	}

	// Validate payload has data with page ID
	if payload.Data == nil || payload.Data.ID == "" {
		l.Error(errors.New("missing data.id"), "no data.id in webhook payload")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errors.New("missing data.id"), nil, ""))
		return
	}

	// Only process page objects
	if payload.Data.Object != "page" {
		l.Debug(fmt.Sprintf("ignoring non-page object type: %s", payload.Data.Object))
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ignored"))
		return
	}

	// Only process automation events (from Notion automations)
	if payload.Source == nil || payload.Source.Type != "automation" {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ignored"))
		return
	}

	pageID := payload.Data.ID
	l.Debug(fmt.Sprintf("parsed webhook payload: source_type=%s page_id=%s", payload.Source.Type, pageID))

	// Create timesheet service
	timesheetService := notion.NewTimesheetService(h.config, h.logger)
	if timesheetService == nil {
		l.Error(errors.New("failed to create timesheet service"), "notion timesheet service not configured")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, errors.New("service not configured"), nil, ""))
		return
	}

	// Fetch timesheet entry from Notion
	ctx := c.Request.Context()
	entry, err := timesheetService.GetTimesheetEntry(ctx, pageID)
	if err != nil {
		l.Error(err, fmt.Sprintf("failed to fetch timesheet entry from Notion: page_id=%s", pageID))
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("fetched timesheet entry: page_id=%s created_by=%s contractor=%s",
		pageID, entry.CreatedByUserID, entry.ContractorPageID))

	// Check if contractor is already filled
	if entry.ContractorPageID != "" {
		l.Debug(fmt.Sprintf("timesheet entry already has contractor filled: page_id=%s", pageID))
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "already_filled"))
		return
	}

	// Find contractor by created by user ID
	if entry.CreatedByUserID == "" {
		l.Error(errors.New("missing created_by user ID"), "timesheet entry has no created_by user")
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "no_created_by"))
		return
	}

	contractorPageID, _, err := timesheetService.FindContractorByPersonID(ctx, entry.CreatedByUserID)
	if err != nil {
		l.Error(err, fmt.Sprintf("contractor not found for created_by user: user_id=%s", entry.CreatedByUserID))
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "contractor_not_found"))
		return
	}

	l.Debug(fmt.Sprintf("found contractor: user_id=%s contractor_id=%s",
		entry.CreatedByUserID, contractorPageID))

	// Update timesheet entry with contractor
	err = timesheetService.UpdateTimesheetEntry(ctx, pageID, contractorPageID)
	if err != nil {
		l.Error(err, fmt.Sprintf("failed to update timesheet entry: page_id=%s", pageID))
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Info(fmt.Sprintf("successfully filled timesheet entry: page_id=%s contractor=%s",
		pageID, contractorPageID))

	c.JSON(http.StatusOK, view.CreateResponse[any](gin.H{
		"page_id":            pageID,
		"contractor_updated": true,
	}, nil, nil, nil, "success"))
}
