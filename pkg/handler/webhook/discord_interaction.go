package webhook

import (
	"context"
	"crypto/ed25519"
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

	invoiceCtrl "github.com/dwarvesf/fortress-api/pkg/controller/invoice"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	googleSvc "github.com/dwarvesf/fortress-api/pkg/service/google"
	"github.com/dwarvesf/fortress-api/pkg/service/nocodb"
	notionSvc "github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

// HandleDiscordInteraction handles Discord interaction webhooks (button clicks, etc.)
func (h *handler) HandleDiscordInteraction(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "webhook",
		"method":  "HandleDiscordInteraction",
	})

	l.Debugf("received discord interaction webhook")

	// Read body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		l.Errorf(err, "failed to read discord interaction body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Debugf("received interaction body: %s", string(body))

	// Verify signature
	publicKey := h.config.Discord.PublicKey
	if publicKey == "" {
		l.Error(errors.New("discord public key not configured"), "cannot verify interaction")
		c.JSON(http.StatusUnauthorized, view.CreateResponse[any](nil, nil, errors.New("discord public key not configured"), nil, ""))
		return
	}

	signature := c.GetHeader("X-Signature-Ed25519")
	timestamp := c.GetHeader("X-Signature-Timestamp")

	if !verifyDiscordSignature(publicKey, signature, timestamp, body) {
		l.Error(errors.New("invalid discord signature"), "signature verification failed")
		c.JSON(http.StatusUnauthorized, view.CreateResponse[any](nil, nil, errors.New("invalid signature"), nil, ""))
		return
	}

	// Parse interaction
	var interaction discordgo.Interaction
	if err := json.Unmarshal(body, &interaction); err != nil {
		l.Errorf(err, "failed to parse discord interaction")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Debugf("parsed interaction: type=%d", interaction.Type)

	// Handle different interaction types
	switch interaction.Type {
	case discordgo.InteractionPing:
		// Respond to ping (required for Discord to verify endpoint)
		l.Debugf("responding to discord ping")
		c.JSON(http.StatusOK, gin.H{"type": discordgo.InteractionResponsePong})
		return

	case discordgo.InteractionMessageComponent:
		h.handleMessageComponentInteraction(c, l, &interaction)
		return

	default:
		l.Infof("ignoring interaction type: %d", interaction.Type)
		c.JSON(http.StatusOK, gin.H{"type": discordgo.InteractionResponsePong})
		return
	}
}

