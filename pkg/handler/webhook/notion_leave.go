package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	nt "github.com/dstotijn/go-notion"
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

// NotionLeaveEventType represents different leave webhook events
type NotionLeaveEventType string

const (
	NotionLeaveEventCreated  NotionLeaveEventType = "created"
	NotionLeaveEventApproved NotionLeaveEventType = "approved"
	NotionLeaveEventRejected NotionLeaveEventType = "rejected"
)

// NotionLeaveWebhookPayload represents the webhook payload from Notion
type NotionLeaveWebhookPayload struct {
	// Verification fields (for endpoint verification challenge)
	VerificationToken string `json:"verification_token"` // Verification token from Notion
	Challenge         string `json:"challenge"`          // Challenge string to respond with

	// Event fields
	Type   string                    `json:"type"`   // "page.created", "page.updated"
	Entity *NotionLeaveWebhookEntity `json:"entity"` // The entity that triggered the event
	Data   *NotionLeaveWebhookData   `json:"data"`   // Additional data (optional)
}

// NotionLeaveWebhookEntity represents the entity in the webhook payload
type NotionLeaveWebhookEntity struct {
	ID   string `json:"id"`   // Entity ID (page ID)
	Type string `json:"type"` // Entity type ("page", "database", etc.)
}

// NotionLeaveWebhookData contains additional webhook data
type NotionLeaveWebhookData struct {
	Status string `json:"status"` // Current status
}

// HandleNotionLeave handles all leave request webhook events from Notion
func (h *handler) HandleNotionLeave(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "webhook",
		"method":  "HandleNotionLeave",
	})

	l.Debug("received notion leave webhook request")

	// Read body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		l.Error(err, "failed to read notion leave webhook body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("received webhook body: %s", string(body)))

	// Parse payload
	var payload NotionLeaveWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		l.Error(err, "failed to parse notion leave webhook payload")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Handle verification challenge
	// When Notion sends a verification request, respond with the challenge or acknowledge the token
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
	verificationToken := h.config.LeaveIntegration.Notion.VerificationToken
	if verificationToken != "" {
		signature := c.GetHeader("X-Notion-Signature")
		l.Debug(fmt.Sprintf("signature header: %s", signature))
		if signature == "" {
			l.Error(errors.New("missing signature"), "no X-Notion-Signature header in webhook request")
			c.JSON(http.StatusUnauthorized, view.CreateResponse[any](nil, nil, errors.New("missing signature"), nil, ""))
			return
		}

		if !h.verifyNotionWebhookSignature(body, signature, verificationToken) {
			l.Error(errors.New("invalid signature"), "webhook signature verification failed")
			c.JSON(http.StatusUnauthorized, view.CreateResponse[any](nil, nil, errors.New("invalid signature"), nil, ""))
			return
		}
		l.Debug("webhook signature verified successfully")
	}

	// Validate payload has entity with page ID
	if payload.Entity == nil || payload.Entity.ID == "" {
		l.Error(errors.New("missing entity.id"), "no entity.id in webhook payload")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errors.New("missing entity.id"), nil, ""))
		return
	}

	// Only process page events
	if payload.Entity.Type != "page" {
		l.Debug(fmt.Sprintf("ignoring non-page entity type: %s", payload.Entity.Type))
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ignored"))
		return
	}

	pageID := payload.Entity.ID
	l.Debug(fmt.Sprintf("parsed webhook payload: type=%s page_id=%s", payload.Type, pageID))

	// Create leave service
	leaveService := notion.NewLeaveService(h.config, h.store, h.repo, h.logger)
	if leaveService == nil {
		l.Error(errors.New("failed to create leave service"), "notion leave service not configured")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, errors.New("service not configured"), nil, ""))
		return
	}

	// Fetch leave request from Notion
	ctx := c.Request.Context()
	leave, err := leaveService.GetLeaveRequest(ctx, pageID)
	if err != nil {
		l.Error(err, fmt.Sprintf("failed to fetch leave request from Notion: page_id=%s", pageID))
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("fetched leave request: page_id=%s status=%s email=%s", pageID, leave.Status, leave.Email))

	// Determine event type based on webhook type and current status
	var eventType NotionLeaveEventType
	switch payload.Type {
	case "page.created":
		eventType = NotionLeaveEventCreated
	case "page.updated", "page.content_updated", "page.properties_updated":
		// Determine event type based on current status
		switch leave.Status {
		case "Approved":
			eventType = NotionLeaveEventApproved
		case "Rejected":
			eventType = NotionLeaveEventRejected
		default:
			l.Debug(fmt.Sprintf("ignoring update event with status: %s", leave.Status))
			c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ignored"))
			return
		}
	default:
		l.Debug(fmt.Sprintf("ignoring unknown webhook type: %s", payload.Type))
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ignored"))
		return
	}

	l.Debug(fmt.Sprintf("processing leave event: type=%s email=%s", eventType, leave.Email))

	// Route to appropriate handler
	switch eventType {
	case NotionLeaveEventCreated:
		h.handleNotionLeaveCreated(c, l, leaveService, leave)
	case NotionLeaveEventApproved:
		h.handleNotionLeaveApproved(c, l, leaveService, leave)
	case NotionLeaveEventRejected:
		h.handleNotionLeaveRejected(c, l, leaveService, leave)
	default:
		l.Debug(fmt.Sprintf("ignoring leave event type: %s", eventType))
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ignored"))
	}
}

