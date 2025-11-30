package webhook

import (
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
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/nocodb"
	sInvoice "github.com/dwarvesf/fortress-api/pkg/store/invoice"
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

	// Look up invoice by number
	invoice, err := h.store.Invoice.One(h.repo.DB(), &sInvoice.Query{Number: invoiceNumber})
	if err != nil {
		l.Errorf(err, "failed to find invoice by number: %s", invoiceNumber)
		h.updateInvoiceMessageWithError(l, channelID, messageID, invoiceNumber, fmt.Sprintf("Invoice %s not found", invoiceNumber))
		return
	}

	l.Debugf("found invoice: id=%s number=%s status=%s", invoice.ID, invoice.Number, invoice.Status)

	// Check if invoice is already paid
	if invoice.Status == model.InvoiceStatusPaid {
		l.Debugf("invoice already paid: invoiceNumber=%s", invoiceNumber)
		h.updateInvoiceMessageWithSuccess(l, channelID, messageID, invoice, "Already Paid", actionBy)
		return
	}

	// Mark invoice as paid
	_, err = h.controller.Invoice.MarkInvoiceAsPaid(invoice, true)
	if err != nil {
		l.Errorf(err, "failed to mark invoice as paid: %s", invoiceNumber)
		h.updateInvoiceMessageWithError(l, channelID, messageID, invoiceNumber, fmt.Sprintf("Failed to mark invoice as paid: %v", err))
		return
	}

	l.Infof("invoice marked as paid via discord button: invoiceNumber=%s user=%s", invoiceNumber, actionBy)

	// Log to Discord audit
	if err := h.controller.Discord.Log(model.LogDiscordInput{
		Type: "invoice_paid",
		Data: map[string]interface{}{
			"invoice_number": invoice.Number,
		},
	}); err != nil {
		l.Errorf(err, "failed to log invoice paid to discord")
	}

	// Update the message to show success
	h.updateInvoiceMessageWithSuccess(l, channelID, messageID, invoice, "Paid", actionBy)
}

// updateInvoiceMessageWithSuccess updates the Discord message with success status
func (h *handler) updateInvoiceMessageWithSuccess(l logger.Logger, channelID, messageID string, invoice *model.Invoice, status string, actionBy string) {
	l.Debugf("updating invoice message with success: invoiceNumber=%s status=%s actionBy=%s channelID=%s messageID=%s", invoice.Number, status, actionBy, channelID, messageID)

	if channelID == "" || messageID == "" {
		l.Warnf("cannot update message: missing channelID or messageID")
		return
	}

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "Status",
			Value:  status,
			Inline: true,
		},
		{
			Name:   "Marked by",
			Value:  actionBy,
			Inline: true,
		},
	}

	successEmbed := &discordgo.MessageEmbed{
		Title:     fmt.Sprintf("Invoice %s - %s", invoice.Number, status),
		Color:     3066993, // Green
		Fields:    fields,
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	_, err := h.service.Discord.UpdateChannelMessage(channelID, messageID, "", []*discordgo.MessageEmbed{successEmbed}, []discordgo.MessageComponent{})
	if err != nil {
		l.Errorf(err, "failed to update discord message with success status")
	} else {
		l.Debugf("successfully updated discord message with success status")
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

// Ensure handler has Nocodb service
func init() {
	// This will be handled by dependency injection in main.go
}
