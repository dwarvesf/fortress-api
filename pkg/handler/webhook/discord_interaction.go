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
	"github.com/dwarvesf/fortress-api/pkg/service/nocodb"
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
