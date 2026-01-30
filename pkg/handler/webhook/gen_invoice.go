package webhook

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/service/ratelimit"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

// GenInvoiceRequest represents the request body for the gen-invoice webhook
type GenInvoiceRequest struct {
	DiscordUsername string `json:"discord_username" binding:"required"`
	Month           string `json:"month" binding:"required"` // Format: YYYY-MM
	DMChannelID     string `json:"dm_channel_id" binding:"required"`
	DMMessageID     string `json:"dm_message_id" binding:"required"`
}

// GenInvoiceResponse represents the response for the gen-invoice webhook
type GenInvoiceResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// invoiceRateLimiter is the rate limiter instance for invoice generation
// This will be set during server initialization
var invoiceRateLimiter ratelimit.IInvoiceRateLimiter

// SetInvoiceRateLimiter sets the rate limiter instance for the webhook handler
func SetInvoiceRateLimiter(rl ratelimit.IInvoiceRateLimiter) {
	invoiceRateLimiter = rl
}

// HandleGenInvoice handles the Discord gen-invoice webhook
// POST /webhooks/discord/gen-invoice
func (h *handler) HandleGenInvoice(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "webhook",
		"method":  "HandleGenInvoice",
	})

	l.Debug("received gen-invoice webhook request")

	// Parse request
	var req GenInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		l.Errorf(err, "failed to parse gen-invoice request")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("parsed request: discord_username=%s month=%s dm_channel_id=%s dm_message_id=%s",
		req.DiscordUsername, req.Month, req.DMChannelID, req.DMMessageID))

	// Validate month format (YYYY-MM)
	if !isValidMonthFormat(req.Month) {
		l.Debug(fmt.Sprintf("invalid month format: %s", req.Month))
		c.JSON(http.StatusBadRequest, GenInvoiceResponse{
			Success: false,
			Message: "Invalid month format. Expected YYYY-MM",
		})
		return
	}

	// Check rate limit
	if invoiceRateLimiter == nil {
		l.Error(nil, "invoice rate limiter not initialized")
		c.JSON(http.StatusInternalServerError, GenInvoiceResponse{
			Success: false,
			Message: "Rate limiter not configured",
		})
		return
	}

	if err := invoiceRateLimiter.CheckLimit(req.DiscordUsername); err != nil {
		l.Debug(fmt.Sprintf("rate limit exceeded for user %s: %v", req.DiscordUsername, err))
		remaining := invoiceRateLimiter.GetRemainingAttempts(req.DiscordUsername)
		resetTime := invoiceRateLimiter.GetResetTime(req.DiscordUsername)

		// Update DM with rate limit error
		go h.updateDMWithRateLimitError(l, req.DMChannelID, req.DMMessageID, req.DiscordUsername, remaining, resetTime)

		c.JSON(http.StatusTooManyRequests, GenInvoiceResponse{
			Success: false,
			Message: fmt.Sprintf("Rate limit exceeded. %d/%d attempts remaining. Resets at %s",
				remaining, ratelimit.MaxInvoiceGenerationsPerDay, resetTime.Format("15:04 UTC")),
		})
		return
	}

	// Return 200 OK immediately (async pattern)
	c.JSON(http.StatusOK, GenInvoiceResponse{
		Success: true,
		Message: "Invoice generation started",
	})

	// Process invoice generation asynchronously
	go h.processGenInvoice(l, req)
}

