package webhook

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	nt "github.com/dstotijn/go-notion"
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	googleSvc "github.com/dwarvesf/fortress-api/pkg/service/google"
	notionSvc "github.com/dwarvesf/fortress-api/pkg/service/notion"
)

// handleNotionLeaveApproveButton handles the approve button click for Notion leave requests
func (h *handler) handleNotionLeaveApproveButton(c *gin.Context, l logger.Logger, interaction *discordgo.Interaction, pageID string) {
	l.Debugf("approving Notion leave request via button: page_id=%s approver=%s", pageID, interaction.Member.User.Username)

	// Respond immediately with processing embed to avoid Discord timeout (3 seconds)
	h.respondWithNotionLeaveProcessingEmbed(c, l, interaction, "Approving")

	// Get approver's email from Discord ID
	approverEmail := ""
	approverUsername := ""
	if interaction.Member != nil && interaction.Member.User != nil {
		approverUsername = interaction.Member.User.Username
		// Try to find employee by Discord ID
		employee, err := h.store.Employee.GetByDiscordID(h.repo.DB(), interaction.Member.User.ID, false)
		if err == nil && employee != nil {
			approverEmail = employee.TeamEmail
			l.Debugf("found approver email: %s", approverEmail)
		} else {
			l.Warnf("could not find employee for discord user: %s", interaction.Member.User.ID)
		}
	}

	// Get channel and message IDs for later update
	channelID := interaction.ChannelID
	messageID := ""
	if interaction.Message != nil {
		messageID = interaction.Message.ID
	}

	// Store original embed for async update
	var originalEmbed *discordgo.MessageEmbed
	if interaction.Message != nil && len(interaction.Message.Embeds) > 0 {
		originalEmbed = interaction.Message.Embeds[0]
	}

	// Process asynchronously
	go func() {
		ctx := context.Background()
		l.Debugf("async processing approval for page_id=%s", pageID)

		// Create Notion leave service
		leaveService := notionSvc.NewLeaveService(h.config, h.store, h.repo, h.logger)
		if leaveService == nil {
			l.Error(errors.New("failed to create notion leave service"), "notion secret may not be configured")
			h.updateNotionLeaveMessageWithError(l, channelID, messageID, originalEmbed, "Notion service not configured")
			return
		}

		// Look up approver's Notion page ID
		approverPageID := ""
		if interaction.Member != nil && interaction.Member.User != nil {
			discordID := interaction.Member.User.ID
			l.Debugf("looking up contractor page for Discord user: %s (discord_id=%s)", approverUsername, discordID)

			// Try to find contractor by Discord account
			var err error
			approverPageID, err = h.getContractorPageIDByDiscordID(ctx, l, discordID)
			if err != nil {
				l.Warnf("could not find contractor page for discord_id %s: %v", discordID, err)
				// Fallback to email lookup if Discord ID lookup fails
				if approverEmail != "" {
					l.Debugf("falling back to email lookup: %s", approverEmail)
					approverPageID, err = leaveService.GetContractorPageIDByEmail(ctx, approverEmail)
					if err != nil {
						l.Warnf("could not find contractor page for approver email %s: %v", approverEmail, err)
					} else if approverPageID != "" {
						l.Debugf("found contractor page by email: %s page_id=%s", approverEmail, approverPageID)
					}
				}
			} else if approverPageID != "" {
				l.Debugf("found contractor page by Discord ID: discord_id=%s page_id=%s", discordID, approverPageID)
			}
		}

		// Execute side effects independently — neither blocks the other
		notionUpdated := true
		calendarCreated := true

		// Side effect 1: Update Notion status
		var notionErr error
		notionErr = leaveService.UpdateLeaveStatus(ctx, pageID, "Acknowledged", approverPageID)
		if notionErr != nil {
			l.Errorf(notionErr, "failed to update leave status in Notion (will still create calendar): page_id=%s", pageID)
			notionUpdated = false
		} else {
			l.Infof("notion_updated=true: page_id=%s approver=%s", pageID, approverEmail)
		}

		// Side effect 2: Create Google Calendar event (independent of Notion result)
		l.Debug("fetching leave details for calendar event creation")
		var calendarErr error
		if calendarErr = h.createCalendarEventForLeave(ctx, l, leaveService, pageID); calendarErr != nil {
			l.Errorf(calendarErr, "calendar_created=false: page_id=%s", pageID)
			calendarCreated = false
		} else {
			l.Infof("calendar_created=true: page_id=%s", pageID)
		}

		// Send failure summary to audit log channel
		if !notionUpdated || !calendarCreated {
			h.sendLeaveOperationLogToAuditLog(l, pageID, "Approved", approverUsername, notionUpdated, calendarCreated, notionErr, calendarErr)
		}

		// Update the original message to show it's been approved
		h.updateNotionLeaveMessageWithStatus(l, channelID, messageID, originalEmbed, "Approved", approverUsername)
	}()
}

