package webhook

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
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
		// Format: invoice_paid_confirm_{invoiceNumber}_{discordUserID}_{resendFlag}
		suffix := strings.TrimPrefix(customID, "invoice_paid_confirm_")
		parts := strings.Split(suffix, "_")
		if len(parts) < 1 || parts[0] == "" {
			l.Errorf(nil, "invalid invoice number in custom_id: %s", customID)
			h.respondToInteraction(c, "Invalid invoice number")
			return
		}
		invoiceNumber := parts[0]
		// Parse resendOnly flag (last part, "1" = true, "0" or missing = false)
		resendOnly := false
		if len(parts) >= 3 && parts[len(parts)-1] == "1" {
			resendOnly = true
		}
		l.Debugf("parsed invoice_paid_confirm: invoiceNumber=%s resendOnly=%v", invoiceNumber, resendOnly)
		h.handleInvoicePaidConfirmButton(c, l, interaction, invoiceNumber, resendOnly)
		return
	}

	// Payout preview button - show ephemeral confirmation
	if strings.HasPrefix(customID, "payout_preview:") {
		parts := strings.Split(customID, ":")
		if len(parts) != 6 {
			l.Errorf(nil, "invalid payout_preview custom_id format: %s", customID)
			h.respondToInteraction(c, "Invalid payout preview format")
			return
		}
		mode := parts[1]
		valueA := parts[2]
		valueB := parts[3]
		channelID := parts[4]
		allowedUserID := parts[5]

		// Validate user - only the user who ran the command can view details
		clickedUserID := interaction.Member.User.ID
		if clickedUserID != allowedUserID {
			l.Debugf("user %s attempted to view details for command run by %s", clickedUserID, allowedUserID)
			h.respondToInteraction(c, "Only the user who ran the command can view details.")
			return
		}

		if mode == "file" {
			h.handlePayoutPreviewByFileButton(c, l, interaction, valueA, valueB, channelID)
			return
		}

		batch, err := strconv.Atoi(valueB)
		if err != nil {
			l.Errorf(err, "invalid batch in payout_preview: %s", valueB)
			h.respondToInteraction(c, "Invalid batch number")
			return
		}

		h.handlePayoutPreviewButton(c, l, interaction, valueA, batch, channelID)
		return
	}

	// Payout commit confirm button
	// Supports multiple formats:
	// - 3 parts (old): payout_commit_confirm:month:batch
	// - 4 parts (current): payout_commit_confirm:month:batch:channelID
	// - 5 parts (new with mode): payout_commit_confirm:mode:valueA:valueB:channelID
	if strings.HasPrefix(customID, "payout_commit_confirm:") {
		parts := strings.Split(customID, ":")

		var mode, valueA, valueB, channelID string

		switch len(parts) {
		case 3:
			// Old format: payout_commit_confirm:month:batch
			l.Debugf("detected old format (3 parts) payout_commit_confirm custom_id: %s", customID)
			mode = "period"
			valueA = parts[1] // month
			valueB = parts[2] // batch
			channelID = interaction.ChannelID
		case 4:
			// Current format: payout_commit_confirm:month:batch:channelID
			l.Debugf("detected current format (4 parts) payout_commit_confirm custom_id: %s", customID)
			mode = "period"
			valueA = parts[1] // month
			valueB = parts[2] // batch
			channelID = parts[3]
		case 5:
			// New format: payout_commit_confirm:mode:valueA:valueB:channelID
			l.Debugf("detected new format (5 parts) payout_commit_confirm custom_id: %s", customID)
			mode = parts[1]
			valueA = parts[2]
			valueB = parts[3]
			channelID = parts[4]
		default:
			l.Errorf(nil, "invalid payout_commit_confirm custom_id format: expected 3, 4, or 5 parts, got %d. custom_id=%s", len(parts), customID)
			h.respondToInteraction(c, "Invalid payout confirm format")
			return
		}

		if mode == "file" {
			h.handlePayoutCommitConfirmByFileButton(c, l, interaction, valueA, valueB, channelID)
			return
		}

		batch, err := strconv.Atoi(valueB)
		if err != nil {
			l.Errorf(err, "invalid batch in payout_commit_confirm: %s", valueB)
			h.respondToInteraction(c, "Invalid batch number")
			return
		}
		h.handlePayoutCommitConfirmButton(c, l, interaction, valueA, batch, channelID)
		return
	}

	// Payout commit cancel button
	// Supports multiple formats:
	// - 3 parts (old): payout_commit_cancel:month:batch
	// - 4 parts (current): payout_commit_cancel:month:batch:channelID
	// - 5 parts (new with mode): payout_commit_cancel:mode:valueA:valueB:channelID
	if strings.HasPrefix(customID, "payout_commit_cancel:") {
		parts := strings.Split(customID, ":")

		var mode, valueA, valueB, channelID string

		switch len(parts) {
		case 3:
			// Old format: payout_commit_cancel:month:batch
			l.Debugf("detected old format (3 parts) payout_commit_cancel custom_id: %s", customID)
			mode = "period"
			valueA = parts[1] // month
			valueB = parts[2] // batch
			channelID = interaction.ChannelID
		case 4:
			// Current format: payout_commit_cancel:month:batch:channelID
			l.Debugf("detected current format (4 parts) payout_commit_cancel custom_id: %s", customID)
			mode = "period"
			valueA = parts[1] // month
			valueB = parts[2] // batch
			channelID = parts[3]
		case 5:
			// New format: payout_commit_cancel:mode:valueA:valueB:channelID
			l.Debugf("detected new format (5 parts) payout_commit_cancel custom_id: %s", customID)
			mode = parts[1]
			valueA = parts[2]
			valueB = parts[3]
			channelID = parts[4]
		default:
			l.Errorf(nil, "invalid payout_commit_cancel custom_id format: expected 3, 4, or 5 parts, got %d. custom_id=%s", len(parts), customID)
			h.respondToInteraction(c, "Invalid payout cancel format")
			return
		}

		if mode == "file" {
			h.handlePayoutCommitCancelByFileButton(c, l, interaction, valueA, valueB, channelID)
			return
		}

		batch, err := strconv.Atoi(valueB)
		if err != nil {
			l.Errorf(err, "invalid batch in payout_commit_cancel: %s", valueB)
			h.respondToInteraction(c, "Invalid batch number")
			return
		}
		h.handlePayoutCommitCancelButton(c, l, interaction, valueA, batch, channelID)
		return
	}

	// Extra payment preview button - show ephemeral confirmation
	// Short format: ep_p:month:discordUsername:testEmail:reasons
	if strings.HasPrefix(customID, "ep_p:") {
		parts := strings.Split(customID, ":")
		if len(parts) < 2 {
			l.Errorf(nil, "invalid ep_p custom_id format: %s", customID)
			h.respondToInteraction(c, "Invalid button format. Please run the command again.")
			return
		}

		month := parts[1]
		discordUsername := ""
		testEmail := ""
		var reasons []string

		if len(parts) >= 3 {
			discordUsername = parts[2]
		}
		if len(parts) >= 4 {
			testEmail = parts[3]
		}
		if len(parts) >= 5 && parts[4] != "" {
			reasons = strings.Split(parts[4], "|")
		}

		l.Debugf("ep_p parsed: month=%s discord=%s testEmail=%s reasons=%v", month, discordUsername, testEmail, reasons)

		// Get channelID from interaction
		channelID := interaction.ChannelID

		h.handleExtraPaymentPreviewButton(c, l, interaction, month, channelID, discordUsername, testEmail, reasons)
		return
	}

	// Extra payment confirm button
	// Short format: ep_c:month:discordUsername:testEmail:reasons
	if strings.HasPrefix(customID, "ep_c:") {
		parts := strings.Split(customID, ":")
		if len(parts) < 2 {
			l.Errorf(nil, "invalid ep_c custom_id format: %s", customID)
			h.respondToInteraction(c, "Invalid button format")
			return
		}

		month := parts[1]
		discordUsername := ""
		testEmail := ""
		var reasons []string

		if len(parts) >= 3 {
			discordUsername = parts[2]
		}
		if len(parts) >= 4 {
			testEmail = parts[3]
		}
		if len(parts) >= 5 && parts[4] != "" {
			reasons = strings.Split(parts[4], "|")
		}

		l.Debugf("ep_c parsed: month=%s discord=%s testEmail=%s reasons=%v (parts=%d)", month, discordUsername, testEmail, reasons, len(parts))

		// Get channelID from interaction
		channelID := interaction.ChannelID

		h.handleExtraPaymentConfirmButton(c, l, interaction, month, channelID, discordUsername, testEmail, reasons)
		return
	}

	// Extra payment cancel button
	// Short format: ep_x:month
	if strings.HasPrefix(customID, "ep_x:") {
		parts := strings.Split(customID, ":")
		if len(parts) < 2 {
			l.Errorf(nil, "invalid ep_x custom_id format: %s", customID)
			h.respondToInteraction(c, "Invalid button format")
			return
		}

		month := parts[1]
		channelID := interaction.ChannelID

		h.handleExtraPaymentCancelButton(c, l, interaction, month, channelID)
		return
	}

	l.Infof("unknown custom_id: %s", customID)
	h.respondToInteraction(c, "Unknown action")
}
