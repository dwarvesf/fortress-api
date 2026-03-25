package webhook

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/extrapayment"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	discordsvc "github.com/dwarvesf/fortress-api/pkg/service/discord"
)

// handleExtraPaymentPreviewButton shows ephemeral confirmation for extra payment notifications
func (h *handler) handleExtraPaymentPreviewButton(c *gin.Context, l logger.Logger, interaction *discordgo.Interaction, month, channelID, discordUsername, testEmail string, reasons []string) {
	l.Debugf("handling extra payment preview button: month=%s channelID=%s discordUsername=%s testEmail=%s", month, channelID, discordUsername, testEmail)

	// Store interaction token and app ID for later editing
	interactionToken := interaction.Token
	appID := interaction.AppID

	// Immediately respond with deferred message (must respond within 3 seconds)
	response := map[string]interface{}{
		"type": 5, // InteractionResponseDeferredChannelMessageWithSource
		"data": map[string]interface{}{
			"flags": 64, // Ephemeral
		},
	}
	c.JSON(http.StatusOK, response)

	// Process in background
	go h.processExtraPaymentPreview(l, appID, interactionToken, month, channelID, discordUsername, testEmail, reasons)
}

// processExtraPaymentPreview fetches data and updates the deferred response
func (h *handler) processExtraPaymentPreview(l logger.Logger, appID, interactionToken, month, channelID, discordUsername, testEmail string, reasons []string) {
	ctx := context.Background()

	// Get preview data
	notionService := h.service.Notion.ContractorPayouts
	if notionService == nil {
		l.Error(errors.New("contractor payouts service not initialized"), "")
		_ = h.service.Discord.EditInteractionResponse(appID, interactionToken, []*discordgo.MessageEmbed{
			{Title: "Error", Description: "Service not configured", Color: 15158332},
		})
		return
	}

	entries, err := notionService.QueryPendingExtraPayments(ctx, month, discordUsername)
	if err != nil {
		l.Errorf(err, "failed to query pending extra payments")
		_ = h.service.Discord.EditInteractionResponse(appID, interactionToken, []*discordgo.MessageEmbed{
			{Title: "Error", Description: fmt.Sprintf("Failed to get preview: %v", err), Color: 15158332},
		})
		return
	}

	if len(entries) == 0 {
		_ = h.service.Discord.EditInteractionResponse(appID, interactionToken, []*discordgo.MessageEmbed{
			{Title: "No Results", Description: fmt.Sprintf("No pending extra payments found on or before %s.", month), Color: 16776960},
		})
		return
	}

	// Group entries by contractor for display
	type contractorGroup struct {
		Name    string
		Discord string
		Total   float64
		Count   int
	}
	groups := make(map[string]*contractorGroup)
	var totalAmount float64
	for _, entry := range entries {
		amountUSD, _, convErr := extrapayment.ResolveAmountUSD(ctx, l, h.service.Wise, h.service.Redis, entry.PageID, entry.Amount, entry.Currency)
		if convErr != nil {
			_ = h.service.Discord.EditInteractionResponse(appID, interactionToken, []*discordgo.MessageEmbed{
				{Title: "Error", Description: convErr.Error(), Color: 15158332},
			})
			return
		}

		key := entry.Discord
		if key == "" {
			key = entry.ContractorName
		}
		if g, exists := groups[key]; exists {
			g.Total += amountUSD
			g.Count++
		} else {
			groups[key] = &contractorGroup{
				Name:    entry.ContractorName,
				Discord: entry.Discord,
				Total:   amountUSD,
				Count:   1,
			}
		}
		totalAmount += amountUSD
	}

	// Build contractor list (limit to 10)
	var contractorList []string
	displayCount := 0
	for _, g := range groups {
		if displayCount >= 10 {
			break
		}
		line := fmt.Sprintf("• **%s** (%s): $%.2f", g.Name, g.Discord, g.Total)
		if g.Count > 1 {
			line += fmt.Sprintf(" (%d items)", g.Count)
		}
		contractorList = append(contractorList, line)
		displayCount++
	}

	description := fmt.Sprintf("**Total Contractors:** %d\n**Total Amount:** $%.2f\n\n",
		len(groups),
		totalAmount)

	if len(contractorList) > 0 {
		description += "**Contractors:**\n" + strings.Join(contractorList, "\n")
	}

	if len(groups) > 10 {
		description += fmt.Sprintf("\n... and %d more contractors", len(groups)-10)
	}

	// Show test email info if provided
	if testEmail != "" {
		description += fmt.Sprintf("\n\n**Test Mode:** All emails will be sent to `%s`", testEmail)
	}

	// Show custom reasons if provided
	if len(reasons) > 0 {
		description += "\n\n**Custom Reasons:**\n"
		for _, r := range reasons {
			description += fmt.Sprintf("• %s\n", r)
		}
	}

	description += "\n\n**This will send email notifications to the listed contractors. Please confirm to proceed.**"

	// Build button IDs with short format (Discord limit: 100 chars)
	// Format: ep_c:month:discordUsername:testEmail:reasons
	reasonsStr := ""
	if len(reasons) > 0 {
		reasonsStr = strings.Join(reasons, "|")
		// Truncate reasons to fit in 100 char limit
		maxReasonLen := 100 - len(fmt.Sprintf("ep_c:%s:%s:%s:", month, discordUsername, testEmail))
		if maxReasonLen > 0 && len(reasonsStr) > maxReasonLen {
			reasonsStr = reasonsStr[:maxReasonLen]
		}
	}
	confirmID := fmt.Sprintf("ep_c:%s:%s:%s:%s", month, discordUsername, testEmail, reasonsStr)
	l.Debugf("building confirm button: confirmID=%s (len=%d) reasons=%v", confirmID, len(confirmID), reasons)

	cancelID := fmt.Sprintf("ep_x:%s", month)

	// Edit the deferred response with actual content
	previewEmbed := &discordgo.MessageEmbed{
		Title:       "Confirm Extra Payment Notification",
		Description: description,
		Color:       16776960, // Orange
		Footer:      &discordgo.MessageEmbedFooter{Text: "Click Confirm to send or Cancel to abort"},
		Timestamp:   time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}
	previewComponents := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Confirm",
					Style:    discordgo.SuccessButton,
					CustomID: confirmID,
					Emoji: discordgo.ComponentEmoji{
						Name: "✅",
					},
				},
				discordgo.Button{
					Label:    "Cancel",
					Style:    discordgo.DangerButton,
					CustomID: cancelID,
					Emoji: discordgo.ComponentEmoji{
						Name: "❌",
					},
				},
			},
		},
	}
	if err := h.service.Discord.EditInteractionResponseFull(appID, interactionToken, []*discordgo.MessageEmbed{previewEmbed}, previewComponents); err != nil {
		l.Errorf(err, "failed to edit interaction response with preview")
	}
}

