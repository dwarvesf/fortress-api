package webhook

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"

	ctrlcontractorpayables "github.com/dwarvesf/fortress-api/pkg/controller/contractorpayables"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

// handlePayoutPreviewButton shows ephemeral confirmation with payout preview details
func (h *handler) handlePayoutPreviewButton(c *gin.Context, l logger.Logger, interaction *discordgo.Interaction, month string, batch int, channelID string) {
	l.Debugf("handling payout preview button: month=%s batch=%d channelID=%s user=%s", month, batch, channelID, interaction.Member.User.Username)

	// Try to get preview from cache first (cached when ?payout commit was called)
	preview, found := h.controller.ContractorPayables.GetCachedPreview(month, batch)
	if !found {
		l.Debugf("preview not found in cache for month=%s batch=%d", month, batch)
		h.respondToInteraction(c, fmt.Sprintf("Preview expired or not found. Please run `?payout commit %s %d` again.", month, batch))
		return
	}

	l.Debugf("found cached preview: count=%d total=%.2f", preview.Count, preview.TotalAmount)

	if preview.Count == 0 {
		h.respondToInteraction(c, fmt.Sprintf("No pending payables found for %s (batch %d).", month, batch))
		return
	}

	// Build contractor list (limit to 10 to avoid embed size limit)
	contractorList := make([]string, 0, len(preview.Contractors))
	displayCount := len(preview.Contractors)
	if displayCount > 10 {
		displayCount = 10
	}

	for i := 0; i < displayCount; i++ {
		contractor := preview.Contractors[i]
		contractorList = append(contractorList, fmt.Sprintf("• **%s**: %s %.2f",
			contractor.Name,
			contractor.Currency,
			contractor.Amount))
	}

	description := fmt.Sprintf("**Month:** %s\n**Batch:** %d\n**Total Contractors:** %d\n**Total Amount:** $%.2f\n\n",
		preview.Month,
		preview.Batch,
		preview.Count,
		preview.TotalAmount)

	if len(contractorList) > 0 {
		description += "**Contractors:**\n" + strings.Join(contractorList, "\n")
	}

	if len(preview.Contractors) > 10 {
		description += fmt.Sprintf("\n... and %d more contractors", len(preview.Contractors)-10)
	}

	description += "\n\n⚠️ **This action will update Notion databases. Please confirm to proceed.**"

	embed := &discordgo.MessageEmbed{
		Title:       "Payout Commit Confirmation",
		Description: description,
		Color:       16776960, // Orange
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Click Confirm to commit or Cancel to abort",
		},
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	// Build response using raw maps to ensure proper JSON serialization for Discord HTTP API
	// discordgo types may not serialize correctly for HTTP webhook responses
	response := map[string]interface{}{
		"type": 4, // InteractionResponseChannelMessageWithSource
		"data": map[string]interface{}{
			"embeds": []map[string]interface{}{
				{
					"title":       embed.Title,
					"description": embed.Description,
					"color":       embed.Color,
					"footer": map[string]interface{}{
						"text": embed.Footer.Text,
					},
					"timestamp": embed.Timestamp,
				},
			},
			"components": []map[string]interface{}{
				{
					"type": 1, // ActionsRow
					"components": []map[string]interface{}{
						{
							"type":      2, // Button
							"label":     "Confirm",
							"style":     3, // Success (green)
							"custom_id": fmt.Sprintf("payout_commit_confirm:%s:%d:%s", month, batch, channelID),
						},
						{
							"type":      2, // Button
							"label":     "Cancel",
							"style":     4, // Danger (red)
							"custom_id": fmt.Sprintf("payout_commit_cancel:%s:%d:%s", month, batch, channelID),
						},
					},
				},
			},
			"flags": 64, // Ephemeral
		},
	}

	l.Debugf("sending payout preview response")
	c.JSON(http.StatusOK, response)
}

// handlePayoutCommitConfirmButton executes the payout commit
func (h *handler) handlePayoutCommitConfirmButton(c *gin.Context, l logger.Logger, interaction *discordgo.Interaction, month string, batch int, channelID string) {
	l.Debugf("handling payout commit confirm: month=%s batch=%d channelID=%s user=%s", month, batch, channelID, interaction.Member.User.Username)

	// Store interaction token and app ID for later editing
	interactionToken := interaction.Token
	appID := interaction.AppID

	// Respond immediately with processing status
	processingEmbed := &discordgo.MessageEmbed{
		Title:       "Processing Payout Commit",
		Description: fmt.Sprintf("Committing payables for **%s** (batch %d)...\n\n⏳ Please wait...", month, batch),
		Color:       16776960, // Yellow/Orange
		Timestamp:   time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{processingEmbed},
			Components: []discordgo.MessageComponent{}, // Remove buttons
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	}
	c.JSON(http.StatusOK, response)

	// Process in background - pass interaction token and app ID to update the same message
	go h.processPayoutCommit(l, appID, interactionToken, month, batch, channelID)
}