// handleNotionLeaveCreated validates a new leave request and sends Discord notification
func (h *handler) handleNotionLeaveCreated(c *gin.Context, l logger.Logger, leaveService *notion.LeaveService, leave *notion.LeaveRequest) {
	l.Debug(fmt.Sprintf("validating leave request: page_id=%s email=%s start_date=%v end_date=%v",
		leave.PageID, leave.Email, leave.StartDate, leave.EndDate))

	// Validate employee exists
	employee, err := h.store.Employee.OneByEmail(h.repo.DB(), leave.Email)
	if err != nil {
		l.Error(err, fmt.Sprintf("employee not found: email=%s", leave.Email))
		h.sendNotionLeaveDiscordNotification(c.Request.Context(),
			"❌ Leave Request Validation Failed",
			"Employee not found in database",
			15158332, // Red color
			[]model.DiscordMessageField{
				{Name: "Email", Value: leave.Email, Inline: nil},
			},
		)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validation_failed:employee_not_found"))
		return
	}

	// Validate dates
	if leave.StartDate == nil {
		l.Error(errors.New("missing start date"), "start date is required")
		h.sendNotionLeaveDiscordNotification(c.Request.Context(),
			"❌ Leave Request Validation Failed",
			"Start date is required",
			15158332,
			[]model.DiscordMessageField{
				{Name: "Employee", Value: employee.FullName, Inline: nil},
			},
		)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validation_failed:missing_start_date"))
		return
	}

	if leave.EndDate == nil {
		l.Error(errors.New("missing end date"), "end date is required")
		h.sendNotionLeaveDiscordNotification(c.Request.Context(),
			"❌ Leave Request Validation Failed",
			"End date is required",
			15158332,
			[]model.DiscordMessageField{
				{Name: "Employee", Value: employee.FullName, Inline: nil},
			},
		)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validation_failed:missing_end_date"))
		return
	}

	// Validate date range (end >= start)
	if leave.EndDate.Before(*leave.StartDate) {
		l.Debug(fmt.Sprintf("end date before start date: start_date=%v end_date=%v", leave.StartDate, leave.EndDate))
		h.sendNotionLeaveDiscordNotification(c.Request.Context(),
			"❌ Leave Request Validation Failed",
			"End date must be after start date",
			15158332,
			[]model.DiscordMessageField{
				{Name: "Employee", Value: employee.FullName, Inline: nil},
			},
		)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validation_failed:invalid_date_range"))
		return
	}

	// Send Discord message with buttons to onleave channel
	channelID := h.config.Discord.IDs.OnLeaveChannel
	if channelID == "" {
		l.Debug("onleave channel not configured, falling back to auditlog webhook")
		// Fallback to auditlog webhook
		inlineTrue := true
		h.sendNotionLeaveDiscordNotification(c.Request.Context(),
			"📋 New Leave Request - Pending Approval",
			fmt.Sprintf("[View in Notion](https://notion.so/%s)", strings.ReplaceAll(leave.PageID, "-", "")),
			3447003, // Blue color
			[]model.DiscordMessageField{
				{Name: "Employee", Value: fmt.Sprintf("%s (%s)", employee.FullName, leave.Email), Inline: nil},
				{Name: "Type", Value: leave.LeaveType, Inline: &inlineTrue},
				{Name: "Shift", Value: leave.Shift, Inline: &inlineTrue},
				{Name: "Dates", Value: fmt.Sprintf("%s to %s", leave.StartDate.Format("2006-01-02"), leave.EndDate.Format("2006-01-02")), Inline: nil},
				{Name: "Reason", Value: leave.Reason, Inline: nil},
			},
		)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validated"))
		return
	}

	// Get AM/DL mentions from deployments
	mentions := h.getAMDLMentionsFromDeployments(c.Request.Context(), l, leaveService, leave.Email)

	var assigneeMentions string
	if len(mentions) == 0 {
		l.Info(fmt.Sprintf("no AM/DL mentions found for leave request: email=%s", leave.Email))
		// Still send notification without mentions
	} else {
		assigneeMentions = fmt.Sprintf("🔔 **Assignees:** %s", strings.Join(mentions, " "))
	}

	// Build embed
	leaveType := leave.LeaveType
	if leaveType == "" {
		leaveType = "Annual Leave"
		l.Debug("leave type is empty, using default: Annual Leave")
	}

	embed := &discordgo.MessageEmbed{
		Title:       "📋 New Leave Request - Pending Approval",
		Description: fmt.Sprintf("[View in Notion](https://notion.so/%s)", strings.ReplaceAll(leave.PageID, "-", "")),
		Color:       3447003, // Blue color
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Employee", Value: fmt.Sprintf("%s (%s)", employee.FullName, leave.Email), Inline: false},
			{Name: "Type", Value: leaveType, Inline: false},
			{Name: "Dates", Value: fmt.Sprintf("%s to %s", leave.StartDate.Format("2006-01-02"), leave.EndDate.Format("2006-01-02")), Inline: false},
			{Name: "Reason", Value: leave.Reason, Inline: false},
		},
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	// Build buttons
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Approve",
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("notion_leave_approve_%s", leave.PageID),
					Emoji: discordgo.ComponentEmoji{
						Name: "✅",
					},
				},
				discordgo.Button{
					Label:    "Reject",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("notion_leave_reject_%s", leave.PageID),
					Emoji: discordgo.ComponentEmoji{
						Name: "❌",
					},
				},
			},
		},
	}

	// Send message with assignee mentions as content
	msg, err := h.service.Discord.SendChannelMessageComplex(channelID, assigneeMentions, []*discordgo.MessageEmbed{embed}, components)
	if err != nil {
		l.Error(err, "failed to send leave request message to discord channel")
		// Fallback to auditlog
		inlineTrue := true
		h.sendNotionLeaveDiscordNotification(c.Request.Context(),
			"📋 New Leave Request - Pending Approval",
			fmt.Sprintf("[View in Notion](https://notion.so/%s)", strings.ReplaceAll(leave.PageID, "-", "")),
			3447003,
			[]model.DiscordMessageField{
				{Name: "Employee", Value: fmt.Sprintf("%s (%s)", employee.FullName, leave.Email), Inline: &inlineTrue},
			},
		)
	} else {
		l.Debug(fmt.Sprintf("sent leave request message to discord channel: message_id=%s", msg.ID))
	}

	l.Debug(fmt.Sprintf("leave request validated successfully: employee_id=%s page_id=%s", employee.ID, leave.PageID))

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validated"))
}