// processGenInvoice handles the async payable lookup process
func (h *handler) processGenInvoice(l logger.Logger, req GenInvoiceRequest) {
	ctx := context.Background()

	l.Debug(fmt.Sprintf("starting payable lookup: discord_username=%s month=%s", req.DiscordUsername, req.Month))

	// Step 1: Query contractor rates
	l.Debug("step 1: fetching contractor rate data")
	ratesService := notion.NewContractorRatesService(h.config, h.logger)
	if ratesService == nil {
		l.Error(nil, "failed to create contractor rates service")
		h.updateDMWithError(l, req.DMChannelID, req.DMMessageID, "Failed to lookup contractor information")
		return
	}

	rateData, err := ratesService.QueryRatesByDiscordAndMonth(ctx, req.DiscordUsername, req.Month)
	if err != nil {
		l.Errorf(err, "failed to query contractor rates")
		h.updateDMWithError(l, req.DMChannelID, req.DMMessageID, "No active contractor rate found. Please contact HR.")
		return
	}

	l.Debug(fmt.Sprintf("found contractor: pageID=%s payDay=%d", rateData.ContractorPageID, rateData.PayDay))

	// Step 2: Calculate period dates
	monthTime, _ := time.Parse("2006-01", req.Month)
	payDay := rateData.PayDay
	if payDay == 0 {
		payDay = 15 // Default to 15
	}
	periodStart := time.Date(monthTime.Year(), monthTime.Month(), payDay, 0, 0, 0, 0, time.UTC)

	l.Debug(fmt.Sprintf("calculated period start: %s", periodStart.Format("2006-01-02")))

	// Step 3: Lookup existing pending payables
	l.Debug("step 3: looking up pending payables")
	payables, err := h.service.Notion.ContractorPayables.FindPayableByContractorAndPeriod(ctx, rateData.ContractorPageID, periodStart.Format("2006-01-02"))
	if err != nil {
		l.Errorf(err, "failed to lookup payables")
		h.updateDMWithError(l, req.DMChannelID, req.DMMessageID, "Failed to lookup invoice. Please try again later.")
		return
	}

	// Step 4: Check if any pending payables found
	if len(payables) == 0 {
		l.Debug("no pending payables found - invoice being processed")
		h.updateDMWithNotReady(l, req.DMChannelID, req.DMMessageID, req.Month)
		return
	}

	l.Debug(fmt.Sprintf("found %d pending payable(s)", len(payables)))

	// Step 5: Get contractor personal email (once for all payables)
	l.Debug("step 5: fetching contractor personal email")
	taskOrderLogService := h.service.Notion.TaskOrderLog
	personalEmail := taskOrderLogService.GetContractorPersonalEmail(ctx, rateData.ContractorPageID)
	if personalEmail == "" {
		l.Debug(fmt.Sprintf("personal email not found for contractor: %s", rateData.ContractorPageID))
		h.updateDMWithError(l, req.DMChannelID, req.DMMessageID, "Personal email not found. Please contact HR to update your profile.")
		return
	}

	l.Debug(fmt.Sprintf("found personal email: %s", personalEmail))

	// Step 6: Process each pending payable separately
	l.Debug(fmt.Sprintf("step 6: processing %d payable(s)", len(payables)))
	successCount := 0
	contractorName := rateData.ContractorName

	for i, payable := range payables {
		l.Debug(fmt.Sprintf("processing payable %d/%d: pageID=%s invoiceID=%s hasFileURL=%v",
			i+1, len(payables), payable.PageID, payable.InvoiceID, payable.FileURL != ""))

		// Check if PDF is available (either in FileURL or we can search for it)
		if payable.InvoiceID == "" {
			l.Debug(fmt.Sprintf("payable %s has no invoice ID - skipping", payable.PageID))
			continue
		}

		// Search for the file in Google Drive by contractor name and invoice ID
		// Files are stored under: ContractorInvoiceDirID/<contractor_name>/<invoice_id>.pdf
		l.Debug(fmt.Sprintf("searching for invoice file in Google Drive: contractorName=%s invoiceID=%s", contractorName, payable.InvoiceID))

		fileID, err := h.service.GoogleDrive.FindContractorInvoiceFileID(contractorName, payable.InvoiceID)
		if err != nil {
			l.Errorf(err, fmt.Sprintf("failed to search Google Drive for payable %s", payable.PageID))
			// If search fails, fall back to showing Notion link if available
			if payable.FileURL != "" {
				h.updateDMWithSuccessNoShare(l, req.DMChannelID, req.DMMessageID, req.Month, payable)
				successCount++
			}
			continue
		}

		if fileID == "" {
			l.Debug(fmt.Sprintf("invoice file not found in Google Drive for payable %s", payable.PageID))
			// File not in Google Drive yet, show Notion link if available
			if payable.FileURL != "" {
				h.updateDMWithSuccessNoShare(l, req.DMChannelID, req.DMMessageID, req.Month, payable)
				successCount++
			} else {
				l.Debug(fmt.Sprintf("payable %s has no file URL and not found in Google Drive - skipping", payable.PageID))
			}
			continue
		}

		l.Debug(fmt.Sprintf("found file in Google Drive: fileID=%s", fileID))

		// Update payable FileURL with Google Drive URL for display
		googleDriveURL := fmt.Sprintf("https://drive.google.com/file/d/%s/view", fileID)
		payable.FileURL = googleDriveURL

		// Share file with contractor email
		if err := h.service.GoogleDrive.ShareFileWithEmail(fileID, personalEmail); err != nil {
			l.Errorf(err, fmt.Sprintf("failed to share file for payable %s", payable.PageID))
			// Non-fatal: file exists, just couldn't share. Update with partial success
			h.updateDMWithPartialSuccess(l, req.DMChannelID, req.DMMessageID, payable, personalEmail, err)
			successCount++
			continue
		}

		l.Debug(fmt.Sprintf("file shared with email: %s", personalEmail))

		// Update DM with success for this payable
		h.updateDMWithSuccess(l, req.DMChannelID, req.DMMessageID, req.Month, payable, personalEmail)
		successCount++
	}

	// Step 7: Check if any payables were processed
	if successCount == 0 {
		l.Debug("no payables with PDF found")
		h.updateDMWithNotReady(l, req.DMChannelID, req.DMMessageID, req.Month)
		return
	}

	l.Info(fmt.Sprintf("payable lookup completed: discord_username=%s month=%s processed=%d/%d",
		req.DiscordUsername, req.Month, successCount, len(payables)))
}

