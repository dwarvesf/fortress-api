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
	"strconv"
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

// isValidLeaveRequestID validates the format of a leave request ID
// Expected format: OOO-YYYY-USERNAME-CODE (e.g., "OOO-2026-nlk0211-GD5F")
func isValidLeaveRequestID(requestID string) bool {
	if requestID == "" {
		return false
	}

	// Split by dash separator
	parts := strings.Split(requestID, "-")

	// Should have exactly 4 parts: OOO, YYYY, USERNAME, CODE
	if len(parts) != 4 {
		return false
	}

	// Validate part 1: prefix should be "OOO"
	if parts[0] != "OOO" {
		return false
	}

	// Validate part 2: year should be 4 digits
	if len(parts[1]) != 4 {
		return false
	}
	year, err := strconv.Atoi(parts[1])
	if err != nil || year < 2020 || year > 2100 {
		return false
	}

	// Validate part 3: username should not be empty
	if parts[2] == "" {
		return false
	}

	// Validate part 4: code should not be empty
	if parts[3] == "" {
		return false
	}

	return true
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

// formatShortDateRange formats a date range in compact format
// Examples: "Jan 15-20, 2025" or "Jan 15, 2025 - Feb 2, 2025"
func formatShortDateRange(start, end time.Time) string {
	if start.Year() == end.Year() && start.Month() == end.Month() {
		// Same month and year: "Jan 15-20, 2025"
		return fmt.Sprintf("%s %d-%d, %d", start.Month().String()[:3], start.Day(), end.Day(), start.Year())
	}
	// Different months: "Jan 15, 2025 - Feb 2, 2025"
	return fmt.Sprintf("%s %d, %d - %s %d, %d",
		start.Month().String()[:3], start.Day(), start.Year(),
		end.Month().String()[:3], end.Day(), end.Year())
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
	Source NotionAutomationSource      `json:"source"`
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
	Object     string                            `json:"object"`
	ID         string                            `json:"id"`
	Properties NotionOnLeaveAutomationProperties `json:"properties"`
	URL        string                            `json:"url"`
}

// NotionOnLeaveAutomationProperties represents the properties of the on-leave page for automation
type NotionOnLeaveAutomationProperties struct {
	Status     NotionAutomationStatusProperty    `json:"Status"`
	TeamEmail  NotionAutomationEmailProperty     `json:"Team Email"`
	Contractor NotionAutomationRelationProperty  `json:"Contractor"`
	CreatedBy  NotionAutomationCreatedByProperty `json:"Created by"`
}

// NotionAutomationStatusProperty represents a status property in automation payload
type NotionAutomationStatusProperty struct {
	ID     string                        `json:"id"`
	Type   string                        `json:"type"`
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
	ID       string                     `json:"id"`
	Type     string                     `json:"type"`
	Relation []NotionAutomationRelation `json:"relation"`
	HasMore  bool                       `json:"has_more"`
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

// NotionAutomationCreatedByProperty represents a created_by property in automation payload
type NotionAutomationCreatedByProperty struct {
	ID        string                      `json:"id"`
	Type      string                      `json:"type"`
	CreatedBy *NotionAutomationUserObject `json:"created_by"`
}

// NotionAutomationUserObject represents a Notion user in automation payload
type NotionAutomationUserObject struct {
	Object string `json:"object"`
	ID     string `json:"id"`
}

// HandleNotionOnLeave handles on-leave webhook events from Notion automation
// This endpoint auto-fills the Employee relation based on Team Email
//
// Flow (restructured for reliability):
//  1. Parse payload, status gate
//  2. Fetch leave request (dates, basic info)
//  3. Lookup contractor by created_by user ID
//  4. Fetch contractor details (name, Discord, email)
//  5. Send notification (priority — must happen even if auto-fill fails)
//  6. Best-effort: auto-fill Contractor relation on Notion page
//  7. Best-effort: retry for Request ID (generated after Contractor is filled)
//  8. Best-effort: auto-fill Unavailability Type
//  9. If auto-fill fails, send error notification to fortress-logs
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
	createdByUserID := getAutomationCreatedByUserID(payload.Data.Properties.CreatedBy)

	l.Debug(fmt.Sprintf("parsed on-leave: page_id=%s status=%s created_by_user_id=%s", pageID, status, createdByUserID))

	// Only process New status (new leave requests)
	if status != "New" {
		l.Debug(fmt.Sprintf("ignoring non-new status: %s", status))
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

	// Use background context for Notion operations to avoid HTTP request timeout
	ctx := context.Background()

	// Step 1: Fetch leave request from Notion (dates, basic info — Request ID won't be ready yet)
	leave, err := leaveService.GetLeaveRequest(ctx, pageID)
	if err != nil {
		l.Error(err, fmt.Sprintf("failed to fetch leave request from Notion: page_id=%s", pageID))
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("fetched leave request: page_id=%s email=%s start_date=%v end_date=%v request_id=%s",
		pageID, leave.Email, leave.StartDate, leave.EndDate, leave.LeaveRequestTitle))

	// Step 2: Lookup contractor by created_by user ID
	var contractorPageID string
	if createdByUserID != "" {
		l.Debug(fmt.Sprintf("looking up contractor by user_id: %s", createdByUserID))

		contractorPageID, err = h.lookupContractorByUserIDForOnLeave(ctx, l, createdByUserID)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to lookup contractor by user_id: %s", createdByUserID))
		} else if contractorPageID != "" {
			leave.EmployeeID = contractorPageID
			l.Debug(fmt.Sprintf("found contractor page: user_id=%s page_id=%s", createdByUserID, contractorPageID))

			// Fetch email from contractor page
			email, err := h.getContractorEmail(ctx, l, contractorPageID)
			if err != nil {
				l.Error(err, fmt.Sprintf("failed to fetch contractor email: contractor_page_id=%s", contractorPageID))
			} else if email != "" {
				leave.Email = email
				l.Debug(fmt.Sprintf("fetched contractor email: contractor_page_id=%s email=%s", contractorPageID, email))
			}
		}
	}

	// Step 3: Fetch contractor details if we have the ID
	var contractor *notion.ContractorDetails
	var contractorDiscordMention string

	if leave.EmployeeID != "" {
		l.Debug(fmt.Sprintf("fetching contractor details from Notion: contractor_id=%s", leave.EmployeeID))
		contractor, err = leaveService.GetContractorDetails(ctx, leave.EmployeeID)
		if err != nil || contractor == nil {
			l.Error(err, fmt.Sprintf("failed to fetch contractor details: contractor_id=%s", leave.EmployeeID))
		} else if contractor.Status != "Active" {
			l.Debug(fmt.Sprintf("contractor is not active: contractor_id=%s status=%s", leave.EmployeeID, contractor.Status))
			h.sendNotionLeaveDiscordNotification(ctx,
				"❌ Leave Request Validation Failed",
				fmt.Sprintf("Contractor is not active (status: %s)", contractor.Status),
				15158332,
				[]model.DiscordMessageField{
					{Name: "Contractor", Value: contractor.FullName, Inline: nil},
				},
			)
			c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validation_failed:contractor_inactive"))
			return
		} else {
			l.Debug(fmt.Sprintf("found contractor: id=%s full_name=%s email=%s discord=%s",
				leave.EmployeeID, contractor.FullName, contractor.TeamEmail, contractor.DiscordUsername))
			contractorDiscordMention = h.getDiscordMentionFromUsername(l, contractor.DiscordUsername)
		}
	}

	// Validate dates
	if leave.StartDate == nil || leave.EndDate == nil {
		errMsg := "missing start or end date"
		l.Error(errors.New(errMsg), errMsg)
		contractorName := "Unknown"
		if contractor != nil {
			contractorName = contractor.FullName
		}
		h.sendNotionLeaveDiscordNotification(ctx,
			"❌ Leave Request Validation Failed",
			errMsg,
			15158332,
			[]model.DiscordMessageField{
				{Name: "Contractor", Value: contractorName, Inline: nil},
				{Name: "Page ID", Value: pageID, Inline: nil},
			},
		)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validation_failed:missing_dates"))
		return
	}

	if leave.EndDate.Before(*leave.StartDate) {
		contractorName := "Unknown"
		if contractor != nil {
			contractorName = contractor.FullName
		}
		h.sendNotionLeaveDiscordNotification(ctx,
			"❌ Leave Request Validation Failed",
			"End date must be after start date",
			15158332,
			[]model.DiscordMessageField{
				{Name: "Contractor", Value: contractorName, Inline: nil},
			},
		)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validation_failed:invalid_date_range"))
		return
	}

	// Step 4: Send notification (priority path — must happen even if auto-fill fails later)
	if contractor != nil {
		h.sendLeaveNotification(ctx, l, leaveService, leave, contractor, contractorDiscordMention)
	} else {
		// Contractor lookup failed — we can't send a proper leave notification
		l.Error(errors.New("contractor not found"), fmt.Sprintf("cannot send leave notification: page_id=%s created_by=%s", pageID, createdByUserID))
	}

	// Respond to webhook immediately — remaining steps are best-effort
	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "validated"))

	// Step 5: Best-effort auto-fill Contractor relation on Notion page
	autoFillFailed := false
	var autoFillError string

	if contractorPageID != "" {
		if err := h.updateOnLeaveContractor(ctx, l, pageID, contractorPageID); err != nil {
			l.Error(err, fmt.Sprintf("failed to update contractor relation: page_id=%s", pageID))
			autoFillFailed = true
			autoFillError = fmt.Sprintf("contractor relation update failed: %v", err)
		} else {
			l.Debug(fmt.Sprintf("successfully updated contractor relation: onleave_page=%s contractor_page=%s", pageID, contractorPageID))

			// Step 6: Best-effort retry for Request ID (generated after Contractor is filled)
			maxRetries := 3
			baseDelay := 2 * time.Second
			for i := 0; i < maxRetries && !isValidLeaveRequestID(leave.LeaveRequestTitle); i++ {
				retryDelay := baseDelay * (1 << i) // 2s, 4s, 8s
				l.Debug(fmt.Sprintf("waiting for request ID (attempt %d/%d): '%s', waiting %v",
					i+1, maxRetries, leave.LeaveRequestTitle, retryDelay))
				time.Sleep(retryDelay)

				leave, err = leaveService.GetLeaveRequest(ctx, pageID)
				if err != nil {
					l.Error(err, fmt.Sprintf("failed to fetch leave request on retry: page_id=%s", pageID))
					break
				}
				l.Debug(fmt.Sprintf("retry %d: fetched request_id='%s'", i+1, leave.LeaveRequestTitle))
			}

			if !isValidLeaveRequestID(leave.LeaveRequestTitle) {
				l.Warn(fmt.Sprintf("leave request ID still invalid after %d retries: page_id=%s request_id='%s'",
					maxRetries, pageID, leave.LeaveRequestTitle))
			}
		}
	} else {
		autoFillFailed = true
		autoFillError = fmt.Sprintf("contractor lookup failed for user_id=%s", createdByUserID)
	}

	// Step 7: Best-effort auto-fill Unavailability Type
	if leave.UnavailabilityType == "" {
		l.Debug("unavailability type is empty, auto-filling with Personal Time")
		if err := h.updateUnavailabilityType(ctx, l, pageID, "Personal Time"); err != nil {
			l.Error(err, fmt.Sprintf("failed to update unavailability type: page_id=%s", pageID))
		} else {
			l.Debug(fmt.Sprintf("successfully updated unavailability type: page_id=%s type=Personal Time", pageID))
		}
	}

	// Step 8: Send error notification to fortress-logs if auto-fill failed
	if autoFillFailed {
		h.sendOnLeaveAutoFillErrorNotification(pageID, createdByUserID, autoFillError)
	}

	l.Debug(fmt.Sprintf("leave request processing complete: page_id=%s contractor_id=%s auto_fill_failed=%v",
		pageID, leave.EmployeeID, autoFillFailed))
}