// handleNotionLeaveApproved handles approval of a leave request
func (h *handler) handleNotionLeaveApproved(c *gin.Context, l logger.Logger, leaveService *notion.LeaveService, leave *notion.LeaveRequest) {
	l.Debug(fmt.Sprintf("approving leave request: page_id=%s email=%s",
		leave.PageID, leave.Email))

	// Lookup employee by email
	employee, err := h.store.Employee.OneByEmail(h.repo.DB(), leave.Email)
	if err != nil {
		l.Error(err, fmt.Sprintf("employee not found: email=%s", leave.Email))
		h.sendNotionLeaveDiscordNotification(c.Request.Context(),
			"❌ Leave Approval Failed",
			"Employee not found in database",
			15158332, // Red color
			[]model.DiscordMessageField{
				{Name: "Email", Value: leave.Email, Inline: nil},
			},
		)
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errors.New("employee_not_found"), nil, ""))
		return
	}

	// Lookup approver
	var approverID model.UUID
	if leave.ApprovedByID != "" {
		// Try to find approver's email from Notion contractor page
		approverEmail, err := h.getNotionContractorEmail(c.Request.Context(), leaveService, leave.ApprovedByID)
		if err == nil && approverEmail != "" {
			approver, err := h.store.Employee.OneByEmail(h.repo.DB(), approverEmail)
			if err == nil {
				approverID = approver.ID
			}
		}
	}
	if approverID.String() == "" {
		approverID = employee.ID
	}

	// Check if leave request already exists (prevent duplicates on re-approval)
	existingLeave, err := h.store.OnLeaveRequest.GetByNotionPageID(h.repo.DB(), leave.PageID)
	if err == nil && existingLeave != nil {
		l.Debug(fmt.Sprintf("leave request already approved (skipping duplicate): page_id=%s db_id=%s", leave.PageID, existingLeave.ID))
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, fmt.Sprintf("already_approved:%s", existingLeave.ID)))
		return
	}

	// Generate title
	startDateFormatted := leave.StartDate.Format("2006/01/02")
	endDateFormatted := leave.EndDate.Format("2006/01/02")
	title := fmt.Sprintf("%s | %s | %s - %s",
		employee.FullName,
		leave.LeaveType,
		startDateFormatted,
		endDateFormatted,
	)
	if leave.Shift != "" {
		title += " | " + leave.Shift
	}

	// Map leave type
	leaveType := mapNotionLeaveType(leave.LeaveType)

	// Create new on_leave_request record
	pageID := leave.PageID
	leaveRequest := &model.OnLeaveRequest{
		Type:         leaveType,
		StartDate:    leave.StartDate,
		EndDate:      leave.EndDate,
		Shift:        leave.Shift,
		Title:        title,
		Description:  leave.Reason,
		CreatorID:    employee.ID,
		ApproverID:   approverID,
		AssigneeIDs:  model.JSONArrayString{employee.ID.String()},
		NotionPageID: &pageID,
	}

	_, err = h.store.OnLeaveRequest.Create(h.repo.DB(), leaveRequest)
	if err != nil {
		l.Error(err, "failed to create leave request")
		h.sendNotionLeaveDiscordNotification(c.Request.Context(),
			"❌ Leave Approval Failed",
			"Database error",
			15158332, // Red color
			[]model.DiscordMessageField{
				{Name: "Employee", Value: employee.FullName, Inline: nil},
			},
		)
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, errors.New("failed_to_create_record"), nil, ""))
		return
	}

	// Send Discord notification
	inlineTrue := true
	h.sendNotionLeaveDiscordNotification(c.Request.Context(),
		"✅ Leave Request Approved",
		"",
		3066993, // Green color
		[]model.DiscordMessageField{
			{Name: "Employee", Value: employee.FullName, Inline: nil},
			{Name: "Type", Value: leave.LeaveType, Inline: &inlineTrue},
			{Name: "Shift", Value: leave.Shift, Inline: &inlineTrue},
			{Name: "Dates", Value: fmt.Sprintf("%s to %s", leave.StartDate.Format("2006-01-02"), leave.EndDate.Format("2006-01-02")), Inline: nil},
		},
	)

	l.Debug(fmt.Sprintf("leave request approved and persisted: id=%s employee_id=%s page_id=%s",
		leaveRequest.ID, employee.ID, leave.PageID))

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, fmt.Sprintf("approved:%s", leaveRequest.ID)))
}