// updateDMWithSuccess updates the Discord DM with success status
func (h *handler) updateDMWithSuccess(l logger.Logger, channelID, messageID, month string, payable *notion.PayableInfo, email string) {
	l.Debug(fmt.Sprintf("updating DM with success: channelID=%s messageID=%s", channelID, messageID))

	// Check if Discord service is available
	if h.service == nil || h.service.Discord == nil {
		l.Error(nil, "Discord service not available, cannot update DM")
		return
	}

	// Build success embed
	successEmbed := &discordgo.MessageEmbed{
		Title:       "✅ Invoice Ready",
		Description: "Your invoice has been found and shared with your personal email.",
		Color:       3066993, // Green
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "File",
				Value:  fmt.Sprintf("[View Invoice](%s)", payable.FileURL),
				Inline: false,
			},
			{
				Name:   "Email Notification",
				Value:  fmt.Sprintf("A notification has been sent to %s", email),
				Inline: false,
			},
			{
				Name:   "Status",
				Value:  payable.Status,
				Inline: true,
			},
			{
				Name:   "Amount",
				Value:  fmt.Sprintf("%.2f %s", payable.Total, payable.Currency),
				Inline: true,
			},
		},
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	_, err := h.service.Discord.UpdateChannelMessage(channelID, messageID, "", []*discordgo.MessageEmbed{successEmbed}, nil)
	if err != nil {
		l.Errorf(err, "failed to update discord DM with success")
	} else {
		l.Debug("successfully updated discord DM with success status")
	}
}

// updateDMWithSuccessNoShare updates the Discord DM with success status (file found, no sharing)
func (h *handler) updateDMWithSuccessNoShare(l logger.Logger, channelID, messageID, month string, payable *notion.PayableInfo) {
	l.Debug(fmt.Sprintf("updating DM with success (no share): channelID=%s messageID=%s", channelID, messageID))

	// Check if Discord service is available
	if h.service == nil || h.service.Discord == nil {
		l.Error(nil, "Discord service not available, cannot update DM")
		return
	}

	// Build success embed
	successEmbed := &discordgo.MessageEmbed{
		Title:       "✅ Invoice Ready",
		Description: "Your invoice has been found and is ready for download.",
		Color:       3066993, // Green
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "File",
				Value:  fmt.Sprintf("[Download Invoice](%s)", payable.FileURL),
				Inline: false,
			},
			{
				Name:   "Status",
				Value:  payable.Status,
				Inline: true,
			},
			{
				Name:   "Amount",
				Value:  fmt.Sprintf("%.2f %s", payable.Total, payable.Currency),
				Inline: true,
			},
		},
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	_, err := h.service.Discord.UpdateChannelMessage(channelID, messageID, "", []*discordgo.MessageEmbed{successEmbed}, nil)
	if err != nil {
		l.Errorf(err, "failed to update discord DM with success")
	} else {
		l.Debug("successfully updated discord DM with success status")
	}
}

// updateDMWithError updates the Discord DM with error status
func (h *handler) updateDMWithError(l logger.Logger, channelID, messageID, errorMsg string) {
	l.Debug(fmt.Sprintf("updating DM with error: channelID=%s messageID=%s error=%s", channelID, messageID, errorMsg))

	// Check if Discord service is available
	if h.service == nil || h.service.Discord == nil {
		l.Error(nil, "Discord service not available, cannot update DM")
		return
	}

	errorEmbed := &discordgo.MessageEmbed{
		Title:       "❌ Invoice Generation Failed",
		Description: errorMsg,
		Color:       15158332, // Red
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "What to do",
				Value:  "Please try again later or contact support if the issue persists.",
				Inline: false,
			},
		},
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	_, err := h.service.Discord.UpdateChannelMessage(channelID, messageID, "", []*discordgo.MessageEmbed{errorEmbed}, nil)
	if err != nil {
		l.Errorf(err, "failed to update discord DM with error")
	} else {
		l.Debug("successfully updated discord DM with error status")
	}
}

