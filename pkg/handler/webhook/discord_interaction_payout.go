package webhook

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
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
	interactionToken := interaction.Token
	appID := interaction.AppID

	response := map[string]interface{}{
		"type": 5,
		"data": map[string]interface{}{
			"flags": 64,
		},
	}
	c.JSON(http.StatusOK, response)

	go h.processPayoutPreview(l, appID, interactionToken, month, batch, channelID)
	return
}

func (h *handler) processPayoutPreview(l logger.Logger, appID, interactionToken, month string, batch int, channelID string) {

	// Try to get preview from cache first (cached when ?payout commit was called)
	preview, found := h.controller.ContractorPayables.GetCachedPreview(month, batch)
	if !found {
		l.Debugf("preview not found in cache for month=%s batch=%d", month, batch)
		_ = h.service.Discord.EditInteractionResponse(appID, interactionToken, []*discordgo.MessageEmbed{{
			Title:       "Preview Expired",
			Description: fmt.Sprintf("Preview expired or not found. Please run `?payout commit %s %d` again.", month, batch),
			Color:       16776960,
		}})
		return
	}

	l.Debugf("found cached preview: count=%d total=%.2f", preview.Count, preview.TotalAmount)

	if preview.Count == 0 {
		_ = h.service.Discord.EditInteractionResponse(appID, interactionToken, []*discordgo.MessageEmbed{{
			Title:       "No Pending Payables",
			Description: fmt.Sprintf("No pending payables found for %s (batch %d).", month, batch),
			Color:       5793266,
		}})
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

	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{Components: []discordgo.MessageComponent{
					discordgo.Button{Label: "Confirm", Style: discordgo.SuccessButton, CustomID: fmt.Sprintf("payout_commit_confirm:%s:%d:%s", month, batch, channelID), Emoji: discordgo.ComponentEmoji{Name: "✅"}},
					discordgo.Button{Label: "Cancel", Style: discordgo.DangerButton, CustomID: fmt.Sprintf("payout_commit_cancel:%s:%d:%s", month, batch, channelID), Emoji: discordgo.ComponentEmoji{Name: "❌"}},
				}},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	}

	l.Debugf("sending payout preview response")
	if err := h.service.Discord.EditInteractionResponseFull(appID, interactionToken, []*discordgo.MessageEmbed{embed}, response.Data.Components); err != nil {
		l.Errorf(err, "failed to edit payout preview response")
	}
}

func (h *handler) handlePayoutPreviewByFileButton(c *gin.Context, l logger.Logger, interaction *discordgo.Interaction, fileName string, year string, channelID string) {
	l.Debugf("handling payout preview button by file: fileName=%s year=%s channelID=%s user=%s", fileName, year, channelID, interaction.Member.User.Username)
	interactionToken := interaction.Token
	appID := interaction.AppID

	response := map[string]interface{}{
		"type": 5,
		"data": map[string]interface{}{
			"flags": 64,
		},
	}
	c.JSON(http.StatusOK, response)

	go h.processPayoutPreviewByFile(l, appID, interactionToken, fileName, year, channelID)
	return
}

