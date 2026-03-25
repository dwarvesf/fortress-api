package webhook

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/nocodb"
)

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