// handleNotionLeaveRejectButton handles the reject button click for Notion leave requests
func (h *handler) handleNotionLeaveRejectButton(c *gin.Context, l logger.Logger, interaction *discordgo.Interaction, pageID string) {
	l.Debugf("rejecting Notion leave request via button: page_id=%s rejector=%s", pageID, interaction.Member.User.Username)

	// Respond immediately with processing embed to avoid Discord timeout (3 seconds)
	h.respondWithNotionLeaveProcessingEmbed(c, l, interaction, "Rejecting")

	rejectorUsername := ""
	if interaction.Member != nil && interaction.Member.User != nil {
		rejectorUsername = interaction.Member.User.Username
	}

	// Get channel and message IDs for later update
	channelID := interaction.ChannelID
	messageID := ""
	if interaction.Message != nil {
		messageID = interaction.Message.ID
	}

	// Store original embed for async update
	var originalEmbed *discordgo.MessageEmbed
	if interaction.Message != nil && len(interaction.Message.Embeds) > 0 {
		originalEmbed = interaction.Message.Embeds[0]
	}

	// Process asynchronously
	go func() {
		ctx := context.Background()
		l.Debugf("async processing rejection for page_id=%s", pageID)

		// Create Notion leave service
		leaveService := notionSvc.NewLeaveService(h.config, h.store, h.repo, h.logger)
		if leaveService == nil {
			l.Error(errors.New("failed to create notion leave service"), "notion secret may not be configured")
			h.updateNotionLeaveMessageWithError(l, channelID, messageID, originalEmbed, "Notion service not configured")
			return
		}

		// Look up rejector's Notion page ID
		rejectorPageID := ""
		if interaction.Member != nil && interaction.Member.User != nil {
			discordID := interaction.Member.User.ID
			l.Debugf("looking up contractor page for Discord user: %s (discord_id=%s)", rejectorUsername, discordID)

			// Try to find contractor by Discord account
			var err error
			rejectorPageID, err = h.getContractorPageIDByDiscordID(ctx, l, discordID)
			if err != nil {
				l.Warnf("could not find contractor page for discord_id %s: %v", discordID, err)
			} else if rejectorPageID != "" {
				l.Debugf("found contractor page by Discord ID: discord_id=%s page_id=%s", discordID, rejectorPageID)
			}
		}

		// Update Notion status with rejector page ID
		notionUpdated := true
		var notionErr error
		notionErr = leaveService.UpdateLeaveStatus(ctx, pageID, "Not Applicable", rejectorPageID)
		if notionErr != nil {
			l.Errorf(notionErr, "notion_updated=false: failed to update leave status: page_id=%s", pageID)
			notionUpdated = false
		} else {
			l.Infof("notion_updated=true: page_id=%s rejector=%s", pageID, rejectorUsername)
		}

		// Send failure summary to audit log channel
		if !notionUpdated {
			h.sendLeaveOperationLogToAuditLog(l, pageID, "Rejected", rejectorUsername, notionUpdated, false, notionErr, nil)
		}

		// Update the original message to show it's been rejected
		h.updateNotionLeaveMessageWithStatus(l, channelID, messageID, originalEmbed, "Rejected", rejectorUsername)
	}()
}

