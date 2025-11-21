package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/nocodb"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

// LeaveEventType represents different leave webhook events
type LeaveEventType string

const (
	LeaveEventValidate LeaveEventType = "validate"
	LeaveEventApprove  LeaveEventType = "approve"
	LeaveEventReject   LeaveEventType = "reject"
)

// NocodbLeaveWebhookPayload represents the webhook payload from NocoDB
type NocodbLeaveWebhookPayload struct {
	Type   string `json:"type"`
	ID     string `json:"id"`
	BaseID string `json:"base_id"`
	Data   struct {
		TableName    string              `json:"table_name"`
		TableID      string              `json:"table_id"`
		Rows         []NocodbLeaveRecord `json:"rows"`
		PreviousRows []NocodbLeaveRecord `json:"previous_rows"`
	} `json:"data"`
}

type NocodbLeaveRecord struct {
	ID            int                  `json:"Id"`
	EmployeeEmail string               `json:"employee_email"`
	Type          string               `json:"type"`
	StartDate     string               `json:"start_date"`
	EndDate       string               `json:"end_date"`
	Shift         string               `json:"shift"`
	Reason        string               `json:"reason"`
	Status        string               `json:"status"`
	ApprovedBy    string               `json:"approved_by"`
	ApprovedAt    *string              `json:"approved_at"` // NocoDB sends string or null
	AssigneeLinks []NocodbEmployeeLink `json:"_nc_m2m_leave_requests_employees"`
	CreatedAt     string               `json:"CreatedAt"` // NocoDB format: "2025-11-19 06:04:22+00:00"
	UpdatedAt     string               `json:"UpdatedAt"` // NocoDB format: "2025-11-19 06:04:22+00:00"
}

type NocodbEmployeeLink struct {
	LeaveRequestsID int `json:"leave_requests_id"`
	EmployeesID     int `json:"employees_id"`
}

// HandleNocodbLeave handles all leave request webhook events
func (h *handler) HandleNocodbLeave(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "webhook",
		"method":  "HandleNocodbLeave",
	})

	l.Debugf("received nocodb leave webhook request")

	secret := h.config.LeaveIntegration.Noco.WebhookSecret
	if secret == "" {
		l.Error(errors.New("missing nocodb leave webhook secret"), "cannot verify leave webhook")
		c.JSON(http.StatusUnauthorized, view.CreateResponse[any](nil, nil, errors.New("signature verification disabled"), nil, ""))
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		l.Error(err, "failed to read nocodb leave webhook body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Debugf("received webhook body: %s", string(body))

	signature := c.GetHeader("X-NocoDB-Signature")
	authHeader := c.GetHeader("Authorization")
	if !verifyNocoSignature(secret, signature, authHeader, body) {
		l.Error(errors.New("invalid signature"), "nocodb leave signature mismatch")
		c.JSON(http.StatusUnauthorized, view.CreateResponse[any](nil, nil, errors.New("invalid signature"), nil, ""))
		return
	}

	var payload NocodbLeaveWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		l.Errorf(err, "failed to parse nocodb leave webhook payload")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Validate payload has data
	if len(payload.Data.Rows) == 0 {
		l.Error(errors.New("empty rows array"), "no data in webhook payload")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errors.New("empty rows"), nil, ""))
		return
	}

	record := payload.Data.Rows[0]
	var oldRecord *NocodbLeaveRecord
	if len(payload.Data.PreviousRows) > 0 {
		oldRecord = &payload.Data.PreviousRows[0]
	}

	l.Debugf("parsed webhook payload: type=%s table=%s record_id=%d status=%s old_status=%s",
		payload.Type, payload.Data.TableName, record.ID, record.Status,
		func() string {
			if oldRecord != nil {
				return oldRecord.Status
			}
			return "nil"
		}())

	// Determine event type based on webhook type and status
	var eventType LeaveEventType
	switch payload.Type {
	case "records.after.insert":
		eventType = LeaveEventValidate
	case "records.after.update":
		if oldRecord != nil && oldRecord.Status != "Approved" && record.Status == "Approved" {
			eventType = LeaveEventApprove
		} else if oldRecord != nil && oldRecord.Status != "Rejected" && record.Status == "Rejected" {
			eventType = LeaveEventReject
		} else {
			l.Infof("ignoring update event with no status transition: old=%s new=%s",
				func() string {
					if oldRecord != nil {
						return oldRecord.Status
					}
					return "nil"
				}(), record.Status)
			c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ignored"))
			return
		}
	default:
		l.Infof("ignoring unknown webhook type: %s", payload.Type)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ignored"))
		return
	}

	l.Debugf("processing leave event: type=%s employee=%s", eventType, record.EmployeeEmail)

	// Route to appropriate handler
	switch eventType {
	case LeaveEventValidate:
		h.handleLeaveValidation(c, l, &record)
	case LeaveEventApprove:
		h.handleLeaveApproval(c, l, &record)
	case LeaveEventReject:
		h.handleLeaveRejection(c, l, &record)
	default:
		l.Infof("ignoring leave event type: %s", eventType)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ignored"))
	}
}