// handleNotionLeaveRejected handles rejection of a leave request
func (h *handler) handleNotionLeaveRejected(c *gin.Context, l logger.Logger, leaveService *notion.LeaveService, leave *notion.LeaveRequest) {
	l.Debug(fmt.Sprintf("rejecting leave request: page_id=%s email=%s",
		leave.PageID, leave.Email))

	// Lookup employee by email
	employee, err := h.store.Employee.OneByEmail(h.repo.DB(), leave.Email)
	if err != nil {
		l.Debug(fmt.Sprintf("employee not found for rejection notification: email=%s", leave.Email))
	}

	// Check if this record was previously approved and exists in DB
	existingLeave, err := h.store.OnLeaveRequest.GetByNotionPageID(h.repo.DB(), leave.PageID)
	if err == nil && existingLeave != nil {
		// Record exists in DB - delete it since it's now rejected
		l.Debug(fmt.Sprintf("deleting previously approved leave request from DB: page_id=%s db_id=%s", leave.PageID, existingLeave.ID))
		if err := h.store.OnLeaveRequest.Delete(h.repo.DB(), existingLeave.ID.String()); err != nil {
			l.Error(err, "failed to delete rejected leave request from DB")
		} else {
			l.Debug(fmt.Sprintf("deleted previously approved leave request: page_id=%s db_id=%s", leave.PageID, existingLeave.ID))
		}
	}

	// Send Discord notification
	inlineTrue := true
	employeeName := leave.Email
	if employee != nil {
		employeeName = employee.FullName
	}
	h.sendNotionLeaveDiscordNotification(c.Request.Context(),
		"❌ Leave Request Rejected",
		"",
		15158332, // Red color
		[]model.DiscordMessageField{
			{Name: "Employee", Value: employeeName, Inline: nil},
			{Name: "Type", Value: leave.LeaveType, Inline: &inlineTrue},
			{Name: "Dates", Value: fmt.Sprintf("%s to %s", leave.StartDate.Format("2006-01-02"), leave.EndDate.Format("2006-01-02")), Inline: nil},
			{Name: "Reason", Value: leave.Reason, Inline: nil},
		},
	)

	l.Debug(fmt.Sprintf("leave request rejected: page_id=%s email=%s",
		leave.PageID, leave.Email))

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "rejected"))
}

// getNotionContractorEmail fetches the email from a Notion contractor page
func (h *handler) getNotionContractorEmail(ctx context.Context, leaveService *notion.LeaveService, contractorPageID string) (string, error) {
	// Fetch contractor page from Notion
	leave, err := leaveService.GetLeaveRequest(ctx, contractorPageID)
	if err != nil {
		return "", err
	}
	return leave.Email, nil
}

// getDiscordMentionFromUsername converts Discord username to mention format
// Returns empty string if username not found in database (graceful handling)
func (h *handler) getDiscordMentionFromUsername(
	l logger.Logger,
	discordUsername string,
) string {
	if discordUsername == "" {
		return ""
	}

	username := strings.TrimSpace(discordUsername)
	l.Debug(fmt.Sprintf("looking up Discord ID for username: %s", username))

	db := h.repo.DB()
	account, err := h.store.DiscordAccount.OneByUsername(db, username)
	if err != nil {
		l.Error(err, fmt.Sprintf("failed to lookup Discord account: username=%s", username))
		return ""
	}

	if account == nil {
		l.Info(fmt.Sprintf("Discord username not found in database: %s", username))
		return ""
	}

	mention := fmt.Sprintf("<@%s>", account.DiscordID)
	l.Debug(fmt.Sprintf("converted username to mention: %s -> %s", username, mention))

	return mention
}