// getContractorPageIDByDiscordID looks up contractor page by Discord ID
func (h *handler) getContractorPageIDByDiscordID(ctx context.Context, l logger.Logger, discordID string) (string, error) {
	l.Debugf("looking up contractor by Discord ID: %s", discordID)

	// 1. Look up discord_accounts table to get the account UUID
	var discordAccount struct {
		ID string
	}
	err := h.repo.DB().WithContext(ctx).
		Table("discord_accounts").
		Select("id").
		Where("discord_id = ? AND deleted_at IS NULL", discordID).
		First(&discordAccount).Error
	if err != nil {
		l.Debugf("discord account not found for discord_id: %s, error: %v", discordID, err)
		return "", fmt.Errorf("discord account not found: %w", err)
	}

	l.Debugf("found discord account: discord_id=%s account_id=%s", discordID, discordAccount.ID)

	// 2. Look up employees table by discord_account_id to get team_email
	var employee struct {
		TeamEmail string
	}
	err = h.repo.DB().WithContext(ctx).
		Table("employees").
		Select("team_email").
		Where("discord_account_id = ? AND deleted_at IS NULL", discordAccount.ID).
		First(&employee).Error
	if err != nil {
		l.Debugf("employee not found for discord_account_id: %s, error: %v", discordAccount.ID, err)
		return "", fmt.Errorf("employee not found for discord account: %w", err)
	}

	l.Debugf("found employee: discord_account_id=%s team_email=%s", discordAccount.ID, employee.TeamEmail)

	// 3. Query Notion Contractor database by Team Email
	leaveService := notionSvc.NewLeaveService(h.config, h.store, h.repo, h.logger)
	if leaveService == nil {
		return "", fmt.Errorf("notion service not configured")
	}

	contractorPageID, err := leaveService.GetContractorPageIDByEmail(ctx, employee.TeamEmail)
	if err != nil {
		l.Debugf("contractor page not found for email: %s, error: %v", employee.TeamEmail, err)
		return "", fmt.Errorf("contractor page not found: %w", err)
	}

	if contractorPageID == "" {
		l.Debugf("no contractor page found for email: %s", employee.TeamEmail)
		return "", nil
	}

	l.Debugf("found contractor page: email=%s page_id=%s", employee.TeamEmail, contractorPageID)
	return contractorPageID, nil
}

// respondWithNotionLeaveProcessingEmbed responds immediately with a processing status embed
func (h *handler) respondWithNotionLeaveProcessingEmbed(c *gin.Context, l logger.Logger, interaction *discordgo.Interaction, action string) {
	l.Debugf("responding with processing embed for action: %s", action)

	// Get original embed to preserve info
	var originalEmbed *discordgo.MessageEmbed
	if interaction.Message != nil && len(interaction.Message.Embeds) > 0 {
		originalEmbed = interaction.Message.Embeds[0]
	}

	// Build processing embed preserving original fields
	var fields []*discordgo.MessageEmbedField
	if originalEmbed != nil {
		fields = originalEmbed.Fields
	}

	// Add processing status field
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "Status",
		Value:  fmt.Sprintf("⏳ %s...", action),
		Inline: false,
	})

	title := "Leave Request"
	description := ""
	if originalEmbed != nil {
		if originalEmbed.Title != "" {
			title = originalEmbed.Title
		}
		description = originalEmbed.Description
	}

	processingEmbed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("⏳ %s - Processing", title),
		Description: description,
		Color:       16776960, // Yellow
		Fields:      fields,
		Timestamp:   time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	// Respond immediately with processing status (removes buttons to prevent double-click)
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{processingEmbed},
			Components: []discordgo.MessageComponent{}, // Remove buttons
		},
	}

	c.JSON(http.StatusOK, response)
}

// updateNotionLeaveMessageWithStatus updates the message with final status embed
func (h *handler) updateNotionLeaveMessageWithStatus(l logger.Logger, channelID, messageID string, originalEmbed *discordgo.MessageEmbed, status, actionBy string) {
	l.Debugf("updating notion leave message with status: channelID=%s messageID=%s status=%s actionBy=%s", channelID, messageID, status, actionBy)

	if channelID == "" || messageID == "" {
		l.Warnf("cannot update message: missing channelID or messageID")
		return
	}

	// Determine color and title based on status
	var color int
	var title string

	if status == "Approved" {
		color = 3066993 // Green
		title = "✅ Leave Request Approved"
	} else {
		color = 15158332 // Red
		title = "❌ Leave Request Rejected"
	}

	// Build updated embed preserving original fields
	var fields []*discordgo.MessageEmbedField
	if originalEmbed != nil {
		fields = originalEmbed.Fields
	}

	// Add status field
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "Status",
		Value:  fmt.Sprintf("%s by %s", status, actionBy),
		Inline: false,
	})

	description := ""
	if originalEmbed != nil {
		description = originalEmbed.Description
	}

	updatedEmbed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       color,
		Fields:      fields,
		Timestamp:   time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	// Update the message (removes buttons)
	_, err := h.service.Discord.UpdateChannelMessage(channelID, messageID, "", []*discordgo.MessageEmbed{updatedEmbed}, []discordgo.MessageComponent{})
	if err != nil {
		l.Errorf(err, "failed to update discord message with status")
	} else {
		l.Debugf("successfully updated discord message with status: %s", status)
	}
}