// processPayoutCommit executes the commit and updates the original interaction response
func (h *handler) processPayoutCommit(l logger.Logger, appID, interactionToken, month string, batch int, channelID string) {
	ctx := context.Background()

	l.Debugf("background processing payout commit: month=%s batch=%d channelID=%s appID=%s", month, batch, channelID, appID)

	// Execute commit
	result, err := h.controller.ContractorPayables.CommitPayables(ctx, month, batch, "")
	if err != nil {
		l.Errorf(err, "failed to execute payout commit")
		h.updatePayoutInteractionResponse(l, appID, interactionToken, month, batch, nil, err.Error())
		return
	}

	l.Infof("payout commit completed: month=%s batch=%d updated=%d failed=%d", month, batch, result.Updated, result.Failed)
	h.updatePayoutInteractionResponse(l, appID, interactionToken, month, batch, result, "")
}

// updatePayoutInteractionResponse edits the original interaction response with the result
func (h *handler) updatePayoutInteractionResponse(l logger.Logger, appID, interactionToken, month string, batch int, result interface{}, errorMsg string) {
	l.Debugf("updating payout interaction response: appID=%s month=%s batch=%d", appID, month, batch)

	var embed *discordgo.MessageEmbed

	if errorMsg != "" {
		embed = &discordgo.MessageEmbed{
			Title:       "❌ Payout Commit Failed",
			Description: fmt.Sprintf("**Month:** %s\n**Batch:** %d\n**Error:** %s", month, batch, errorMsg),
			Color:       15158332, // Red
			Timestamp:   time.Now().Format("2006-01-02T15:04:05.000-07:00"),
		}
	} else if commitResult, ok := result.(*ctrlcontractorpayables.CommitResponse); ok {
		var title string
		var description string
		var color int

		if commitResult.Failed == 0 {
			title = "✅ Payout Commit Successful"
			description = fmt.Sprintf("**Month:** %s\n**Batch:** %d\n**Updated:** %d contractors\n\nAll payables have been successfully committed.",
				commitResult.Month, commitResult.Batch, commitResult.Updated)
			color = 3066993 // Green
		} else {
			title = "⚠️ Payout Commit Completed with Errors"
			description = fmt.Sprintf("**Month:** %s\n**Batch:** %d\n**Updated:** %d contractors\n**Failed:** %d contractors\n\n",
				commitResult.Month, commitResult.Batch, commitResult.Updated, commitResult.Failed)

			// Show error details (limit to 5)
			if len(commitResult.Errors) > 0 {
				description += "**Errors:**\n"
				displayCount := len(commitResult.Errors)
				if displayCount > 5 {
					displayCount = 5
				}
				for i := 0; i < displayCount; i++ {
					err := commitResult.Errors[i]
					description += fmt.Sprintf("• Payable ID `%s`: %s\n", err.PayableID, err.Error)
				}
				if len(commitResult.Errors) > 5 {
					description += fmt.Sprintf("... and %d more errors\n", len(commitResult.Errors)-5)
				}
			}
			color = 16776960 // Orange
		}

		embed = &discordgo.MessageEmbed{
			Title:       title,
			Description: description,
			Color:       color,
			Timestamp:   time.Now().Format("2006-01-02T15:04:05.000-07:00"),
		}
	}

	if embed == nil {
		l.Warnf("no embed to send for payout result")
		return
	}

	// Edit with empty components to remove any remaining buttons
	if err := h.service.Discord.EditInteractionResponseFull(appID, interactionToken, []*discordgo.MessageEmbed{embed}, []discordgo.MessageComponent{}); err != nil {
		l.Errorf(err, "failed to update payout interaction response")
		return
	}

	l.Debugf("successfully updated payout interaction response")
}

// handlePayoutCommitCancelButton handles cancellation of payout commit
func (h *handler) handlePayoutCommitCancelButton(c *gin.Context, l logger.Logger, interaction *discordgo.Interaction, month string, batch int, channelID string) {
	l.Debugf("handling payout commit cancel: month=%s batch=%d channelID=%s user=%s", month, batch, channelID, interaction.Member.User.Username)

	cancelEmbed := &discordgo.MessageEmbed{
		Title:       "Payout Commit Canceled",
		Description: fmt.Sprintf("Payout commit for **%s** (batch %d) has been canceled.", month, batch),
		Color:       5793266, // Blue
		Timestamp:   time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{cancelEmbed},
			Components: []discordgo.MessageComponent{}, // Remove buttons
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	}

	c.JSON(http.StatusOK, response)
}