// getAMDLMentionsFromDeployments gets Discord mentions for AM/DL from active deployments
// Returns array of Discord mentions (may be empty if no stakeholders found)
func (h *handler) getAMDLMentionsFromDeployments(
	ctx context.Context,
	l logger.Logger,
	leaveService *notion.LeaveService,
	teamEmail string,
) []string {
	var mentions []string
	var contractorPageID string

	// Step 1: Lookup contractor by email
	contractorID, err := leaveService.LookupContractorByEmail(ctx, teamEmail)
	if err != nil {
		l.Error(err, fmt.Sprintf("failed to lookup contractor: email=%s", teamEmail))
		// Don't return early - continue to fallback
	} else if contractorID == "" {
		l.Info(fmt.Sprintf("contractor not found for email: %s", teamEmail))
		// Don't return early - continue to fallback
	} else {
		contractorPageID = contractorID
		l.Debug(fmt.Sprintf("found contractor: email=%s contractor_id=%s", teamEmail, contractorPageID))

		// Step 2: Get active deployments
		deployments, err := leaveService.GetActiveDeploymentsForContractor(ctx, contractorPageID)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to get active deployments: contractor_id=%s", contractorPageID))
			// Don't return early - continue to fallback
		} else if len(deployments) == 0 {
			l.Info(fmt.Sprintf("no active deployments found: contractor_id=%s", contractorPageID))
			// Don't return early - continue to fallback
		} else {
			l.Debug(fmt.Sprintf("found %d active deployments for contractor: contractor_id=%s", len(deployments), contractorPageID))

			// Step 3: Extract stakeholders from all deployments (excluding the employee)
			stakeholderMap := make(map[string]bool)
			for i, deployment := range deployments {
				l.Debug(fmt.Sprintf("processing deployment %d/%d: deployment_id=%s", i+1, len(deployments), deployment.ID))

				stakeholders := leaveService.ExtractStakeholdersFromDeployment(ctx, deployment)
				l.Debug(fmt.Sprintf("extracted %d stakeholders from deployment: deployment_id=%s", len(stakeholders), deployment.ID))

				for _, stakeholderID := range stakeholders {
					// Filter out the employee themselves
					if stakeholderID == contractorPageID {
						l.Debug(fmt.Sprintf("skipping employee contractor from stakeholders: contractor_id=%s", stakeholderID))
						continue
					}
					stakeholderMap[stakeholderID] = true
				}
			}

			l.Debug(fmt.Sprintf("total unique stakeholders after filtering: %d", len(stakeholderMap)))

			// Step 4: Get Discord usernames from Notion
			for stakeholderID := range stakeholderMap {
				l.Debug(fmt.Sprintf("fetching Discord username for stakeholder: contractor_id=%s", stakeholderID))

				username, err := leaveService.GetDiscordUsernameFromContractor(ctx, stakeholderID)
				if err != nil {
					l.Error(err, fmt.Sprintf("failed to get Discord username: contractor_id=%s", stakeholderID))
					continue // Skip this stakeholder
				}
				if username == "" {
					l.Debug(fmt.Sprintf("Discord username not set for stakeholder: contractor_id=%s", stakeholderID))
					continue // Skip this stakeholder
				}

				l.Debug(fmt.Sprintf("found Discord username for stakeholder: contractor_id=%s username=%s", stakeholderID, username))

				// Step 5: Convert username to mention
				mention := h.getDiscordMentionFromUsername(l, username)
				if mention != "" {
					l.Debug(fmt.Sprintf("converted username to mention: username=%s mention=%s", username, mention))
					mentions = append(mentions, mention)
				} else {
					l.Debug(fmt.Sprintf("no Discord account found for username: %s", username))
				}
			}

			l.Debug(fmt.Sprintf("extracted %d Discord mentions from %d deployments (filtered employee)", len(mentions), len(deployments)))
		}
	}

	// Fallback: if no mentions found, assign to default assignees
	if len(mentions) == 0 {
		l.Info("no AM/DL mentions found, using fallback assignees: han@d.foundation, thanhpd@d.foundation")
		fallbackEmails := []string{"han@d.foundation", "thanhpd@d.foundation"}

		db := h.repo.DB()
		for _, fallbackEmail := range fallbackEmails {
			l.Debug(fmt.Sprintf("looking up fallback assignee in database: email=%s", fallbackEmail))

			// Lookup employee by email directly in database
			employee, err := h.store.Employee.OneByEmail(db, fallbackEmail)
			if err != nil {
				l.Error(err, fmt.Sprintf("failed to lookup fallback employee: email=%s", fallbackEmail))
				continue
			}
			if employee == nil {
				l.Debug(fmt.Sprintf("fallback employee not found in database: email=%s", fallbackEmail))
				continue
			}

			l.Debug(fmt.Sprintf("found fallback employee: email=%s employee_id=%s discord_account_id=%s", fallbackEmail, employee.ID, employee.DiscordAccountID))

			// Check if employee has Discord account
			if employee.DiscordAccountID.String() == "00000000-0000-0000-0000-000000000000" || employee.DiscordAccountID.String() == "" {
				l.Debug(fmt.Sprintf("employee has no Discord account ID: email=%s", fallbackEmail))
				continue
			}

			// Get Discord account
			discordAccount, err := h.store.DiscordAccount.One(db, employee.DiscordAccountID.String())
			if err != nil {
				l.Error(err, fmt.Sprintf("failed to get Discord account for fallback: email=%s discord_account_id=%s", fallbackEmail, employee.DiscordAccountID))
				continue
			}

			if discordAccount == nil {
				l.Debug(fmt.Sprintf("Discord account not found for fallback: email=%s", fallbackEmail))
				continue
			}

			l.Debug(fmt.Sprintf("found Discord account for fallback: email=%s discord_id=%s username=%s", fallbackEmail, discordAccount.DiscordID, discordAccount.DiscordUsername))

			// Build mention directly from Discord ID
			mention := fmt.Sprintf("<@%s>", discordAccount.DiscordID)
			l.Debug(fmt.Sprintf("created fallback mention: email=%s mention=%s", fallbackEmail, mention))
			mentions = append(mentions, mention)
		}

		l.Debug(fmt.Sprintf("added %d fallback mentions", len(mentions)))
	}

	return mentions
}