func (h *handler) handleLeaveValidation(c *gin.Context, l logger.Logger, record *NocodbLeaveRecord) {
	l.Debugf("validating leave request: record_id=%d employee_email=%s start_date=%s end_date=%s",
		record.ID, record.EmployeeEmail, record.StartDate, record.EndDate)

	// Validate employee exists
	employee, err := h.store.Employee.OneByEmail(h.repo.DB(), record.EmployeeEmail)
	if err != nil {
		l.Errorf(err, "employee not found: email=%s", record.EmployeeEmail)
		h.sendLeaveDiscordNotification(c.Request.Context(),
			"❌ Leave Request Validation Failed",
			"Employee not found in database",
			15158332, // Red color
			[]model.DiscordMessageField{
				{Name: "Employee", Value: record.EmployeeEmail, Inline: nil},
			},
		)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validation_failed:employee_not_found"))
		return
	}

	// Validate date range
	startDate, err := time.Parse("2006-01-02", record.StartDate)
	if err != nil {
		l.Errorf(err, "invalid start date: start_date=%s", record.StartDate)
		h.sendLeaveDiscordNotification(c.Request.Context(),
			"❌ Leave Request Validation Failed",
			"Invalid start date format",
			15158332, // Red color
			[]model.DiscordMessageField{
				{Name: "Employee", Value: record.EmployeeEmail, Inline: nil},
			},
		)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validation_failed:invalid_start_date"))
		return
	}

	endDate, err := time.Parse("2006-01-02", record.EndDate)
	if err != nil {
		l.Errorf(err, "invalid end date: end_date=%s", record.EndDate)
		h.sendLeaveDiscordNotification(c.Request.Context(),
			"❌ Leave Request Validation Failed",
			"Invalid end date format",
			15158332, // Red color
			[]model.DiscordMessageField{
				{Name: "Employee", Value: record.EmployeeEmail, Inline: nil},
			},
		)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validation_failed:invalid_end_date"))
		return
	}

	// Validate date range (end >= start, start >= today)
	now := time.Now().Truncate(24 * time.Hour)
	if startDate.Before(now) {
		l.Debugf("start date in past: start_date=%v now=%v", startDate, now)
		h.sendLeaveDiscordNotification(c.Request.Context(),
			"❌ Leave Request Validation Failed",
			"Start date cannot be in the past",
			15158332, // Red color
			[]model.DiscordMessageField{
				{Name: "Employee", Value: record.EmployeeEmail, Inline: nil},
			},
		)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validation_failed:start_date_in_past"))
		return
	}

	if endDate.Before(startDate) {
		l.Debugf("end date before start date: start_date=%v end_date=%v", startDate, endDate)
		h.sendLeaveDiscordNotification(c.Request.Context(),
			"❌ Leave Request Validation Failed",
			"End date must be after start date",
			15158332, // Red color
			[]model.DiscordMessageField{
				{Name: "Employee", Value: record.EmployeeEmail, Inline: nil},
			},
		)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validation_failed:invalid_date_range"))
		return
	}

	// Fetch assignee emails from NocoDB
	leaveService := nocodb.NewLeaveService(h.service.NocoDB, h.config, h.store, h.repo, h.logger)
	assigneeEmails, err := leaveService.GetLeaveAssigneeEmails(record.ID)
	if err != nil {
		l.Warnf("failed to fetch assignee emails from nocodb: %v", err)
	}
	l.Debugf("fetched assignee emails: %v", assigneeEmails)

	// Get Discord mentions for assignees
	var mentions []string
	for _, email := range assigneeEmails {
		mention := h.getEmployeeDiscordMention(l, email)
		if mention != "" {
			mentions = append(mentions, mention)
		}
	}
	assigneeMentions := strings.Join(mentions, " ")
	l.Debugf("assignee mentions: %s", assigneeMentions)

	// Send Discord message with buttons to onleave channel
	channelID := h.config.Discord.IDs.OnLeaveChannel
	if channelID == "" {
		l.Warnf("onleave channel not configured, falling back to auditlog webhook")
		// Fallback to auditlog webhook
		nocodbURL := h.config.Noco.BaseURL
		inlineTrue := true
		h.sendLeaveDiscordNotification(c.Request.Context(),
			"📋 New Leave Request - Pending Approval",
			fmt.Sprintf("[View in NocoDB](%s)", nocodbURL),
			3447003, // Blue color
			[]model.DiscordMessageField{
				{Name: "Employee", Value: fmt.Sprintf("%s (%s)", employee.FullName, record.EmployeeEmail), Inline: nil},
				{Name: "Type", Value: record.Type, Inline: &inlineTrue},
				{Name: "Shift", Value: record.Shift, Inline: &inlineTrue},
				{Name: "Dates", Value: fmt.Sprintf("%s to %s", record.StartDate, record.EndDate), Inline: nil},
				{Name: "Reason", Value: record.Reason, Inline: nil},
			},
		)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validated"))
		return
	}

	// Build content with mentions
	content := ""
	if assigneeMentions != "" {
		content = fmt.Sprintf("🔔 **Assignees:** %s", assigneeMentions)
	}

	// Build embed
	nocodbURL := h.config.Noco.BaseURL
	embed := &discordgo.MessageEmbed{
		Title:       "📋 New Leave Request - Pending Approval",
		Description: fmt.Sprintf("[View in NocoDB](%s)", nocodbURL),
		Color:       3447003, // Blue color
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Employee", Value: fmt.Sprintf("%s (%s)", employee.FullName, record.EmployeeEmail), Inline: false},
			{Name: "Type", Value: record.Type, Inline: true},
			{Name: "Shift", Value: record.Shift, Inline: true},
			{Name: "Dates", Value: fmt.Sprintf("%s to %s", record.StartDate, record.EndDate), Inline: false},
			{Name: "Reason", Value: record.Reason, Inline: false},
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
					CustomID: fmt.Sprintf("leave_approve_%d", record.ID),
					Emoji: discordgo.ComponentEmoji{
						Name: "✅",
					},
				},
				discordgo.Button{
					Label:    "Reject",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("leave_reject_%d", record.ID),
					Emoji: discordgo.ComponentEmoji{
						Name: "❌",
					},
				},
			},
		},
	}

	// Send message
	msg, err := h.service.Discord.SendChannelMessageComplex(channelID, content, []*discordgo.MessageEmbed{embed}, components)
	if err != nil {
		l.Errorf(err, "failed to send leave request message to discord channel")
		// Fallback to auditlog
		h.sendLeaveDiscordNotification(c.Request.Context(),
			"📋 New Leave Request - Pending Approval",
			fmt.Sprintf("[View in NocoDB](%s)", nocodbURL),
			3447003,
			[]model.DiscordMessageField{
				{Name: "Employee", Value: fmt.Sprintf("%s (%s)", employee.FullName, record.EmployeeEmail), Inline: nil},
			},
		)
	} else {
		l.Debugf("sent leave request message to discord channel: message_id=%s", msg.ID)
	}

	l.Infof("leave request validated successfully: employee_id=%s record_id=%d", employee.ID, record.ID)

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validated"))
}