// sendLeaveNotification sends the leave request notification to Discord
func (h *handler) sendLeaveNotification(
	ctx context.Context,
	l logger.Logger,
	leaveService *notion.LeaveService,
	leave *notion.LeaveRequest,
	contractor *notion.ContractorDetails,
	contractorDiscordMention string,
) {
	channelID := h.config.Discord.IDs.OnLeaveChannel
	if channelID == "" {
		l.Debug("onleave channel not configured, falling back to auditlog webhook")

		description := fmt.Sprintf("%s request time off", contractor.FullName)
		if contractorDiscordMention != "" {
			description = fmt.Sprintf("%s request time off", contractorDiscordMention)
		}
		inlineTrue := true
		h.sendNotionLeaveDiscordNotification(ctx,
			"📋 Leave Request",
			description,
			3447003,
			[]model.DiscordMessageField{
				{Name: "Request", Value: leave.LeaveRequestTitle, Inline: &inlineTrue},
				{Name: "Type", Value: leave.UnavailabilityType, Inline: &inlineTrue},
				{Name: "Dates", Value: formatShortDateRange(*leave.StartDate, *leave.EndDate), Inline: nil},
				{Name: "Details", Value: leave.AdditionalContext, Inline: nil},
			},
		)
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

	description := fmt.Sprintf("%s request time off", contractor.FullName)
	if contractorDiscordMention != "" {
		description = fmt.Sprintf("%s request time off", contractorDiscordMention)
	}

	embed := &discordgo.MessageEmbed{
		Title:       "📋 Leave Request",
		Description: description,
		Color:       3447003,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Request", Value: leave.LeaveRequestTitle, Inline: true},
			{Name: "Type", Value: leave.UnavailabilityType, Inline: true},
			{Name: "Dates", Value: formatShortDateRange(*leave.StartDate, *leave.EndDate), Inline: true},
			{Name: "Details", Value: leave.AdditionalContext, Inline: false},
		},
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Approve",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("notion_leave_approve_%s", leave.PageID),
					Emoji: discordgo.ComponentEmoji{
						Name: "✅",
					},
				},
				discordgo.Button{
					Label:    "Reject",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("notion_leave_reject_%s", leave.PageID),
					Emoji: discordgo.ComponentEmoji{
						Name: "❌",
					},
				},
			},
		},
	}

	msg, err := h.service.Discord.SendChannelMessageComplex(channelID, assigneeMentions, []*discordgo.MessageEmbed{embed}, components)
	if err != nil {
		l.Error(err, "failed to send leave request message to discord channel")
		// Fallback to auditlog
		fallbackDescription := fmt.Sprintf("%s request time off", contractor.FullName)
		if contractorDiscordMention != "" {
			fallbackDescription = fmt.Sprintf("%s request time off", contractorDiscordMention)
		}
		inlineTrue := true
		h.sendNotionLeaveDiscordNotification(ctx,
			"📋 Leave Request",
			fallbackDescription,
			3447003,
			[]model.DiscordMessageField{
				{Name: "Request", Value: leave.LeaveRequestTitle, Inline: &inlineTrue},
				{Name: "Type", Value: leave.UnavailabilityType, Inline: &inlineTrue},
				{Name: "Dates", Value: formatShortDateRange(*leave.StartDate, *leave.EndDate), Inline: nil},
				{Name: "Details", Value: leave.AdditionalContext, Inline: nil},
			},
		)
	} else {
		l.Debug(fmt.Sprintf("sent leave request message to discord channel: message_id=%s", msg.ID))
	}
}