// updateNotionLeaveMessageWithError updates the message with error status embed
func (h *handler) updateNotionLeaveMessageWithError(l logger.Logger, channelID, messageID string, originalEmbed *discordgo.MessageEmbed, errorMsg string) {
	l.Debugf("updating notion leave message with error: channelID=%s messageID=%s error=%s", channelID, messageID, errorMsg)

	if channelID == "" || messageID == "" {
		l.Warnf("cannot update message: missing channelID or messageID")
		return
	}

	// Build error embed preserving original fields
	var fields []*discordgo.MessageEmbedField
	if originalEmbed != nil {
		fields = originalEmbed.Fields
	}

	// Add error field
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "❌ Error",
		Value:  errorMsg,
		Inline: false,
	})

	description := ""
	if originalEmbed != nil {
		description = originalEmbed.Description
	}

	errorEmbed := &discordgo.MessageEmbed{
		Title:       "❌ Leave Request - Failed",
		Description: description,
		Color:       15158332, // Red
		Fields:      fields,
		Timestamp:   time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	// Update the message
	_, err := h.service.Discord.UpdateChannelMessage(channelID, messageID, "", []*discordgo.MessageEmbed{errorEmbed}, []discordgo.MessageComponent{})
	if err != nil {
		l.Errorf(err, "failed to update discord message with error status")
	} else {
		l.Debugf("successfully updated discord message with error status")
	}
}

// createCalendarEventForLeave creates a Google Calendar event for an approved leave request
func (h *handler) createCalendarEventForLeave(ctx context.Context, l logger.Logger, leaveService *notionSvc.LeaveService, pageID string) error {
	l.Debugf("creating calendar event for leave page_id=%s", pageID)

	// Fetch leave details from Notion
	leave, err := leaveService.GetLeaveRequest(ctx, pageID)
	if err != nil {
		return fmt.Errorf("failed to fetch leave request: %w", err)
	}

	l.Debugf("fetched leave request: start=%v end=%v email=%s contractor_page_id=%s", leave.StartDate, leave.EndDate, leave.Email, leave.EmployeeID)

	contractor, err := leaveService.LookupContractorDetailsByEmail(ctx, leave.Email)
	if err != nil {
		return fmt.Errorf("failed to resolve contractor by leave email %s: %w", leave.Email, err)
	}
	if contractor == nil {
		return fmt.Errorf("no contractor found with leave email %s", leave.Email)
	}

	contractorEmail := contractor.TeamEmail
	if contractorEmail == "" {
		contractorEmail = contractor.PersonalEmail
	}
	if contractorEmail == "" {
		contractorEmail = leave.Email
	}

	l.Debugf("resolved contractor for leave calendar flow: page_id=%s contractor_page_id=%s full_name=%s has_team_email=%t has_personal_email=%t has_discord_username=%t",
		pageID, contractor.PageID, contractor.FullName, contractor.TeamEmail != "", contractor.PersonalEmail != "", contractor.DiscordUsername != "")

	// Create calendar service
	calService := googleSvc.NewCalendarService(h.config, h.logger)

	// Validate dates
	if leave.StartDate == nil {
		return fmt.Errorf("start date is required for calendar event")
	}
	if leave.EndDate == nil {
		// If no end date, use start date as end date (single day leave)
		leave.EndDate = leave.StartDate
	}

	// Get Discord username from contractor
	discordUsername := ""
	if contractor.DiscordUsername != "" {
		discordUsername = contractor.DiscordUsername
	} else {
		// Fallback to full name if no Discord username
		discordUsername = contractor.FullName
	}

	// Build event summary and description matching the format: "👾 <discord_username> off"
	summary := fmt.Sprintf("👾 %s off", discordUsername)
	description := fmt.Sprintf("Type: %s\nRequest: %s\nDetails: %s\nRequested by: %s (%s)",
		leave.UnavailabilityType, leave.LeaveRequestTitle, leave.AdditionalContext, contractor.FullName, contractorEmail)

	// Create all-day event (Notion database doesn't have Shift field)
	// Add assigned approvers as attendees
	attendees := []string{contractorEmail}
	if len(leave.Assignees) > 0 {
		attendees = append(attendees, leave.Assignees...)
		l.Debugf("added %d assignees as attendees: %v", len(leave.Assignees), leave.Assignees)
	} else {
		// Fallback: get stakeholder emails from deployments if no assignees
		l.Debug("no assignees found, fetching stakeholders from deployments")
		stakeholderLookupEmail := contractor.TeamEmail
		if stakeholderLookupEmail == "" {
			stakeholderLookupEmail = leave.Email
		}
		stakeholderEmails := h.getStakeholderEmailsFromDeployments(ctx, l, leaveService, stakeholderLookupEmail)
		if len(stakeholderEmails) > 0 {
			attendees = append(attendees, stakeholderEmails...)
			l.Debugf("added %d stakeholder emails as attendees: %v", len(stakeholderEmails), stakeholderEmails)
		} else {
			l.Debug("no stakeholders found from deployments")
		}
	}

	event := googleSvc.CalendarEvent{
		Summary:     summary,
		Description: description,
		StartDate:   *leave.StartDate,
		EndDate:     *leave.EndDate,
		AllDay:      true,
		Email:       contractorEmail,
		Attendees:   attendees,
	}

	l.Debugf("creating calendar event: summary=%s start=%v end=%v all_day=true",
		event.Summary, event.StartDate, event.EndDate)

	createdEvent, err := calService.CreateLeaveEvent(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to create calendar event: %w", err)
	}

	l.Infof("successfully created calendar event: id=%s link=%s", createdEvent.Id, createdEvent.HtmlLink)
	return nil
}