// handleExtraPaymentConfirmButton sends extra payment notifications with progress updates
func (h *handler) handleExtraPaymentConfirmButton(c *gin.Context, l logger.Logger, interaction *discordgo.Interaction, month, channelID, discordUsername, testEmail string, reasons []string) {
	l.Debugf("handling extra payment confirm button: month=%s channelID=%s discordUsername=%s testEmail=%s", month, channelID, discordUsername, testEmail)

	// Store interaction token and app ID for later editing
	interactionToken := interaction.Token
	appID := interaction.AppID

	// Respond immediately with processing status
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Sending Extra Payment Notifications",
					Description: "Initializing...",
					Color:       5793266, // Blurple
					Timestamp:   time.Now().Format("2006-01-02T15:04:05.000-07:00"),
				},
			},
			Components: []discordgo.MessageComponent{}, // Remove buttons
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	}
	c.JSON(http.StatusOK, response)

	// Process in background
	go h.processExtraPaymentSend(l, appID, interactionToken, month, channelID, discordUsername, testEmail, reasons)
}

// processExtraPaymentSend groups entries by contractor and sends ONE email per contractor
func (h *handler) processExtraPaymentSend(l logger.Logger, appID, interactionToken, month, channelID, discordUsername, testEmail string, reasons []string) {
	ctx := context.Background()

	l.Debugf("background processing extra payment send: month=%s discordUsername=%s", month, discordUsername)

	// Build ProgressBar backed by InteractionReporter for ephemeral message updates.
	// Constructed early so early-error paths can report via typed embeds.
	reporter := discordsvc.NewInteractionReporter(h.service.Discord, appID, interactionToken)
	pb := discordsvc.NewProgressBar(reporter, l)

	// Get entries
	notionService := h.service.Notion.ContractorPayouts
	if notionService == nil {
		l.Error(errors.New("contractor payouts service not initialized"), "")
		h.updateExtraPaymentInteractionResponse(pb, 0, 0, nil, "Service not configured")
		return
	}

	entries, err := notionService.QueryPendingExtraPayments(ctx, month, discordUsername)
	if err != nil {
		l.Errorf(err, "failed to query pending extra payments")
		h.updateExtraPaymentInteractionResponse(pb, 0, 0, nil, fmt.Sprintf("Failed to get contractors: %v", err))
		return
	}

	if len(entries) == 0 {
		h.updateExtraPaymentInteractionResponse(pb, 0, 0, nil, "No pending extra payments found")
		return
	}

	// Group entries by contractor email (aggregate amounts and collect reasons)
	type contractorAggregate struct {
		Name    string
		Email   string
		Total   float64
		Reasons []string
	}
	contractors := make(map[string]*contractorAggregate)

	for _, entry := range entries {
		amountUSD, _, convErr := extrapayment.ResolveAmountUSD(ctx, l, h.service.Wise, h.service.Redis, entry.PageID, entry.Amount, entry.Currency)
		if convErr != nil {
			l.Error(convErr, "failed to resolve extra payment amount before sending Discord notification")
			h.updateExtraPaymentInteractionResponse(pb, 0, 0, nil, convErr.Error())
			return
		}

		key := entry.ContractorEmail
		if key == "" {
			key = entry.Discord // fallback to discord if no email
		}
		if key == "" {
			key = entry.ContractorName // fallback to name
		}

		if agg, exists := contractors[key]; exists {
			agg.Total += amountUSD
			if entry.Description != "" {
				agg.Reasons = append(agg.Reasons, stripInvoiceIDFromReason(entry.Description))
			}
		} else {
			var entryReasons []string
			if entry.Description != "" {
				entryReasons = []string{stripInvoiceIDFromReason(entry.Description)}
			}
			contractors[key] = &contractorAggregate{
				Name:    entry.ContractorName,
				Email:   entry.ContractorEmail,
				Total:   amountUSD,
				Reasons: entryReasons,
			}
		}
	}

	total := len(contractors)
	l.Debugf("grouped %d entries into %d contractors", len(entries), total)

	// Update with initial progress
	h.updateExtraPaymentProgress(pb, 0, total)

	// Parse month for email formatting
	monthTime, _ := time.Parse("2006-01", month)
	formattedMonth := monthTime.Format("January 2006")

	// Get Gmail service
	gmailService := h.service.GoogleMail
	if gmailService == nil {
		l.Error(errors.New("gmail service not initialized"), "")
		h.updateExtraPaymentInteractionResponse(pb, 0, 0, nil, "Email service not configured")
		return
	}

	// Send ONE email per contractor
	var sent, failed int
	var errs []string
	idx := 0

	for _, agg := range contractors {
		// Determine email recipient
		recipientEmail := agg.Email
		if testEmail != "" {
			recipientEmail = testEmail
		}

		if recipientEmail == "" {
			errMsg := fmt.Sprintf("%s: no email", agg.Name)
			errs = append(errs, errMsg)
			failed++
			l.Debugf("skipping contractor %s: no email", agg.Name)
		} else {
			// Determine reasons to use (custom reasons override entry descriptions)
			var emailReasons []string
			if len(reasons) > 0 {
				emailReasons = reasons
				l.Debugf("using custom reasons for %s: %v", agg.Name, emailReasons)
			} else {
				emailReasons = agg.Reasons
				l.Debugf("using entry descriptions for %s: %v", agg.Name, emailReasons)
			}

			// Format total amount with comma separators
			amountFormatted := formatCurrency(agg.Total)

			// Build email data with aggregated total
			emailData := &model.ExtraPaymentNotificationEmail{
				ContractorName:  agg.Name,
				ContractorEmail: recipientEmail,
				Month:           formattedMonth,
				Amount:          agg.Total,
				AmountFormatted: amountFormatted,
				Reasons:         emailReasons,
				SenderName:      "Team Dwarves",
			}

			// Send email
			if err := gmailService.SendExtraPaymentNotificationMail(emailData); err != nil {
				errMsg := fmt.Sprintf("%s: %v", agg.Name, err)
				errs = append(errs, errMsg)
				failed++
				l.Debugf("failed to send to %s: %v", agg.Name, err)
			} else {
				sent++
				l.Debugf("sent to %s (%s) - total: %s", agg.Name, recipientEmail, amountFormatted)
			}
		}

		idx++
		// Update progress
		h.updateExtraPaymentProgress(pb, idx, total)
	}

	// Show final result
	h.updateExtraPaymentInteractionResponse(pb, sent, failed, errs, "")
}