// sendOnLeaveAutoFillErrorNotification sends an error notification to the fortress-logs Discord channel
// when on-leave contractor auto-fill fails
func (h *handler) sendOnLeaveAutoFillErrorNotification(pageID, userID, errorMsg string) {
	const fortressLogsChannelID = "1409767264298860665"

	l := h.logger.Fields(logger.Fields{
		"handler": "webhook",
		"method":  "sendOnLeaveAutoFillErrorNotification",
		"page_id": pageID,
	})

	if h.service.Discord == nil {
		l.Debug("discord service not configured, skipping error notification")
		return
	}

	l.Debug(fmt.Sprintf("sending on-leave auto-fill error notification to fortress-logs channel: %s", fortressLogsChannelID))

	embed := &discordgo.MessageEmbed{
		Title:       "⚠️ On-Leave Contractor Fill Failed",
		Description: "Failed to automatically fill contractor information in leave request.",
		Color:       0xFF0000,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Page ID",
				Value:  pageID,
				Inline: true,
			},
			{
				Name:   "Created By User",
				Value:  userID,
				Inline: true,
			},
			{
				Name:   "Error",
				Value:  errorMsg,
				Inline: false,
			},
			{
				Name:   "Action Required",
				Value:  "Please manually fill the contractor field in Notion",
				Inline: false,
			},
		},
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000-07:00"),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Fortress API - Notion Webhook Handler",
		},
	}

	_, err := h.service.Discord.SendChannelMessageComplex(
		fortressLogsChannelID,
		"",
		[]*discordgo.MessageEmbed{embed},
		nil,
	)
	if err != nil {
		l.Error(err, "failed to send on-leave auto-fill error notification to discord")
		return
	}

	l.Info(fmt.Sprintf("successfully sent on-leave auto-fill error notification to fortress-logs channel: page_id=%s", pageID))
}