// sendNotionLeaveDiscordNotification sends an embed message to Discord auditlog channel
func (h *handler) sendNotionLeaveDiscordNotification(ctx context.Context, title, description string, color int64, fields []model.DiscordMessageField) {
	if h.service.Discord == nil {
		h.logger.Debug("discord service not configured, skipping notification")
		return
	}

	embed := model.DiscordMessageEmbed{
		Title:       title,
		Description: description,
		Color:       color,
		Fields:      fields,
		Timestamp:   time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	// Send to auditlog webhook
	_, err := h.service.Discord.SendMessage(model.DiscordMessage{
		Embeds: []model.DiscordMessageEmbed{embed},
	}, h.config.Discord.Webhooks.AuditLog)
	if err != nil {
		h.logger.Error(err, "failed to send discord notification to auditlog")
		return
	}

	h.logger.Debug(fmt.Sprintf("sent discord embed notification to auditlog: %s", title))
}

// mapNotionLeaveType maps Notion leave type to internal type
func mapNotionLeaveType(notionType string) string {
	switch notionType {
	case "Off":
		return "off"
	case "Remote":
		return "remote"
	default:
		return "off"
	}
}

// verifyNotionWebhookSignature verifies the HMAC-SHA256 signature of a Notion webhook request
func (h *handler) verifyNotionWebhookSignature(body []byte, signature, token string) bool {
	// Remove "sha256=" prefix if present
	cleanSignature := signature
	if len(signature) > 7 && signature[:7] == "sha256=" {
		cleanSignature = signature[7:]
	}

	// Calculate expected HMAC-SHA256 signature
	mac := hmac.New(sha256.New, []byte(token))
	mac.Write(body)
	expectedMAC := mac.Sum(nil)
	expectedSignature := hex.EncodeToString(expectedMAC)

	h.logger.Debug(fmt.Sprintf("signature verification: received=%s expected=%s token_len=%d", cleanSignature, expectedSignature, len(token)))

	// Compare signatures using constant-time comparison
	return hmac.Equal([]byte(expectedSignature), []byte(cleanSignature))
}

// NotionOnLeaveAutomationPayload represents the webhook payload from Notion automation for on-leave auto-fill
type NotionOnLeaveAutomationPayload struct {
	Source NotionAutomationSource     `json:"source"`
	Data   NotionOnLeaveAutomationData `json:"data"`
}

// NotionAutomationSource represents the source of Notion automation webhooks
type NotionAutomationSource struct {
	Type         string `json:"type"`
	AutomationID string `json:"automation_id"`
	ActionID     string `json:"action_id"`
	EventID      string `json:"event_id"`
	Attempt      int    `json:"attempt"`
}

// NotionOnLeaveAutomationData represents the page data in the automation webhook
type NotionOnLeaveAutomationData struct {
	Object     string                          `json:"object"`
	ID         string                          `json:"id"`
	Properties NotionOnLeaveAutomationProperties `json:"properties"`
	URL        string                          `json:"url"`
}

// NotionOnLeaveAutomationProperties represents the properties of the on-leave page for automation
type NotionOnLeaveAutomationProperties struct {
	Status     NotionAutomationStatusProperty   `json:"Status"`
	TeamEmail  NotionAutomationEmailProperty    `json:"Team Email"`
	Contractor NotionAutomationRelationProperty `json:"Contractor"`
}

// NotionAutomationStatusProperty represents a status property in automation payload
type NotionAutomationStatusProperty struct {
	ID     string                      `json:"id"`
	Type   string                      `json:"type"`
	Status *NotionAutomationSelectOption `json:"status"`
}

// NotionAutomationEmailProperty represents an email property in automation payload
type NotionAutomationEmailProperty struct {
	ID    string  `json:"id"`
	Type  string  `json:"type"`
	Email *string `json:"email"`
}

// NotionAutomationRelationProperty represents a relation property in automation payload
type NotionAutomationRelationProperty struct {
	ID       string                    `json:"id"`
	Type     string                    `json:"type"`
	Relation []NotionAutomationRelation `json:"relation"`
	HasMore  bool                      `json:"has_more"`
}

// NotionAutomationRelation represents a relation item in automation payload
type NotionAutomationRelation struct {
	ID string `json:"id"`
}

// NotionAutomationSelectOption represents a select option in automation payload
type NotionAutomationSelectOption struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// HandleNotionOnLeave handles on-leave webhook events from Notion automation
// This endpoint auto-fills the Employee relation based on Team Email
func (h *handler) HandleNotionOnLeave(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "webhook",
		"method":  "HandleNotionOnLeave",
	})

	l.Debug("received notion on-leave webhook request")

	// Log all headers for debugging
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	headersJSON, _ := json.Marshal(headers)
	l.Debug(fmt.Sprintf("request headers: %s", string(headersJSON)))

	// Read body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		l.Error(err, "failed to read notion on-leave webhook body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("received webhook body: %s", string(body)))

	// Parse payload
	var payload NotionOnLeaveAutomationPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		l.Error(err, "failed to parse notion on-leave webhook payload")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Get basic info from payload
	pageID := payload.Data.ID
	status := getAutomationStatusName(payload.Data.Properties.Status)
	email := getAutomationEmailValue(payload.Data.Properties.TeamEmail)

	l.Debug(fmt.Sprintf("parsed on-leave: page_id=%s status=%s email=%s", pageID, status, email))

	// Only process Pending status (new leave requests)
	if status != "Pending" {
		l.Debug(fmt.Sprintf("ignoring non-pending status: %s", status))
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ignored"))
		return
	}

	// Create leave service
	leaveService := notion.NewLeaveService(h.config, h.store, h.repo, h.logger)
	if leaveService == nil {
		l.Error(errors.New("failed to create leave service"), "notion leave service not configured")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, errors.New("service not configured"), nil, ""))
		return
	}

	// Fetch full leave request from Notion
	ctx := c.Request.Context()
	leave, err := leaveService.GetLeaveRequest(ctx, pageID)
	if err != nil {
		l.Error(err, fmt.Sprintf("failed to fetch leave request from Notion: page_id=%s", pageID))
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("fetched leave request: page_id=%s email=%s start_date=%v end_date=%v",
		pageID, leave.Email, leave.StartDate, leave.EndDate))

	// Update contractor relation if empty
	if len(payload.Data.Properties.Contractor.Relation) == 0 && email != "" {
		l.Debug(fmt.Sprintf("contractor relation is empty, looking up contractor by email: %s", email))

		contractorPageID, err := h.lookupContractorByEmailForOnLeave(ctx, l, email)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to lookup contractor by email: %s", email))
		} else if contractorPageID != "" {
			l.Debug(fmt.Sprintf("found contractor page: email=%s page_id=%s", email, contractorPageID))

			if err := h.updateOnLeaveContractor(ctx, l, pageID, contractorPageID); err != nil {
				l.Error(err, fmt.Sprintf("failed to update contractor relation: page_id=%s", pageID))
			} else {
				l.Debug(fmt.Sprintf("successfully updated contractor relation: onleave_page=%s contractor_page=%s", pageID, contractorPageID))
			}
		}
	} else {
		l.Debug(fmt.Sprintf("contractor relation already set: %v", payload.Data.Properties.Contractor.Relation))
	}

	// Auto-fill Leave Type to "Annual Leave" if empty
	if leave.LeaveType == "" {
		l.Debug("leave type is empty, auto-filling with Annual Leave")
		if err := h.updateLeaveType(ctx, l, pageID, "Annual Leave"); err != nil {
			l.Error(err, fmt.Sprintf("failed to update leave type: page_id=%s", pageID))
		} else {
			l.Debug(fmt.Sprintf("successfully updated leave type: page_id=%s type=Annual Leave", pageID))
			leave.LeaveType = "Annual Leave" // Update local object
		}
	} else {
		l.Debug(fmt.Sprintf("leave type already set: %s", leave.LeaveType))
	}

	// Validate employee exists
	employee, err := h.store.Employee.OneByEmail(h.repo.DB(), leave.Email)
	if err != nil {
		l.Error(err, fmt.Sprintf("employee not found: email=%s", leave.Email))
		h.sendNotionLeaveDiscordNotification(ctx,
			"❌ Leave Request Validation Failed",
			"Employee not found in database",
			15158332, // Red color
			[]model.DiscordMessageField{
				{Name: "Email", Value: leave.Email, Inline: nil},
			},
		)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validation_failed:employee_not_found"))
		return
	}

	// Validate dates
	if leave.StartDate == nil {
		l.Error(errors.New("missing start date"), "start date is required")
		h.sendNotionLeaveDiscordNotification(ctx,
			"❌ Leave Request Validation Failed",
			"Start date is required",
			15158332,
			[]model.DiscordMessageField{
				{Name: "Employee", Value: employee.FullName, Inline: nil},
			},
		)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validation_failed:missing_start_date"))
		return
	}

	if leave.EndDate == nil {
		l.Error(errors.New("missing end date"), "end date is required")
		h.sendNotionLeaveDiscordNotification(ctx,
			"❌ Leave Request Validation Failed",
			"End date is required",
			15158332,
			[]model.DiscordMessageField{
				{Name: "Employee", Value: employee.FullName, Inline: nil},
			},
		)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validation_failed:missing_end_date"))
		return
	}

	if leave.EndDate.Before(*leave.StartDate) {
		l.Debug(fmt.Sprintf("end date before start date: start_date=%v end_date=%v", leave.StartDate, leave.EndDate))
		h.sendNotionLeaveDiscordNotification(ctx,
			"❌ Leave Request Validation Failed",
			"End date must be after start date",
			15158332,
			[]model.DiscordMessageField{
				{Name: "Employee", Value: employee.FullName, Inline: nil},
			},
		)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validation_failed:invalid_date_range"))
		return
	}

	// Send Discord notification with AM/DL mentions
	channelID := h.config.Discord.IDs.OnLeaveChannel
	if channelID == "" {
		l.Debug("onleave channel not configured, falling back to auditlog webhook")

		// Use default Leave Type if empty
		fallbackLeaveType := leave.LeaveType
		if fallbackLeaveType == "" {
			fallbackLeaveType = "Annual Leave"
			l.Debug("leave type is empty in fallback notification, using default: Annual Leave")
		}

		h.sendNotionLeaveDiscordNotification(ctx,
			"📋 New Leave Request - Pending Approval",
			fmt.Sprintf("[View in Notion](https://notion.so/%s)", strings.ReplaceAll(leave.PageID, "-", "")),
			3447003, // Blue color
			[]model.DiscordMessageField{
				{Name: "Employee", Value: fmt.Sprintf("%s (%s)", employee.FullName, leave.Email), Inline: nil},
				{Name: "Type", Value: fallbackLeaveType, Inline: nil},
				{Name: "Dates", Value: fmt.Sprintf("%s to %s", leave.StartDate.Format("2006-01-02"), leave.EndDate.Format("2006-01-02")), Inline: nil},
				{Name: "Reason", Value: leave.Reason, Inline: nil},
			},
		)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validated"))
		return
	}

	// Get AM/DL mentions from deployments
	mentions := h.getAMDLMentionsFromDeployments(ctx, l, leaveService, leave.Email)

	var assigneeMentions string
	if len(mentions) == 0 {
		l.Info(fmt.Sprintf("no AM/DL mentions found for leave request: email=%s", leave.Email))
	} else {
		assigneeMentions = fmt.Sprintf("🔔 **Assignees:** %s", strings.Join(mentions, " "))
	}

	// Build embed
	leaveType := leave.LeaveType
	if leaveType == "" {
		leaveType = "Annual Leave"
		l.Debug("leave type is empty, using default: Annual Leave")
	}

	embed := &discordgo.MessageEmbed{
		Title:       "📋 New Leave Request - Pending Approval",
		Description: fmt.Sprintf("[View in Notion](https://notion.so/%s)", strings.ReplaceAll(leave.PageID, "-", "")),
		Color:       3447003, // Blue color
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Employee", Value: fmt.Sprintf("%s (%s)", employee.FullName, leave.Email), Inline: false},
			{Name: "Type", Value: leaveType, Inline: false},
			{Name: "Dates", Value: fmt.Sprintf("%s to %s", leave.StartDate.Format("2006-01-02"), leave.EndDate.Format("2006-01-02")), Inline: false},
			{Name: "Reason", Value: leave.Reason, Inline: false},
		},
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	// Build buttons
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Approve",
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("notion_leave_approve_%s", leave.PageID),
					Emoji: discordgo.ComponentEmoji{
						Name: "✅",
					},
				},
				discordgo.Button{
					Label:    "Reject",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("notion_leave_reject_%s", leave.PageID),
					Emoji: discordgo.ComponentEmoji{
						Name: "❌",
					},
				},
			},
		},
	}

	// Send message with assignee mentions as content
	msg, err := h.service.Discord.SendChannelMessageComplex(channelID, assigneeMentions, []*discordgo.MessageEmbed{embed}, components)
	if err != nil {
		l.Error(err, "failed to send leave request message to discord channel")
		// Fallback to auditlog
		inlineTrue := true
		h.sendNotionLeaveDiscordNotification(ctx,
			"📋 New Leave Request - Pending Approval",
			fmt.Sprintf("[View in Notion](https://notion.so/%s)", strings.ReplaceAll(leave.PageID, "-", "")),
			3447003,
			[]model.DiscordMessageField{
				{Name: "Employee", Value: fmt.Sprintf("%s (%s)", employee.FullName, leave.Email), Inline: &inlineTrue},
			},
		)
	} else {
		l.Debug(fmt.Sprintf("sent leave request message to discord channel: message_id=%s", msg.ID))
	}

	l.Debug(fmt.Sprintf("leave request validated successfully: employee_id=%s page_id=%s", employee.ID, leave.PageID))

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validated"))
}