// getStakeholderEmailsFromDeployments retrieves stakeholder emails from active deployments
func (h *handler) getStakeholderEmailsFromDeployments(
	ctx context.Context,
	l logger.Logger,
	leaveService *notionSvc.LeaveService,
	teamEmail string,
) []string {
	l.Debug(fmt.Sprintf("fetching stakeholder emails from deployments: team_email=%s", teamEmail))

	// Get stakeholder IDs from deployments
	stakeholderIDs := h.getStakeholderIDsFromDeployments(ctx, l, leaveService, teamEmail)
	if len(stakeholderIDs) == 0 {
		l.Info("no stakeholder IDs found, using fallback assignees: thanhpd@d.foundation")
		return []string{"thanhpd@d.foundation"}
	}

	// Convert stakeholder IDs to emails
	var emails []string
	for _, stakeholderID := range stakeholderIDs {
		l.Debug(fmt.Sprintf("fetching email for stakeholder: contractor_id=%s", stakeholderID))

		email, err := h.getContractorEmailFromNotion(ctx, l, stakeholderID)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to get contractor email: contractor_id=%s", stakeholderID))
			continue
		}
		if email == "" {
			l.Debug(fmt.Sprintf("email not found for stakeholder: contractor_id=%s", stakeholderID))
			continue
		}

		l.Debug(fmt.Sprintf("found email for stakeholder: contractor_id=%s email=%s", stakeholderID, email))
		emails = append(emails, email)
	}

	l.Debug(fmt.Sprintf("converted %d stakeholder IDs to %d emails", len(stakeholderIDs), len(emails)))

	// Fallback: if no emails found, use default assignees
	if len(emails) == 0 {
		l.Info("no stakeholder emails found, using fallback assignees: thanhpd@d.foundation")
		emails = []string{"thanhpd@d.foundation"}
	}

	return emails
}