// updateExtraPaymentProgress reports the current send progress via ProgressBar.
func (h *handler) updateExtraPaymentProgress(pb *discordsvc.ProgressBar, current, total int) {
	if pb == nil {
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Sending Extra Payment Notifications",
		Description: fmt.Sprintf("Sending notifications...\n\n%s", discordsvc.BuildBar(current, total)),
		Color:       5793266, // Blurple
	}

	pb.Report(embed)
}

// updateExtraPaymentInteractionResponse reports the final send result via ProgressBar.
func (h *handler) updateExtraPaymentInteractionResponse(pb *discordsvc.ProgressBar, sent, failed int, errs []string, errorMsg string) {
	if pb == nil {
		return
	}

	var embed *discordgo.MessageEmbed

	if errorMsg != "" {
		embed = &discordgo.MessageEmbed{
			Title:       "❌ Extra Payment Notification Failed",
			Description: errorMsg,
			Color:       15158332, // Red
			Timestamp:   time.Now().Format("2006-01-02T15:04:05.000-07:00"),
		}
	} else if failed == 0 && sent > 0 {
		embed = &discordgo.MessageEmbed{
			Title:       "✅ Extra Payment Notifications Sent",
			Description: fmt.Sprintf("Successfully sent **%d** notification emails.", sent),
			Color:       3066993, // Green
			Timestamp:   time.Now().Format("2006-01-02T15:04:05.000-07:00"),
		}
	} else if sent > 0 {
		description := fmt.Sprintf("**Sent:** %d\n**Failed:** %d\n\n", sent, failed)

		if len(errs) > 0 {
			description += "**Errors:**\n"
			displayCount := len(errs)
			if displayCount > 5 {
				displayCount = 5
			}
			for i := 0; i < displayCount; i++ {
				description += fmt.Sprintf("• %s\n", errs[i])
			}
			if len(errs) > 5 {
				description += fmt.Sprintf("... and %d more errors\n", len(errs)-5)
			}
		}

		embed = &discordgo.MessageEmbed{
			Title:       "⚠️ Extra Payment Notifications Completed with Errors",
			Description: description,
			Color:       16776960, // Orange
			Timestamp:   time.Now().Format("2006-01-02T15:04:05.000-07:00"),
		}
	} else {
		description := "No notifications were sent.\n\n"
		if len(errs) > 0 {
			description += "**Errors:**\n"
			for _, e := range errs {
				description += fmt.Sprintf("• %s\n", e)
			}
		}
		embed = &discordgo.MessageEmbed{
			Title:       "❌ Extra Payment Notifications Failed",
			Description: description,
			Color:       15158332, // Red
			Timestamp:   time.Now().Format("2006-01-02T15:04:05.000-07:00"),
		}
	}

	pb.Report(embed)
}

// handleExtraPaymentCancelButton handles cancellation of extra payment notification
func (h *handler) handleExtraPaymentCancelButton(c *gin.Context, l logger.Logger, interaction *discordgo.Interaction, month, channelID string) {
	l.Debugf("handling extra payment cancel button: month=%s channelID=%s", month, channelID)

	cancelEmbed := &discordgo.MessageEmbed{
		Title:       "Extra Payment Notification Cancelled",
		Description: fmt.Sprintf("Extra payment notification for **%s** has been cancelled.", month),
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