// lookupContractorByEmailForOnLeave queries the contractor database to find a contractor by email
func (h *handler) lookupContractorByEmailForOnLeave(ctx context.Context, l logger.Logger, email string) (string, error) {
	contractorDBID := h.config.Notion.Databases.Contractor
	if contractorDBID == "" {
		return "", fmt.Errorf("NOTION_CONTRACTOR_DB_ID not configured")
	}

	l.Debug(fmt.Sprintf("querying contractor database: db_id=%s email=%s", contractorDBID, email))

	// Create Notion client
	client := nt.NewClient(h.config.Notion.Secret)

	// Query contractor database for matching email
	filter := &nt.DatabaseQueryFilter{
		Property: "Team Email",
		DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
			Email: &nt.TextPropertyFilter{
				Equals: email,
			},
		},
	}

	resp, err := client.QueryDatabase(ctx, contractorDBID, &nt.DatabaseQuery{
		Filter:   filter,
		PageSize: 1,
	})
	if err != nil {
		return "", fmt.Errorf("failed to query contractor database: %w", err)
	}

	if len(resp.Results) == 0 {
		l.Debug(fmt.Sprintf("no contractor found for email: %s", email))
		return "", nil
	}

	pageID := resp.Results[0].ID
	l.Debug(fmt.Sprintf("found contractor: email=%s page_id=%s", email, pageID))
	return pageID, nil
}