// getStakeholderIDsFromDeployments extracts stakeholder contractor IDs from active deployments
func (h *handler) getStakeholderIDsFromDeployments(
	ctx context.Context,
	l logger.Logger,
	leaveService *notionSvc.LeaveService,
	teamEmail string,
) []string {
	var stakeholderIDs []string

	l.Debug(fmt.Sprintf("fetching stakeholders from deployments: team_email=%s", teamEmail))

	// Step 1: Lookup contractor by email
	contractorID, err := leaveService.LookupContractorByEmail(ctx, teamEmail)
	if err != nil {
		l.Error(err, fmt.Sprintf("failed to lookup contractor: email=%s", teamEmail))
		return stakeholderIDs
	}
	if contractorID == "" {
		l.Info(fmt.Sprintf("contractor not found for email: %s", teamEmail))
		return stakeholderIDs
	}

	l.Debug(fmt.Sprintf("found contractor: email=%s contractor_id=%s", teamEmail, contractorID))

	// Step 2: Get active deployments
	deployments, err := leaveService.GetActiveDeploymentsForContractor(ctx, contractorID)
	if err != nil {
		l.Error(err, fmt.Sprintf("failed to get active deployments: contractor_id=%s", contractorID))
		return stakeholderIDs
	}
	if len(deployments) == 0 {
		l.Info(fmt.Sprintf("no active deployments found: contractor_id=%s", contractorID))
		return stakeholderIDs
	}

	l.Debug(fmt.Sprintf("found %d active deployments for contractor: contractor_id=%s", len(deployments), contractorID))

	// Step 3: Extract stakeholders from all deployments (excluding the employee)
	stakeholderMap := make(map[string]bool)
	for i, deployment := range deployments {
		l.Debug(fmt.Sprintf("processing deployment %d/%d: deployment_id=%s", i+1, len(deployments), deployment.ID))

		stakeholders := leaveService.ExtractStakeholdersFromDeployment(ctx, deployment)
		l.Debug(fmt.Sprintf("extracted %d stakeholders from deployment: deployment_id=%s", len(stakeholders), deployment.ID))

		for _, stakeholderID := range stakeholders {
			// Filter out the employee themselves
			if stakeholderID == contractorID {
				l.Debug(fmt.Sprintf("skipping employee contractor from stakeholders: contractor_id=%s", stakeholderID))
				continue
			}
			stakeholderMap[stakeholderID] = true
		}
	}

	l.Debug(fmt.Sprintf("total unique stakeholders after filtering: %d", len(stakeholderMap)))

	// Convert map to slice
	for stakeholderID := range stakeholderMap {
		stakeholderIDs = append(stakeholderIDs, stakeholderID)
	}

	l.Debug(fmt.Sprintf("extracted %d stakeholder IDs from %d deployments", len(stakeholderIDs), len(deployments)))
	return stakeholderIDs
}

// getContractorEmailFromNotion fetches email from a contractor page in Notion
func (h *handler) getContractorEmailFromNotion(ctx context.Context, l logger.Logger, contractorPageID string) (string, error) {
	l.Debug(fmt.Sprintf("fetching contractor email from Notion: contractor_page_id=%s", contractorPageID))

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

// sendLeaveOperationLogToAuditLog sends a plain text failure summary to the audit log webhook.
// Only called when at least one operation failed.
func (h *handler) sendLeaveOperationLogToAuditLog(
	l logger.Logger,
	pageID, action, actionBy string,
	notionUpdated, calendarCreated bool,
	notionErr, calendarErr error,
) {
	auditLogURL := h.config.Discord.Webhooks.AuditLog
	if auditLogURL == "" {
		l.Debug("audit log webhook not configured, skipping operation log")
		return
	}

	if h.service.Discord == nil {
		l.Debug("discord service not configured, skipping operation log")
		return
	}

	l.Debugf("sending leave operation log to audit log: page_id=%s action=%s", pageID, action)

	// Build a simple plain text message listing what failed
	msg := fmt.Sprintf("⚠️ Leave %s by %s (page_id: %s)", action, actionBy, pageID)
	if !notionUpdated {
		msg += fmt.Sprintf("\n❌ notion_updated: %v", notionErr)
	}
	if action == "Approved" && !calendarCreated {
		msg += fmt.Sprintf("\n❌ calendar_created: %v", calendarErr)
	}

	_, err := h.service.Discord.SendMessage(model.DiscordMessage{
		Content: msg,
	}, auditLogURL)
	if err != nil {
		l.Errorf(err, "failed to send leave operation log to audit log: page_id=%s", pageID)
	} else {
		l.Debugf("sent leave operation log to audit log: page_id=%s", pageID)
	}
}