// handleMessageComponentInteraction handles button clicks and other message components
func (h *handler) handleMessageComponentInteraction(c *gin.Context, l logger.Logger, interaction *discordgo.Interaction) {
	data := interaction.MessageComponentData()
	customID := data.CustomID

	l.Debugf("handling message component interaction: custom_id=%s user=%s", customID, interaction.Member.User.Username)

	// Parse custom ID to determine action
	if strings.HasPrefix(customID, "leave_approve_") {
		leaveIDStr := strings.TrimPrefix(customID, "leave_approve_")
		leaveID, err := strconv.Atoi(leaveIDStr)
		if err != nil {
			l.Errorf(err, "invalid leave id in custom_id: %s", customID)
			h.respondToInteraction(c, "Invalid leave request ID")
			return
		}
		h.handleLeaveApproveButton(c, l, interaction, leaveID)
		return
	}

	if strings.HasPrefix(customID, "leave_reject_") {
		leaveIDStr := strings.TrimPrefix(customID, "leave_reject_")
		leaveID, err := strconv.Atoi(leaveIDStr)
		if err != nil {
			l.Errorf(err, "invalid leave id in custom_id: %s", customID)
			h.respondToInteraction(c, "Invalid leave request ID")
			return
		}
		h.handleLeaveRejectButton(c, l, interaction, leaveID)
		return
	}

	// Notion leave handlers
	if strings.HasPrefix(customID, "notion_leave_approve_") {
		pageID := strings.TrimPrefix(customID, "notion_leave_approve_")
		if pageID == "" {
			l.Errorf(nil, "empty page id in custom_id: %s", customID)
			h.respondToInteraction(c, "Invalid leave request ID")
			return
		}
		h.handleNotionLeaveApproveButton(c, l, interaction, pageID)
		return
	}

	if strings.HasPrefix(customID, "notion_leave_reject_") {
		pageID := strings.TrimPrefix(customID, "notion_leave_reject_")
		if pageID == "" {
			l.Errorf(nil, "empty page id in custom_id: %s", customID)
			h.respondToInteraction(c, "Invalid leave request ID")
			return
		}
		h.handleNotionLeaveRejectButton(c, l, interaction, pageID)
		return
	}

	if strings.HasPrefix(customID, "invoice_paid_confirm_") {
		// Format: invoice_paid_confirm_{invoiceNumber}_{discordUserID}
		suffix := strings.TrimPrefix(customID, "invoice_paid_confirm_")
		parts := strings.Split(suffix, "_")
		if len(parts) < 1 || parts[0] == "" {
			l.Errorf(nil, "invalid invoice number in custom_id: %s", customID)
			h.respondToInteraction(c, "Invalid invoice number")
			return
		}
		invoiceNumber := parts[0]
		l.Debugf("parsed invoice_paid_confirm: invoiceNumber=%s", invoiceNumber)
		h.handleInvoicePaidConfirmButton(c, l, interaction, invoiceNumber)
		return
	}

	l.Infof("unknown custom_id: %s", customID)
	h.respondToInteraction(c, "Unknown action")
}

// handleLeaveApproveButton handles the approve button click
func (h *handler) handleLeaveApproveButton(c *gin.Context, l logger.Logger, interaction *discordgo.Interaction, leaveID int) {
	l.Debugf("approving leave request via button: leave_id=%d approver=%s", leaveID, interaction.Member.User.Username)

	// Get approver's email from Discord ID
	approverEmail := ""
	if interaction.Member != nil && interaction.Member.User != nil {
		// Try to find employee by Discord ID
		employee, err := h.store.Employee.GetByDiscordID(h.repo.DB(), interaction.Member.User.ID, false)
		if err == nil && employee != nil {
			approverEmail = employee.TeamEmail
			l.Debugf("found approver email: %s", approverEmail)
		} else {
			l.Warnf("could not find employee for discord user: %s", interaction.Member.User.ID)
		}
	}

	// Update NocoDB status
	leaveService := nocodb.NewLeaveService(h.service.NocoDB, h.config, h.store, h.repo, h.logger)
	err := leaveService.UpdateLeaveStatus(leaveID, "Approved", approverEmail)
	if err != nil {
		l.Errorf(err, "failed to update leave status in nocodb")
		h.respondToInteraction(c, fmt.Sprintf("❌ Failed to approve leave request: %v", err))
		return
	}

	l.Infof("leave request approved via button: leave_id=%d approver=%s", leaveID, approverEmail)

	// Update the message to show it's been approved
	h.updateLeaveMessageStatus(c, l, interaction, leaveID, "Approved", interaction.Member.User.Username)
}

// handleLeaveRejectButton handles the reject button click
func (h *handler) handleLeaveRejectButton(c *gin.Context, l logger.Logger, interaction *discordgo.Interaction, leaveID int) {
	l.Debugf("rejecting leave request via button: leave_id=%d rejector=%s", leaveID, interaction.Member.User.Username)

	// Update NocoDB status
	leaveService := nocodb.NewLeaveService(h.service.NocoDB, h.config, h.store, h.repo, h.logger)
	err := leaveService.UpdateLeaveStatus(leaveID, "Rejected", "")
	if err != nil {
		l.Errorf(err, "failed to update leave status in nocodb")
		h.respondToInteraction(c, fmt.Sprintf("❌ Failed to reject leave request: %v", err))
		return
	}

	l.Infof("leave request rejected via button: leave_id=%d rejector=%s", leaveID, interaction.Member.User.Username)

	// Update the message to show it's been rejected
	h.updateLeaveMessageStatus(c, l, interaction, leaveID, "Rejected", interaction.Member.User.Username)
}