// lookupContractorByUserIDForOnLeave queries the contractor database to find a contractor by Notion user ID
func (h *handler) lookupContractorByUserIDForOnLeave(ctx context.Context, l logger.Logger, userID string) (string, error) {
	contractorDBID := h.config.Notion.Databases.Contractor
	if contractorDBID == "" {
		return "", fmt.Errorf("NOTION_CONTRACTOR_DB_ID not configured")
	}

	l.Debug(fmt.Sprintf("querying contractor database: db_id=%s user_id=%s", contractorDBID, userID))

	// Create Notion client
	client := nt.NewClient(h.config.Notion.Secret)

	// Query contractor database for matching Person (user ID)
	filter := &nt.DatabaseQueryFilter{
		Property: "Person",
		DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
			People: &nt.PeopleDatabaseQueryFilter{
				Contains: userID,
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
		l.Debug(fmt.Sprintf("no contractor found for user_id: %s", userID))
		return "", nil
	}

	pageID := resp.Results[0].ID
	l.Debug(fmt.Sprintf("found contractor: user_id=%s page_id=%s", userID, pageID))
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

// getContractorEmail fetches the Team Email from a contractor page
func (h *handler) getContractorEmail(ctx context.Context, l logger.Logger, contractorPageID string) (string, error) {
	l.Debug(fmt.Sprintf("fetching contractor email: contractor_page_id=%s", contractorPageID))

	// Create Notion client
	client := nt.NewClient(h.config.Notion.Secret)

	// Fetch contractor page
	page, err := client.FindPageByID(ctx, contractorPageID)
	if err != nil {
		l.Error(err, fmt.Sprintf("failed to fetch contractor page: contractor_page_id=%s", contractorPageID))
		return "", fmt.Errorf("failed to fetch contractor page: %w", err)
	}

	// Extract Team Email property
	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		return "", errors.New("failed to cast contractor page properties")
	}

	if emailProp, ok := props["Team Email"]; ok && emailProp.Email != nil {
		email := *emailProp.Email
		l.Debug(fmt.Sprintf("extracted contractor email: contractor_page_id=%s email=%s", contractorPageID, email))
		return email, nil
	}

	l.Debug(fmt.Sprintf("no email found for contractor: contractor_page_id=%s", contractorPageID))
	return "", nil
}

// updateUnavailabilityType updates the Unavailability Type property on a leave request page
func (h *handler) updateUnavailabilityType(ctx context.Context, l logger.Logger, leavePageID, unavailabilityType string) error {
	l.Debug(fmt.Sprintf("updating unavailability type: leave_page=%s type=%s", leavePageID, unavailabilityType))

	// Create Notion client
	client := nt.NewClient(h.config.Notion.Secret)

	// Build update params with Unavailability Type select
	updateParams := nt.UpdatePageParams{
		DatabasePageProperties: nt.DatabasePageProperties{
			"Unavailability Type": nt.DatabasePageProperty{
				Select: &nt.SelectOptions{
					Name: unavailabilityType,
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
	l.Debug(fmt.Sprintf("successfully updated unavailability type on leave request page: %s", leavePageID))
	return nil
}

// Helper functions for automation payload

func getAutomationStatusName(prop NotionAutomationStatusProperty) string {
	if prop.Status != nil {
		return prop.Status.Name
	}
	return ""
}

func getAutomationCreatedByUserID(prop NotionAutomationCreatedByProperty) string {
	if prop.CreatedBy != nil {
		return prop.CreatedBy.ID
	}
	return ""
}
