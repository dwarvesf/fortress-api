package webhook

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"

	invoiceCtrl "github.com/dwarvesf/fortress-api/pkg/controller/invoice"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

// handleInvoicePaidConfirmButton handles the invoice paid confirm button click
func (h *handler) handleInvoicePaidConfirmButton(c *gin.Context, l logger.Logger, interaction *discordgo.Interaction, invoiceNumber string, resendOnly bool) {
	l.Debugf("confirming invoice payment via button: invoiceNumber=%s user=%s resendOnly=%v", invoiceNumber, interaction.Member.User.Username, resendOnly)

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
	go h.processInvoicePaidConfirm(l, channelID, messageID, invoiceNumber, actionBy, resendOnly)
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
func (h *handler) processInvoicePaidConfirm(l logger.Logger, channelID, messageID, invoiceNumber, actionBy string, resendOnly bool) {
	l.Debugf("background processing invoice payment: invoiceNumber=%s channelID=%s messageID=%s resendOnly=%v", invoiceNumber, channelID, messageID, resendOnly)

	// Mark invoice as paid - searches both PostgreSQL and Notion
	result, err := h.controller.Invoice.MarkInvoiceAsPaidByNumber(invoiceNumber, resendOnly)
	if err != nil {
		l.Errorf(err, "failed to mark invoice as paid: %s", invoiceNumber)
		h.updateInvoiceMessageWithError(l, channelID, messageID, invoiceNumber, err.Error())
		return
	}

	l.Infof("invoice marked as paid via discord button: invoiceNumber=%s user=%s source=%s resendOnly=%v", invoiceNumber, actionBy, result.Source, resendOnly)

	// Log to Discord audit
	if err := h.controller.Discord.Log(model.LogDiscordInput{
		Type: "invoice_paid",
		Data: map[string]interface{}{
			"invoice_number": invoiceNumber,
			"source":         result.Source,
			"resend_only":    resendOnly,
		},
	}); err != nil {
		l.Errorf(err, "failed to log invoice paid to discord")
	}

	// Update the message to show success
	h.updateInvoiceMessageWithResult(l, channelID, messageID, result, actionBy, resendOnly)
}

// updateInvoiceMessageWithResult updates the Discord message with MarkPaidResult
func (h *handler) updateInvoiceMessageWithResult(l logger.Logger, channelID, messageID string, result *invoiceCtrl.MarkPaidResult, actionBy string, resendOnly bool) {
	l.Debugf("updating invoice message with result: invoiceNumber=%s source=%s actionBy=%s channelID=%s messageID=%s resendOnly=%v", result.InvoiceNumber, result.Source, actionBy, channelID, messageID, resendOnly)

	if channelID == "" || messageID == "" {
		l.Warnf("cannot update message: missing channelID or messageID")
		return
	}

	var title, statusValue string
	if resendOnly {
		title = fmt.Sprintf("Invoice %s - Email Resent", result.InvoiceNumber)
		statusValue = "✅ Thank you email resent"
	} else {
		title = fmt.Sprintf("Invoice %s", result.InvoiceNumber)
		statusValue = "✅ Paid"
	}

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "Status",
			Value:  statusValue,
			Inline: false,
		},
		{
			Name:   "Marked by",
			Value:  actionBy,
			Inline: false,
		},
	}

	successEmbed := &discordgo.MessageEmbed{
		Title:     title,
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