// handleInvoicePaidConfirmButton handles the invoice paid confirm button click
func (h *handler) handleInvoicePaidConfirmButton(c *gin.Context, l logger.Logger, interaction *discordgo.Interaction, invoiceNumber string) {
	l.Debugf("confirming invoice payment via button: invoiceNumber=%s user=%s", invoiceNumber, interaction.Member.User.Username)

	// Get channel and message IDs for later update
	channelID := interaction.ChannelID
	messageID := ""
	if interaction.Message != nil {
		messageID = interaction.Message.ID
	}
	actionBy := interaction.Member.User.Username

	l.Debugf("interaction context: channelID=%s messageID=%s", channelID, messageID)

	// Immediately respond with "Processing..." to avoid Discord timeout
	h.respondWithProcessing(c, l, interaction, invoiceNumber)

	// Process in background
	go h.processInvoicePaidConfirm(l, channelID, messageID, invoiceNumber, actionBy)
}

// respondWithProcessing sends an immediate "Processing..." response to Discord
func (h *handler) respondWithProcessing(c *gin.Context, l logger.Logger, interaction *discordgo.Interaction, invoiceNumber string) {
	l.Debugf("responding with processing status for invoice: %s", invoiceNumber)

	// Get original embed to preserve info
	var originalEmbed *discordgo.MessageEmbed
	if interaction.Message != nil && len(interaction.Message.Embeds) > 0 {
		originalEmbed = interaction.Message.Embeds[0]
	}

	// Build processing embed
	var fields []*discordgo.MessageEmbedField
	if originalEmbed != nil {
		fields = originalEmbed.Fields
	}

	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "Status",
		Value:  "⏳ Processing...",
		Inline: false,
	})

	processingEmbed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Invoice %s - Processing", invoiceNumber),
		Description: "Please wait while we process the payment confirmation...",
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

// processInvoicePaidConfirm processes the invoice payment in background and updates Discord message
func (h *handler) processInvoicePaidConfirm(l logger.Logger, channelID, messageID, invoiceNumber, actionBy string) {
	l.Debugf("background processing invoice payment: invoiceNumber=%s channelID=%s messageID=%s", invoiceNumber, channelID, messageID)

	// Mark invoice as paid - searches both PostgreSQL and Notion
	result, err := h.controller.Invoice.MarkInvoiceAsPaidByNumber(invoiceNumber)
	if err != nil {
		l.Errorf(err, "failed to mark invoice as paid: %s", invoiceNumber)
		h.updateInvoiceMessageWithError(l, channelID, messageID, invoiceNumber, err.Error())
		return
	}

	l.Infof("invoice marked as paid via discord button: invoiceNumber=%s user=%s source=%s", invoiceNumber, actionBy, result.Source)

	// Log to Discord audit
	if err := h.controller.Discord.Log(model.LogDiscordInput{
		Type: "invoice_paid",
		Data: map[string]interface{}{
			"invoice_number": invoiceNumber,
			"source":         result.Source,
		},
	}); err != nil {
		l.Errorf(err, "failed to log invoice paid to discord")
	}

	// Update the message to show success
	h.updateInvoiceMessageWithResult(l, channelID, messageID, result, actionBy)
}