// getEmployeeDiscordMention returns Discord mention for an employee by email
func (h *handler) getEmployeeDiscordMention(l logger.Logger, email string) string {
	if email == "" {
		return ""
	}

	// Look up employee by email
	employee, err := h.store.Employee.OneByEmail(h.repo.DB(), email)
	if err != nil {
		l.Debugf("could not find employee by email %s: %v", email, err)
		return ""
	}

	// Get Discord account
	if employee.DiscordAccountID.String() == "" {
		l.Debugf("employee %s has no discord account linked", email)
		return ""
	}

	// Get Discord account to get Discord ID
	discordAccount, err := h.store.DiscordAccount.One(h.repo.DB(), employee.DiscordAccountID.String())
	if err != nil {
		l.Debugf("could not find discord account for employee %s: %v", email, err)
		return ""
	}

	if discordAccount.DiscordID == "" {
		l.Debugf("discord account for employee %s has no discord id", email)
		return ""
	}

	l.Debugf("found discord id %s for employee %s", discordAccount.DiscordID, email)
	return fmt.Sprintf("<@%s>", discordAccount.DiscordID)
}

func (h *handler) handleLeaveApproval(c *gin.Context, l logger.Logger, record *NocodbLeaveRecord) {
	l.Debugf("approving leave request: record_id=%d employee_email=%s approved_by=%s",
		record.ID, record.EmployeeEmail, record.ApprovedBy)

	// Lookup employee by email
	employee, err := h.store.Employee.OneByEmail(h.repo.DB(), record.EmployeeEmail)
	if err != nil {
		l.Errorf(err, "employee not found: email=%s", record.EmployeeEmail)
		h.sendLeaveDiscordNotification(c.Request.Context(),
			"❌ Leave Approval Failed",
			"Employee not found in database",
			15158332, // Red color
			[]model.DiscordMessageField{
				{Name: "Employee", Value: record.EmployeeEmail, Inline: nil},
			},
		)
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errors.New("employee_not_found"), nil, ""))
		return
	}

	// Lookup approver by email
	var approverID model.UUID
	if record.ApprovedBy != "" {
		approver, err := h.store.Employee.OneByEmail(h.repo.DB(), record.ApprovedBy)
		if err != nil {
			l.Warnf("approver not found, using creator as approver: approved_by=%s", record.ApprovedBy)
			approverID = employee.ID
		} else {
			approverID = approver.ID
		}
	} else {
		approverID = employee.ID
	}

	// Parse dates
	startDate, _ := time.Parse("2006-01-02", record.StartDate)
	endDate, _ := time.Parse("2006-01-02", record.EndDate)

	// Parse assignees from NocoDB linked records
	assigneeIDs := model.JSONArrayString{employee.ID.String()} // Include creator
	for _, link := range record.AssigneeLinks {
		l.Debugf("processing assignee link: employees_id=%d", link.EmployeesID)
		// Note: AssigneeLinks contain NocoDB employee IDs, would need lookup if using fortress_id
		// For now, just store the employee ID as string
		assigneeIDs = append(assigneeIDs, fmt.Sprintf("%d", link.EmployeesID))
	}

	// Generate title from record data with YYYY/MM/DD format
	startDateFormatted := startDate.Format("2006/01/02")
	endDateFormatted := endDate.Format("2006/01/02")
	title := fmt.Sprintf("%s | %s | %s - %s",
		employee.FullName,
		record.Type,
		startDateFormatted,
		endDateFormatted,
	)
	if record.Shift != "" {
		title += " | " + record.Shift
	}

	// Check if leave request already exists (prevent duplicates on re-approval)
	// Note: Hard-deleted records won't be found, so re-approval after rejection will create new record
	nocodbID := record.ID
	existingLeave, err := h.store.OnLeaveRequest.GetByNocodbID(h.repo.DB(), nocodbID)
	if err == nil && existingLeave != nil {
		// Record already exists - skip duplicate
		l.Infof("leave request already approved (skipping duplicate): nocodb_id=%d db_id=%s", nocodbID, existingLeave.ID)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, fmt.Sprintf("already_approved:%s", existingLeave.ID)))
		return
	}

	// Create new on_leave_request record
	leaveRequest := &model.OnLeaveRequest{
		Type:        record.Type,
		StartDate:   &startDate,
		EndDate:     &endDate,
		Shift:       record.Shift,
		Title:       title,
		Description: record.Reason,
		CreatorID:   employee.ID,
		ApproverID:  approverID,
		AssigneeIDs: assigneeIDs,
		NocodbID:    &nocodbID,
	}

	_, err = h.store.OnLeaveRequest.Create(h.repo.DB(), leaveRequest)
	if err != nil {
		l.Errorf(err, "failed to create leave request")
		h.sendLeaveDiscordNotification(c.Request.Context(),
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
	h.sendLeaveDiscordNotification(c.Request.Context(),
		"✅ Leave Request Approved",
		"",
		3066993, // Green color
		[]model.DiscordMessageField{
			{Name: "Employee", Value: employee.FullName, Inline: nil},
			{Name: "Type", Value: record.Type, Inline: &inlineTrue},
			{Name: "Shift", Value: record.Shift, Inline: &inlineTrue},
			{Name: "Dates", Value: fmt.Sprintf("%s to %s", record.StartDate, record.EndDate), Inline: nil},
		},
	)

	l.Infof("leave request approved and persisted: id=%s employee_id=%s nocodb_id=%d",
		leaveRequest.ID, employee.ID, nocodbID)

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, fmt.Sprintf("approved:%s", leaveRequest.ID)))
}