// updateOnLeaveContractor updates the Contractor relation field on the on-leave page
func (h *handler) updateOnLeaveContractor(ctx context.Context, l logger.Logger, onLeavePageID, contractorPageID string) error {
	l.Debug(fmt.Sprintf("updating on-leave contractor: onleave_page=%s contractor_page=%s", onLeavePageID, contractorPageID))

	// Create Notion client
	client := nt.NewClient(h.config.Notion.Secret)

	// Build update params with Contractor relation
	updateParams := nt.UpdatePageParams{
		DatabasePageProperties: nt.DatabasePageProperties{
			"Contractor": nt.DatabasePageProperty{
				Relation: []nt.Relation{
					{ID: contractorPageID},
				},
			},
		},
	}

	l.Debug(fmt.Sprintf("update params: %+v", updateParams))

	updatedPage, err := client.UpdatePage(ctx, onLeavePageID, updateParams)
	if err != nil {
		l.Error(err, fmt.Sprintf("notion API error updating page: %s", onLeavePageID))
		return fmt.Errorf("failed to update page: %w", err)
	}

	l.Debug(fmt.Sprintf("notion API response - page ID: %s, URL: %s", updatedPage.ID, updatedPage.URL))
	l.Debug(fmt.Sprintf("successfully updated contractor relation on on-leave page: %s", onLeavePageID))
	return nil
}

// updateLeaveType updates the Leave Type property on a leave request page
func (h *handler) updateLeaveType(ctx context.Context, l logger.Logger, leavePageID, leaveType string) error {
	l.Debug(fmt.Sprintf("updating leave type: leave_page=%s type=%s", leavePageID, leaveType))

	// Create Notion client
	client := nt.NewClient(h.config.Notion.Secret)

	// Build update params with Leave Type select
	updateParams := nt.UpdatePageParams{
		DatabasePageProperties: nt.DatabasePageProperties{
			"Leave Type": nt.DatabasePageProperty{
				Select: &nt.SelectOptions{
					Name: leaveType,
				},
			},
		},
	}

	l.Debug(fmt.Sprintf("update params: %+v", updateParams))

	updatedPage, err := client.UpdatePage(ctx, leavePageID, updateParams)
	if err != nil {
		l.Error(err, fmt.Sprintf("notion API error updating page: %s", leavePageID))
		return fmt.Errorf("failed to update page: %w", err)
	}

	l.Debug(fmt.Sprintf("notion API response - page ID: %s, URL: %s", updatedPage.ID, updatedPage.URL))
	l.Debug(fmt.Sprintf("successfully updated leave type on leave request page: %s", leavePageID))
	return nil
}

// Helper functions for automation payload

func getAutomationStatusName(prop NotionAutomationStatusProperty) string {
	if prop.Status != nil {
		return prop.Status.Name
	}
	return ""
}

func getAutomationEmailValue(prop NotionAutomationEmailProperty) string {
	if prop.Email != nil {
		return *prop.Email
	}
	return ""
}