// updateInvoiceMessageWithResult updates the Discord message with MarkPaidResult
func (h *handler) updateInvoiceMessageWithResult(l logger.Logger, channelID, messageID string, result *invoiceCtrl.MarkPaidResult, actionBy string) {
	l.Debugf("updating invoice message with result: invoiceNumber=%s source=%s actionBy=%s channelID=%s messageID=%s", result.InvoiceNumber, result.Source, actionBy, channelID, messageID)

	if channelID == "" || messageID == "" {
		l.Warnf("cannot update message: missing channelID or messageID")
		return
	}

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "Status",
			Value:  "✅ Paid",
			Inline: false,
		},
		{
			Name:   "Marked by",
			Value:  actionBy,
			Inline: false,
		},
	}

	successEmbed := &discordgo.MessageEmbed{
		Title:     fmt.Sprintf("Invoice %s", result.InvoiceNumber),
		Color:     3066993, // Green
		Fields:    fields,
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	_, err := h.service.Discord.UpdateChannelMessage(channelID, messageID, "", []*discordgo.MessageEmbed{successEmbed}, []discordgo.MessageComponent{})
	if err != nil {
		l.Errorf(err, "failed to update discord message with result status")
	} else {
		l.Debugf("successfully updated discord message with result status")
	}
}

// updateInvoiceMessageWithError updates the Discord message with error status
func (h *handler) updateInvoiceMessageWithError(l logger.Logger, channelID, messageID, invoiceNumber, errorMsg string) {
	l.Debugf("updating invoice message with error: invoiceNumber=%s error=%s channelID=%s messageID=%s", invoiceNumber, errorMsg, channelID, messageID)

	if channelID == "" || messageID == "" {
		l.Warnf("cannot update message: missing channelID or messageID")
		return
	}

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "Error",
			Value:  fmt.Sprintf("❌ %s", errorMsg),
			Inline: false,
		},
	}

	errorEmbed := &discordgo.MessageEmbed{
		Title:     fmt.Sprintf("❌ Invoice %s - Failed", invoiceNumber),
		Color:     15158332, // Red
		Fields:    fields,
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	_, err := h.service.Discord.UpdateChannelMessage(channelID, messageID, "", []*discordgo.MessageEmbed{errorEmbed}, []discordgo.MessageComponent{})
	if err != nil {
		l.Errorf(err, "failed to update discord message with error status")
	} else {
		l.Debugf("successfully updated discord message with error status")
	}
}

// updateLeaveMessageStatus updates the original message to show the new status
func (h *handler) updateLeaveMessageStatus(c *gin.Context, l logger.Logger, interaction *discordgo.Interaction, leaveID int, status string, actionBy string) {
	// Get original embed
	var originalEmbed *discordgo.MessageEmbed
	if interaction.Message != nil && len(interaction.Message.Embeds) > 0 {
		originalEmbed = interaction.Message.Embeds[0]
	}

	// Determine color and title based on status
	var color int
	var title string
	var emoji string

	if status == "Approved" {
		color = 3066993 // Green
		title = "✅ Leave Request Approved"
		emoji = "✅"
	} else {
		color = 15158332 // Red
		title = "❌ Leave Request Rejected"
		emoji = "❌"
	}

	// Build updated embed
	var fields []*discordgo.MessageEmbedField
	if originalEmbed != nil {
		fields = originalEmbed.Fields
	}

	// Add status field
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Status", emoji),
		Value:  fmt.Sprintf("%s by %s", status, actionBy),
		Inline: false,
	})

	updatedEmbed := &discordgo.MessageEmbed{
		Title:       title,
		Description: "",
		Color:       color,
		Fields:      fields,
		Timestamp:   time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	// Respond with updated message (removes buttons)
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{updatedEmbed},
			Components: []discordgo.MessageComponent{}, // Remove buttons
		},
	}

	c.JSON(http.StatusOK, response)
}