// updateDMWithRateLimitError updates the Discord DM with rate limit error
func (h *handler) updateDMWithRateLimitError(l logger.Logger, channelID, messageID, username string, remaining int, resetTime time.Time) {
	l.Debug(fmt.Sprintf("updating DM with rate limit error: channelID=%s messageID=%s", channelID, messageID))

	// Check if Discord service is available
	if h.service == nil || h.service.Discord == nil {
		l.Error(nil, "Discord service not available, cannot update DM")
		return
	}

	errorEmbed := &discordgo.MessageEmbed{
		Title:       "⏳ Rate Limit Exceeded",
		Description: "You have reached your daily invoice generation limit.",
		Color:       16776960, // Yellow
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Remaining Today",
				Value:  fmt.Sprintf("%d/%d", remaining, ratelimit.MaxInvoiceGenerationsPerDay),
				Inline: true,
			},
			{
				Name:   "Resets At",
				Value:  resetTime.Format("15:04 UTC"),
				Inline: true,
			},
			{
				Name:   "Note",
				Value:  "The limit resets at midnight UTC each day.",
				Inline: false,
			},
		},
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	_, err := h.service.Discord.UpdateChannelMessage(channelID, messageID, "", []*discordgo.MessageEmbed{errorEmbed}, nil)
	if err != nil {
		l.Errorf(err, "failed to update discord DM with rate limit error")
	} else {
		l.Debug("successfully updated discord DM with rate limit error")
	}
}

// updateDMWithPartialSuccess updates the Discord DM with partial success (file found but not shared)
func (h *handler) updateDMWithPartialSuccess(l logger.Logger, channelID, messageID string, payable *notion.PayableInfo, email string, shareErr error) {
	l.Debug(fmt.Sprintf("updating DM with partial success: channelID=%s messageID=%s", channelID, messageID))

	// Check if Discord service is available
	if h.service == nil || h.service.Discord == nil {
		l.Error(nil, "Discord service not available, cannot update DM")
		return
	}

	// Truncate file URL for display (Discord embed field limit is 1024 chars)
	fileURL := payable.FileURL
	displayURL := truncateString(fileURL, 200) // Keep URL short for display

	// Format issue message with truncated error
	issueMsg := fmt.Sprintf("Failed to share with %s", email)
	if shareErr != nil {
		errStr := shareErr.Error()
		// Truncate error message to keep embed field under 1024 chars
		if len(errStr) > 100 {
			errStr = errStr[:100] + "..."
		}
		issueMsg = fmt.Sprintf("%s: %s", issueMsg, errStr)
	}

	warningEmbed := &discordgo.MessageEmbed{
		Title:       "⚠️ Invoice Found (Partial Success)",
		Description: "Your invoice has been found but we couldn't send the email notification.",
		Color:       16776960, // Yellow
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "File",
				Value:  fmt.Sprintf("[View Invoice](%s)", displayURL),
				Inline: false,
			},
			{
				Name:   "Issue",
				Value:  truncateString(issueMsg, 500),
				Inline: false,
			},
			{
				Name:   "What to do",
				Value:  "You can still access the file using the link above. Contact support if you need help.",
				Inline: false,
			},
		},
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	_, err := h.service.Discord.UpdateChannelMessage(channelID, messageID, "", []*discordgo.MessageEmbed{warningEmbed}, nil)
	if err != nil {
		l.Errorf(err, "failed to update discord DM with partial success")
	} else {
		l.Debug("successfully updated discord DM with partial success status")
	}
}

// updateDMWithNotReady updates the Discord DM with "not ready" message
func (h *handler) updateDMWithNotReady(l logger.Logger, channelID, messageID, month string) {
	l.Debug(fmt.Sprintf("updating DM with not ready: channelID=%s messageID=%s", channelID, messageID))

	// Check if Discord service is available
	if h.service == nil || h.service.Discord == nil {
		l.Error(nil, "Discord service not available, cannot update DM")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "⏳ Invoice Being Processed",
		Description: "Your invoice is currently being processed and is not yet ready for payout; we will notify you immediately once it is cleared.",
		Color:       16776960, // Yellow
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Month", Value: month, Inline: true},
			{Name: "Note", Value: "This process typically takes 1-2 business days", Inline: false},
		},
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	_, err := h.service.Discord.UpdateChannelMessage(channelID, messageID, "", []*discordgo.MessageEmbed{embed}, nil)
	if err != nil {
		l.Errorf(err, "failed to update Discord DM with not ready")
	} else {
		l.Debug("successfully updated Discord DM with not ready status")
	}
}

// isValidMonthFormat validates that the month is in YYYY-MM format
func isValidMonthFormat(month string) bool {
	matched, _ := regexp.MatchString(`^\d{4}-\d{2}$`, month)
	return matched
}

// truncateString truncates a string to maxLen characters, adding "..." if truncated
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