func (h *handler) processPayoutPreviewByFile(l logger.Logger, appID, interactionToken, fileName string, year string, channelID string) {

	yearValue, err := strconv.Atoi(year)
	if err != nil {
		l.Errorf(err, "invalid year in payout preview by file: %s", year)
		_ = h.service.Discord.EditInteractionResponse(appID, interactionToken, []*discordgo.MessageEmbed{{Title: "Error", Description: "Invalid year", Color: 15158332}})
		return
	}

	preview, err := h.controller.ContractorPayables.PreviewCommitByFile(context.Background(), fileName, yearValue)
	if err != nil {
		l.Errorf(err, "failed to preview payout by file")
		_ = h.service.Discord.EditInteractionResponse(appID, interactionToken, []*discordgo.MessageEmbed{{Title: "Error", Description: fmt.Sprintf("Failed to get preview: %v", err), Color: 15158332}})
		return
	}

	if preview.Count == 0 {
		_ = h.service.Discord.EditInteractionResponse(appID, interactionToken, []*discordgo.MessageEmbed{{Title: "No Pending Payables", Description: fmt.Sprintf("No pending payables found for file %s in year %d.", fileName, yearValue), Color: 5793266}})
		return
	}

	contractorList := make([]string, 0, len(preview.Contractors))
	displayCount := len(preview.Contractors)
	if displayCount > 10 {
		displayCount = 10
	}
	for i := 0; i < displayCount; i++ {
		contractor := preview.Contractors[i]
		contractorList = append(contractorList, fmt.Sprintf("• **%s**: %s %.2f", contractor.Name, contractor.Currency, contractor.Amount))
	}

	description := fmt.Sprintf("**Mode:** file\n**File:** %s\n**Year:** %d\n**Total Contractors:** %d\n**Total Amount:** $%.2f\n\n", preview.FileName, preview.Year, preview.Count, preview.TotalAmount)
	if len(contractorList) > 0 {
		description += "**Contractors:**\n" + strings.Join(contractorList, "\n")
	}
	if len(preview.Contractors) > 10 {
		description += fmt.Sprintf("\n... and %d more contractors", len(preview.Contractors)-10)
	}
	description += "\n\n⚠️ **This action will update Notion databases. Please confirm to proceed.**"

	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{{
				Title:       "Payout Commit Confirmation",
				Description: description,
				Color:       16776960,
				Footer: &discordgo.MessageEmbedFooter{
					Text: "Click Confirm to commit or Cancel to abort",
				},
				Timestamp: time.Now().Format("2006-01-02T15:04:05.000-07:00"),
			}},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{Components: []discordgo.MessageComponent{
					discordgo.Button{Label: "Confirm", Style: discordgo.SuccessButton, CustomID: fmt.Sprintf("payout_commit_confirm:file:%s:%d:%s", fileName, yearValue, channelID), Emoji: discordgo.ComponentEmoji{Name: "✅"}},
					discordgo.Button{Label: "Cancel", Style: discordgo.DangerButton, CustomID: fmt.Sprintf("payout_commit_cancel:file:%s:%d:%s", fileName, yearValue, channelID), Emoji: discordgo.ComponentEmoji{Name: "❌"}},
				}},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	}

	if err := h.service.Discord.EditInteractionResponseFull(appID, interactionToken, response.Data.Embeds, response.Data.Components); err != nil {
		l.Errorf(err, "failed to edit payout preview by file response")
	}
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

func (h *handler) handlePayoutCommitConfirmByFileButton(c *gin.Context, l logger.Logger, interaction *discordgo.Interaction, fileName string, year string, channelID string) {
	l.Debugf("handling payout commit confirm by file: fileName=%s year=%s channelID=%s user=%s", fileName, year, channelID, interaction.Member.User.Username)

	yearValue, err := strconv.Atoi(year)
	if err != nil {
		l.Errorf(err, "invalid year in payout commit confirm by file: %s", year)
		h.respondToInteraction(c, "Invalid year")
		return
	}

	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{{
				Title:       "Processing Payout Commit",
				Description: fmt.Sprintf("Committing payables for file **%s** (year %d)...\n\n⏳ Please wait...", fileName, yearValue),
				Color:       16776960,
				Timestamp:   time.Now().Format("2006-01-02T15:04:05.000-07:00"),
			}},
			Components: []discordgo.MessageComponent{},
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	}
	c.JSON(http.StatusOK, response)

	go h.processPayoutCommitByFile(l, interaction.AppID, interaction.Token, fileName, yearValue, channelID)
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

