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

	// Build Discord mentions for assignees
	var assigneeMentions string
	if len(leave.Assignees) > 0 {
		l.Debug(fmt.Sprintf("found %d assignees for leave request: %v", len(leave.Assignees), leave.Assignees))
		var mentions []string
		for _, email := range leave.Assignees {
			mention := h.getEmployeeDiscordMention(l, email)
			if mention != "" {
				mentions = append(mentions, mention)
			}
		}
		if len(mentions) > 0 {
			assigneeMentions = fmt.Sprintf("🔔 **Assignees:** %s", strings.Join(mentions, " "))
		}
	}

	// Build embed
	embed := &discordgo.MessageEmbed{
		Title:       "📋 New Leave Request - Pending Approval",
		Description: fmt.Sprintf("[View in Notion](https://notion.so/%s)", strings.ReplaceAll(leave.PageID, "-", "")),
		Color:       3447003, // Blue color
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Employee", Value: fmt.Sprintf("%s (%s)", employee.FullName, leave.Email), Inline: false},
			{Name: "Type", Value: leave.LeaveType, Inline: true},
			{Name: "Shift", Value: leave.Shift, Inline: true},
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
	Status    NotionAutomationStatusProperty   `json:"Status"`
	TeamEmail NotionAutomationEmailProperty    `json:"Team Email"`
	Employee  NotionAutomationRelationProperty `json:"Employee"`
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

	// Log parsed data
	l.Debug(fmt.Sprintf("parsed on-leave: page_id=%s status=%s email=%v",
		payload.Data.ID,
		getAutomationStatusName(payload.Data.Properties.Status),
		getAutomationEmailValue(payload.Data.Properties.TeamEmail),
	))

	// Check if Employee relation is empty and we have a Team Email
	if len(payload.Data.Properties.Employee.Relation) == 0 {
		email := getAutomationEmailValue(payload.Data.Properties.TeamEmail)
		if email != "" {
			l.Debug(fmt.Sprintf("employee relation is empty, looking up contractor by email: %s", email))

			// Look up contractor page ID by email
			contractorPageID, err := h.lookupContractorByEmailForOnLeave(c.Request.Context(), l, email)
			if err != nil {
				l.Error(err, fmt.Sprintf("failed to lookup contractor by email: %s", email))
				// Continue without updating - don't fail the webhook
			} else if contractorPageID != "" {
				l.Debug(fmt.Sprintf("found contractor page: email=%s page_id=%s", email, contractorPageID))

				// Update the on-leave page with Employee relation
				if err := h.updateOnLeaveEmployee(c.Request.Context(), l, payload.Data.ID, contractorPageID); err != nil {
					l.Error(err, fmt.Sprintf("failed to update employee relation: page_id=%s", payload.Data.ID))
				} else {
					l.Debug(fmt.Sprintf("successfully updated employee relation: onleave_page=%s contractor_page=%s", payload.Data.ID, contractorPageID))
				}
			}
		} else {
			l.Debug("employee relation is empty but no team email provided")
		}
	} else {
		l.Debug(fmt.Sprintf("employee relation already set: %v", payload.Data.Properties.Employee.Relation))
	}

	// Return success to acknowledge receipt
	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "processed"))
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

// updateOnLeaveEmployee updates the Employee relation field on the on-leave page
func (h *handler) updateOnLeaveEmployee(ctx context.Context, l logger.Logger, onLeavePageID, contractorPageID string) error {
	l.Debug(fmt.Sprintf("updating on-leave employee: onleave_page=%s contractor_page=%s", onLeavePageID, contractorPageID))

	// Create Notion client
	client := nt.NewClient(h.config.Notion.Secret)

	// Build update params with Employee relation
	updateParams := nt.UpdatePageParams{
		DatabasePageProperties: nt.DatabasePageProperties{
			"Employee": nt.DatabasePageProperty{
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
	l.Debug(fmt.Sprintf("successfully updated employee relation on on-leave page: %s", onLeavePageID))
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