// respondToInteraction sends a simple response to an interaction
func (h *handler) respondToInteraction(c *gin.Context, message string) {
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral, // Only visible to the user
		},
	}
	c.JSON(http.StatusOK, response)
}

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

		// Update Notion status
		err := leaveService.UpdateLeaveStatus(ctx, pageID, "Approved", approverPageID)
		if err != nil {
			l.Errorf(err, "failed to update leave status in Notion")
			h.updateNotionLeaveMessageWithError(l, channelID, messageID, originalEmbed, fmt.Sprintf("Failed to approve: %v", err))
			return
		}

		l.Infof("Notion leave request approved via button: page_id=%s approver=%s", pageID, approverEmail)

		// Create Google Calendar event
		l.Debug("fetching leave details for calendar event creation")
		if err := h.createCalendarEventForLeave(ctx, l, leaveService, pageID); err != nil {
			l.Error(err, "failed to create calendar event (non-fatal, continuing)")
			// Non-fatal - continue even if calendar creation fails
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
		err := leaveService.UpdateLeaveStatus(ctx, pageID, "Rejected", rejectorPageID)
		if err != nil {
			l.Errorf(err, "failed to update leave status in Notion")
			h.updateNotionLeaveMessageWithError(l, channelID, messageID, originalEmbed, fmt.Sprintf("Failed to reject: %v", err))
			return
		}

		l.Infof("Notion leave request rejected via button: page_id=%s rejector=%s", pageID, rejectorUsername)

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

	l.Debugf("fetched leave request: start=%v end=%v email=%s", leave.StartDate, leave.EndDate, leave.Email)

	// Look up employee by email with DiscordAccount preloaded
	employees, err := h.store.Employee.GetByEmails(h.repo.DB().Preload("DiscordAccount"), []string{leave.Email})
	if err != nil {
		return fmt.Errorf("failed to find employee by email %s: %w", leave.Email, err)
	}
	if len(employees) == 0 {
		return fmt.Errorf("no employee found with email %s", leave.Email)
	}
	employee := employees[0]

	l.Debugf("found employee: full_name=%s team_email=%s", employee.FullName, employee.TeamEmail)

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

	// Get Discord username from employee
	discordUsername := ""
	if employee.DiscordAccount != nil && employee.DiscordAccount.DiscordUsername != "" {
		discordUsername = employee.DiscordAccount.DiscordUsername
	} else {
		// Fallback to full name if no Discord username
		discordUsername = employee.FullName
	}

	// Build event summary and description matching the format: "👾 <discord_username> off"
	summary := fmt.Sprintf("👾 %s off", discordUsername)
	description := fmt.Sprintf("Leave Type: %s\nReason: %s\nRequested by: %s (%s)",
		leave.LeaveType, leave.Reason, employee.FullName, employee.TeamEmail)

	// Create all-day event (Notion database doesn't have Shift field)
	// Add assigned approvers as attendees
	attendees := []string{employee.TeamEmail}
	if len(leave.Assignees) > 0 {
		attendees = append(attendees, leave.Assignees...)
		l.Debugf("added %d assignees as attendees: %v", len(leave.Assignees), leave.Assignees)
	} else {
		// Fallback: get stakeholder emails from deployments if no assignees
		l.Debug("no assignees found, fetching stakeholders from deployments")
		stakeholderEmails := h.getStakeholderEmailsFromDeployments(ctx, l, leaveService, employee.TeamEmail)
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
		Email:       employee.TeamEmail,
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

// verifyDiscordSignature verifies the Discord interaction signature
func verifyDiscordSignature(publicKey, signature, timestamp string, body []byte) bool {
	// Decode public key
	pubKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return false
	}

	// Decode signature
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	// Create message to verify
	message := append([]byte(timestamp), body...)

	// Verify signature
	return ed25519.Verify(pubKeyBytes, message, sigBytes)
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
		l.Info("no stakeholder IDs found, using fallback assignees: han@d.foundation, thanhpd@d.foundation")
		return []string{"han@d.foundation", "thanhpd@d.foundation"}
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
		l.Info("no stakeholder emails found, using fallback assignees: han@d.foundation, thanhpd@d.foundation")
		emails = []string{"han@d.foundation", "thanhpd@d.foundation"}
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

// Ensure handler has Nocodb service
func init() {
	// This will be handled by dependency injection in main.go
}