func (h *handler) processPayoutCommitByFile(l logger.Logger, appID, interactionToken, fileName string, year int, channelID string) {
	ctx := context.Background()

	l.Debugf("background processing payout commit by file: fileName=%s year=%d channelID=%s appID=%s", fileName, year, channelID, appID)

	result, err := h.controller.ContractorPayables.CommitPayablesByFile(ctx, fileName, year)
	if err != nil {
		l.Errorf(err, "failed to execute payout commit by file")
		h.updatePayoutInteractionResponseByFile(l, appID, interactionToken, fileName, year, nil, err.Error())
		return
	}

	l.Infof("payout commit by file completed: fileName=%s year=%d updated=%d failed=%d", fileName, year, result.Updated, result.Failed)
	h.updatePayoutInteractionResponseByFile(l, appID, interactionToken, fileName, year, result, "")
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

func (h *handler) updatePayoutInteractionResponseByFile(l logger.Logger, appID, interactionToken, fileName string, year int, result interface{}, errorMsg string) {
	l.Debugf("updating payout interaction response by file: appID=%s fileName=%s year=%d", appID, fileName, year)

	var embed *discordgo.MessageEmbed
	if errorMsg != "" {
		embed = &discordgo.MessageEmbed{
			Title:       "❌ Payout Commit Failed",
			Description: fmt.Sprintf("**Mode:** file\n**File:** %s\n**Year:** %d\n**Error:** %s", fileName, year, errorMsg),
			Color:       15158332,
			Timestamp:   time.Now().Format("2006-01-02T15:04:05.000-07:00"),
		}
	} else if commitResult, ok := result.(*ctrlcontractorpayables.CommitResponse); ok {
		title := "✅ Payout Commit Successful"
		description := fmt.Sprintf("**Mode:** file\n**File:** %s\n**Year:** %d\n**Updated:** %d contractors\n\nAll payables have been successfully committed.", commitResult.FileName, commitResult.Year, commitResult.Updated)
		color := 3066993
		if commitResult.Failed > 0 {
			title = "⚠️ Payout Commit Completed with Errors"
			description = fmt.Sprintf("**Mode:** file\n**File:** %s\n**Year:** %d\n**Updated:** %d contractors\n**Failed:** %d contractors\n\n", commitResult.FileName, commitResult.Year, commitResult.Updated, commitResult.Failed)
			if len(commitResult.Errors) > 0 {
				description += "**Errors:**\n"
				displayCount := len(commitResult.Errors)
				if displayCount > 5 {
					displayCount = 5
				}
				for i := 0; i < displayCount; i++ {
					err := commitResult.Errors[i]
					subject := err.PayableID
					if err.InvoiceID != "" {
						subject = err.InvoiceID
					}
					description += fmt.Sprintf("• `%s`: %s\n", subject, err.Error)
				}
			}
			color = 16776960
		}

		embed = &discordgo.MessageEmbed{Title: title, Description: description, Color: color, Timestamp: time.Now().Format("2006-01-02T15:04:05.000-07:00")}
	}

	if embed == nil {
		return
	}
	if err := h.service.Discord.EditInteractionResponseFull(appID, interactionToken, []*discordgo.MessageEmbed{embed}, []discordgo.MessageComponent{}); err != nil {
		l.Errorf(err, "failed to update payout interaction response by file")
	}
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

func (h *handler) handlePayoutCommitCancelByFileButton(c *gin.Context, l logger.Logger, interaction *discordgo.Interaction, fileName string, year string, channelID string) {
	l.Debugf("handling payout commit cancel by file: fileName=%s year=%s channelID=%s user=%s", fileName, year, channelID, interaction.Member.User.Username)

	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{{
				Title:       "Payout Commit Canceled",
				Description: fmt.Sprintf("Payout commit for file **%s** (year %s) has been canceled.", fileName, year),
				Color:       5793266,
				Timestamp:   time.Now().Format("2006-01-02T15:04:05.000-07:00"),
			}},
			Components: []discordgo.MessageComponent{},
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	}

	c.JSON(http.StatusOK, response)
}