func (h *handler) handleLeaveRejection(c *gin.Context, l logger.Logger, record *NocodbLeaveRecord) {
	l.Debugf("rejecting leave request: record_id=%d employee_email=%s",
		record.ID, record.EmployeeEmail)

	// Lookup employee by email
	employee, err := h.store.Employee.OneByEmail(h.repo.DB(), record.EmployeeEmail)
	if err != nil {
		l.Warnf("employee not found for rejection notification: email=%s", record.EmployeeEmail)
	}

	// Check if this record was previously approved and exists in DB
	nocodbID := record.ID
	existingLeave, err := h.store.OnLeaveRequest.GetByNocodbID(h.repo.DB(), nocodbID)
	if err == nil && existingLeave != nil {
		// Record exists in DB - delete it since it's now rejected
		l.Debugf("deleting previously approved leave request from DB: nocodb_id=%d db_id=%s", nocodbID, existingLeave.ID)
		if err := h.store.OnLeaveRequest.Delete(h.repo.DB(), existingLeave.ID.String()); err != nil {
			l.Errorf(err, "failed to delete rejected leave request from DB")
		} else {
			l.Infof("deleted previously approved leave request: nocodb_id=%d db_id=%s", nocodbID, existingLeave.ID)
		}
	}

	// Send Discord notification
	inlineTrue := true
	employeeName := record.EmployeeEmail
	if employee != nil {
		employeeName = employee.FullName
	}
	h.sendLeaveDiscordNotification(c.Request.Context(),
		"❌ Leave Request Rejected",
		"",
		15158332, // Red color
		[]model.DiscordMessageField{
			{Name: "Employee", Value: employeeName, Inline: nil},
			{Name: "Type", Value: record.Type, Inline: &inlineTrue},
			{Name: "Dates", Value: fmt.Sprintf("%s to %s", record.StartDate, record.EndDate), Inline: nil},
			{Name: "Reason", Value: record.Reason, Inline: nil},
		},
	)

	l.Infof("leave request rejected: record_id=%d employee_email=%s",
		record.ID, record.EmployeeEmail)

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "rejected"))
}

// sendLeaveDiscordNotification sends an embed message to Discord auditlog channel
func (h *handler) sendLeaveDiscordNotification(ctx context.Context, title, description string, color int64, fields []model.DiscordMessageField) {
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
		h.logger.Errorf(err, "failed to send discord notification to auditlog")
		return
	}

	h.logger.Debugf("sent discord embed notification to auditlog: %s", title)
}
