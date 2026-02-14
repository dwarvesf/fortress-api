package webhook

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	nt "github.com/dstotijn/go-notion"
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/controller/invoice"
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
	InvoiceType     string `json:"invoice_type"` // "service_and_refund" | "extra_payment"
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

	l.Debug(fmt.Sprintf("parsed request: discord_username=%s month=%s dm_channel_id=%s dm_message_id=%s invoice_type=%s",
		req.DiscordUsername, req.Month, req.DMChannelID, req.DMMessageID, req.InvoiceType))

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

// processGenInvoice handles the async payable lookup and generation process
// Flow:
// 1. Query contractor rates
// 2. Calculate period dates
// 3. Lookup ALL payables (any status) for contractor + period
// 4. Branch based on payable status:
//   - No payable exists → generate invoice, create payable, share PDF
//   - Status = "New" → regenerate invoice, update payable, share PDF
//   - Status = "Pending" → find existing PDF and share (existing behavior)
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
	// Period start is payday of the invoice month (work month)
	// Period end is payday of the next month
	// e.g., for month=2026-01 with payDay=15: periodStart=2026-01-15, periodEnd=2026-02-15
	monthTime, _ := time.Parse("2006-01", req.Month)
	payDay := rateData.PayDay
	if payDay == 0 {
		payDay = 15 // Default to 15
	}
	periodStart := time.Date(monthTime.Year(), monthTime.Month(), payDay, 0, 0, 0, 0, time.UTC)

	l.Debug(fmt.Sprintf("calculated period start: %s", periodStart.Format("2006-01-02")))

	// Step 3: Lookup ALL existing payables (any status)
	l.Debug("step 3: looking up all payables (any status)")
	allPayables, err := h.service.Notion.ContractorPayables.FindPayableByContractorAndPeriodAllStatus(ctx, rateData.ContractorPageID, periodStart.Format("2006-01-02"))
	if err != nil {
		l.Errorf(err, "failed to lookup payables")
		h.updateDMWithError(l, req.DMChannelID, req.DMMessageID, "Failed to lookup invoice. Please try again later.")
		return
	}

	l.Debug(fmt.Sprintf("found %d payable(s) of any status", len(allPayables)))

	// Step 4: Branch based on payable status
	// Priority: Pending > New > None
	var pendingPayables []*notion.PayableInfo
	var newPayable *notion.PayableInfo

	for _, p := range allPayables {
		l.Debug(fmt.Sprintf("payable: pageID=%s status=%s invoiceID=%s", p.PageID, p.Status, p.InvoiceID))
		switch p.Status {
		case "Pending":
			pendingPayables = append(pendingPayables, p)
		case "New":
			if newPayable == nil {
				newPayable = p // Keep first "New" payable for regeneration
			}
		}
	}

	// Case 1: Pending payables exist
	if len(pendingPayables) > 0 {
		l.Debug(fmt.Sprintf("found %d pending payable(s)", len(pendingPayables)))

		// If no specific invoice type requested, use existing behavior (return first pending payable)
		if req.InvoiceType == "" {
			l.Debug("no invoice type specified - returning existing pending payable(s)")
			h.processExistingPendingPayables(l, req, rateData, pendingPayables)
			return
		}

		// If invoice type is specified, check if pending payable matches the requested type
		l.Debug(fmt.Sprintf("invoice type specified: %s - checking if pending payable matches", req.InvoiceType))

		// Check the first pending payable (there should typically be only one per period)
		firstPendingPayable := pendingPayables[0]
		existingPayableType := h.determinePayableInvoiceType(ctx, l, firstPendingPayable, rateData.ContractorPageID)

		l.Debug(fmt.Sprintf("existing payable type: %s, requested type: %s", existingPayableType, req.InvoiceType))

		if existingPayableType == req.InvoiceType {
			l.Debug("payable type matches - returning existing pending payable(s)")
			h.processExistingPendingPayables(l, req, rateData, pendingPayables)
			return
		}

		// Type doesn't match - fall through to generate new invoice
		l.Debug("payable type doesn't match - will generate new invoice with requested type")
	}

	// Case 2: New payable exists → regenerate invoice and update payable
	if newPayable != nil {
		l.Debug(fmt.Sprintf("found payable with status=New (pageID=%s) - regenerating invoice", newPayable.PageID))
		h.generateAndCreatePayable(ctx, l, req, rateData, newPayable.PageID)
		return
	}

	// Case 3: No payable exists → generate new invoice and create payable
	l.Debug("no payable exists - generating new invoice")
	h.generateAndCreatePayable(ctx, l, req, rateData, "")
}

// processExistingPendingPayables handles the existing flow for pending payables
// This is the original behavior: find existing PDF and share with contractor
func (h *handler) processExistingPendingPayables(l logger.Logger, req GenInvoiceRequest, rateData *notion.ContractorRateData, payables []*notion.PayableInfo) {
	ctx := context.Background()

	// Get contractor personal email
	l.Debug("fetching contractor personal email")
	taskOrderLogService := h.service.Notion.TaskOrderLog
	personalEmail := taskOrderLogService.GetContractorPersonalEmail(ctx, rateData.ContractorPageID)
	if personalEmail == "" {
		l.Debug(fmt.Sprintf("personal email not found for contractor: %s", rateData.ContractorPageID))
		h.updateDMWithError(l, req.DMChannelID, req.DMMessageID, "Personal email not found. Please contact HR to update your profile.")
		return
	}

	l.Debug(fmt.Sprintf("found personal email: %s", personalEmail))

	// Process each pending payable separately
	l.Debug(fmt.Sprintf("processing %d pending payable(s)", len(payables)))
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

	// Check if any payables were processed
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

	// Build success embed - use Description for the link (4096 char limit vs 1024 for fields)
	// This handles long Notion URLs that exceed Discord's field limit
	successEmbed := &discordgo.MessageEmbed{
		Title:       "✅ Invoice Ready",
		Description: fmt.Sprintf("Your invoice has been found and is ready for download.\n\n[📄 Download Invoice](%s)", payable.FileURL),
		Color:       3066993, // Green
		Fields: []*discordgo.MessageEmbedField{
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

// extractFileIDFromURL extracts the file ID from a Google Drive URL
// URL format: https://drive.google.com/file/d/{FILE_ID}/view
// Returns empty string if URL format doesn't match
func extractFileIDFromURL(url string) string {
	// Pattern: https://drive.google.com/file/d/{FILE_ID}/view
	prefix := "https://drive.google.com/file/d/"
	suffix := "/view"

	if len(url) <= len(prefix)+len(suffix) {
		return ""
	}

	// Check if URL starts with prefix
	if url[:len(prefix)] != prefix {
		return ""
	}

	// Remove prefix
	remainder := url[len(prefix):]

	// Find /view suffix and extract file ID
	viewIdx := len(remainder) - len(suffix)
	if viewIdx <= 0 {
		return ""
	}

	if remainder[viewIdx:] != suffix {
		return ""
	}

	return remainder[:viewIdx]
}

// updateDMWithGenerating updates the Discord DM with "generating invoice" progress message
func (h *handler) updateDMWithGenerating(l logger.Logger, channelID, messageID, month string) {
	l.Debug(fmt.Sprintf("updating DM with generating status: channelID=%s messageID=%s", channelID, messageID))

	if h.service == nil || h.service.Discord == nil {
		l.Error(nil, "Discord service not available, cannot update DM")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "⏳ Generating Invoice",
		Description: "Your invoice is being generated. This may take a moment...",
		Color:       3447003, // Blue
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Month", Value: month, Inline: true},
		},
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	_, err := h.service.Discord.UpdateChannelMessage(channelID, messageID, "", []*discordgo.MessageEmbed{embed}, nil)
	if err != nil {
		l.Errorf(err, "failed to update Discord DM with generating status")
	} else {
		l.Debug("successfully updated Discord DM with generating status")
	}
}

// updateDMWithInvoiceGenerated updates the Discord DM with success status for newly generated invoice
func (h *handler) updateDMWithInvoiceGenerated(l logger.Logger, channelID, messageID, month string, invoiceNumber string, total float64, currency string, fileURL string, email string) {
	l.Debug(fmt.Sprintf("updating DM with invoice generated: channelID=%s messageID=%s invoiceNumber=%s", channelID, messageID, invoiceNumber))

	if h.service == nil || h.service.Discord == nil {
		l.Error(nil, "Discord service not available, cannot update DM")
		return
	}

	successEmbed := &discordgo.MessageEmbed{
		Title:       "✅ Invoice Generated",
		Description: "Your invoice has been generated and is ready for review.",
		Color:       3066993, // Green
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Invoice Number",
				Value:  invoiceNumber,
				Inline: true,
			},
			{
				Name:   "Amount",
				Value:  fmt.Sprintf("%.2f %s", total, currency),
				Inline: true,
			},
			{
				Name:   "File",
				Value:  fmt.Sprintf("[View Invoice](%s)", fileURL),
				Inline: false,
			},
			{
				Name:   "Email Notification",
				Value:  fmt.Sprintf("A notification has been sent to %s", email),
				Inline: false,
			},
		},
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	_, err := h.service.Discord.UpdateChannelMessage(channelID, messageID, "", []*discordgo.MessageEmbed{successEmbed}, nil)
	if err != nil {
		l.Errorf(err, "failed to update Discord DM with invoice generated")
	} else {
		l.Debug("successfully updated Discord DM with invoice generated status")
	}
}

// updateDMWithNoPayouts updates the Discord DM with "no pending payouts" message
func (h *handler) updateDMWithNoPayouts(l logger.Logger, channelID, messageID, month string) {
	l.Debug(fmt.Sprintf("updating DM with no payouts: channelID=%s messageID=%s", channelID, messageID))

	if h.service == nil || h.service.Discord == nil {
		l.Error(nil, "Discord service not available, cannot update DM")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "ℹ️ No Pending Payouts",
		Description: fmt.Sprintf("No pending payouts found for %s.", month),
		Color:       3447003, // Blue
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Note",
				Value:  "Payouts are typically available after payday. If you believe this is an error, please contact HR.",
				Inline: false,
			},
		},
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000-07:00"),
	}

	_, err := h.service.Discord.UpdateChannelMessage(channelID, messageID, "", []*discordgo.MessageEmbed{embed}, nil)
	if err != nil {
		l.Errorf(err, "failed to update Discord DM with no payouts")
	} else {
		l.Debug("successfully updated Discord DM with no payouts status")
	}
}

// determinePayableInvoiceType inspects a payable's linked payout items to determine its invoice type
// Returns "service_and_refund" or "extra_payment" based on the payout source types
func (h *handler) determinePayableInvoiceType(ctx context.Context, l logger.Logger, payable *notion.PayableInfo, contractorPageID string) string {
	l.Debug(fmt.Sprintf("determining invoice type for payable: pageID=%s", payable.PageID))

	// Step 1: Get payable page properties to extract Payout Items relation
	l.Debug("step 1: fetching payable page properties")
	page, err := h.service.Notion.GetPage(payable.PageID)
	if err != nil {
		l.Errorf(err, "failed to fetch payable page")
		return "service_and_refund" // Default on error
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		l.Debug("failed to cast page properties")
		return "service_and_refund" // Default on error
	}

	// Extract Payout Items relation IDs (using the same pattern as contractor_payables.go)
	payoutItemIDs := extractAllRelationIDs(props, "Payout Items")
	l.Debug(fmt.Sprintf("found %d payout items linked to payable", len(payoutItemIDs)))

	if len(payoutItemIDs) == 0 {
		l.Debug("no payout items found, defaulting to service_and_refund")
		return "service_and_refund"
	}

	// Step 2: Query all payout entries for this contractor
	// We don't have a specific cutoff date, so pass empty string to get all pending payouts
	l.Debug("step 2: querying contractor payouts")
	allPayouts, err := h.service.Notion.ContractorPayouts.QueryPendingPayoutsByContractor(ctx, contractorPageID, "")
	if err != nil {
		l.Errorf(err, "failed to query contractor payouts")
		return "service_and_refund" // Default on error
	}

	l.Debug(fmt.Sprintf("fetched %d total pending payouts for contractor", len(allPayouts)))

	// Step 3: Filter to only payouts in the payable's Payout Items list
	payoutItemIDSet := make(map[string]bool)
	for _, id := range payoutItemIDs {
		payoutItemIDSet[id] = true
	}

	var relevantPayouts []notion.PayoutEntry
	for _, payout := range allPayouts {
		if payoutItemIDSet[payout.PageID] {
			relevantPayouts = append(relevantPayouts, payout)
		}
	}

	l.Debug(fmt.Sprintf("filtered to %d relevant payouts", len(relevantPayouts)))

	if len(relevantPayouts) == 0 {
		l.Debug("no relevant payouts found, defaulting to service_and_refund")
		return "service_and_refund"
	}

	// Step 4: Count source types to determine invoice type
	serviceAndRefundCount := 0
	extraPaymentCount := 0

	for _, payout := range relevantPayouts {
		l.Debug(fmt.Sprintf("payout: pageID=%s sourceType=%s taskOrderID=%s invoiceSplitID=%s refundRequestID=%s description=%s",
			payout.PageID, payout.SourceType, payout.TaskOrderID, payout.InvoiceSplitID, payout.RefundRequestID, payout.Description))

		// Determine source type based on relations
		// Priority: TaskOrderID > InvoiceSplitID > RefundRequestID > None (ExtraPayment)
		var sourceType notion.PayoutSourceType

		if payout.TaskOrderID != "" {
			sourceType = notion.PayoutSourceTypeServiceFee
		} else if payout.InvoiceSplitID != "" {
			// Check if it's AM/DL fee (goes to extra_payment) or Commission
			descLower := strings.ToLower(payout.Description)
			if strings.Contains(descLower, "delivery lead") || strings.Contains(descLower, "account management") {
				// Fee section (AM/DL) → extra_payment
				sourceType = notion.PayoutSourceTypeServiceFee // Still ServiceFee type but goes to extra_payment
				extraPaymentCount++
				continue // Don't count in service_and_refund
			} else {
				// Commission → extra_payment
				sourceType = notion.PayoutSourceTypeCommission
			}
		} else if payout.RefundRequestID != "" {
			sourceType = notion.PayoutSourceTypeRefund
		} else {
			// No relations set → ExtraPayment
			sourceType = notion.PayoutSourceTypeExtraPayment
		}

		// Map source type to invoice type
		switch sourceType {
		case notion.PayoutSourceTypeServiceFee, notion.PayoutSourceTypeRefund:
			serviceAndRefundCount++
		case notion.PayoutSourceTypeCommission, notion.PayoutSourceTypeExtraPayment:
			extraPaymentCount++
		}
	}

	l.Debug(fmt.Sprintf("source type counts: serviceAndRefund=%d extraPayment=%d", serviceAndRefundCount, extraPaymentCount))

	// Step 5: Return invoice type based on majority
	// If any extra payment types exist, consider it an extra payment invoice
	if extraPaymentCount > 0 {
		l.Debug("determined invoice type: extra_payment")
		return "extra_payment"
	}

	l.Debug("determined invoice type: service_and_refund")
	return "service_and_refund"
}

// extractAllRelationIDs extracts all relation IDs from a database page property
func extractAllRelationIDs(props nt.DatabasePageProperties, propName string) []string {
	prop, ok := props[propName]
	if !ok || len(prop.Relation) == 0 {
		return nil
	}
	ids := make([]string, len(prop.Relation))
	for i, rel := range prop.Relation {
		ids[i] = rel.ID
	}
	return ids
}

// generateAndCreatePayable generates a contractor invoice, creates/updates payable in Notion, and shares with contractor
// existingPayablePageID is empty for new payables, or the page ID to update for regeneration
func (h *handler) generateAndCreatePayable(
	ctx context.Context,
	l logger.Logger,
	req GenInvoiceRequest,
	rateData *notion.ContractorRateData,
	existingPayablePageID string,
) {
	// Update DM with generating status
	h.updateDMWithGenerating(l, req.DMChannelID, req.DMMessageID, req.Month)

	// Default invoice type to service_and_refund
	invoiceType := req.InvoiceType
	if invoiceType == "" {
		invoiceType = "service_and_refund"
	}

	// Step 1: Generate invoice data
	l.Debug(fmt.Sprintf("step 1: generating contractor invoice data with invoiceType=%s", invoiceType))
	opts := &invoice.ContractorInvoiceOptions{
		GroupFeeByProject: false,
		InvoiceType:       invoiceType,
	}
	invoiceData, err := h.controller.Invoice.GenerateContractorInvoice(ctx, req.DiscordUsername, req.Month, opts)
	if err != nil {
		l.Errorf(err, "failed to generate contractor invoice")
		// Check if it's a "no pending payouts" error
		if err.Error() == fmt.Sprintf("no pending payouts found for contractor=%s", req.DiscordUsername) {
			h.updateDMWithNoPayouts(l, req.DMChannelID, req.DMMessageID, req.Month)
		} else {
			h.updateDMWithError(l, req.DMChannelID, req.DMMessageID, "Failed to generate invoice. Please try again later.")
		}
		return
	}

	l.Debug(fmt.Sprintf("invoice data generated: invoiceNumber=%s total=%.2f currency=%s lineItems=%d",
		invoiceData.InvoiceNumber, invoiceData.TotalUSD, invoiceData.Currency, len(invoiceData.LineItems)))

	// Step 2: Generate PDF
	l.Debug("step 2: generating invoice PDF")
	pdfBytes, err := h.controller.Invoice.GenerateContractorInvoicePDF(l, invoiceData)
	if err != nil {
		l.Errorf(err, "failed to generate invoice PDF")
		// Check if it's a "no payouts found" error from section filtering
		if strings.Contains(err.Error(), "payouts found for this month") {
			h.updateDMWithError(l, req.DMChannelID, req.DMMessageID, fmt.Sprintf("No %s payouts found for %s.", invoiceType, req.Month))
		} else {
			h.updateDMWithError(l, req.DMChannelID, req.DMMessageID, "Failed to generate PDF. Please try again later.")
		}
		return
	}

	l.Debug(fmt.Sprintf("PDF generated: size=%d bytes", len(pdfBytes)))

	// Step 3: Get contractor personal email for sharing
	l.Debug("step 3: fetching contractor personal email")
	taskOrderLogService := h.service.Notion.TaskOrderLog
	personalEmail := taskOrderLogService.GetContractorPersonalEmail(ctx, rateData.ContractorPageID)
	if personalEmail == "" {
		l.Debug(fmt.Sprintf("personal email not found for contractor: %s, using team email", rateData.ContractorPageID))
		personalEmail = rateData.TeamEmail // Fallback to team email
	}

	l.Debug(fmt.Sprintf("using email for sharing: %s", personalEmail))

	// Step 4: Upload PDF to Google Drive
	l.Debug("step 4: uploading PDF to Google Drive")
	fileName := invoiceData.InvoiceNumber + ".pdf"
	// UploadContractorInvoicePDF returns a full Google Drive URL, not just the file ID
	uploadedURL, err := h.service.GoogleDrive.UploadContractorInvoicePDF(invoiceData.ContractorName, fileName, pdfBytes)
	if err != nil {
		l.Errorf(err, "failed to upload PDF to Google Drive")
		// Non-fatal: continue to create payable without Google Drive file
		// The PDF is still in Notion attachment
	} else {
		l.Debug(fmt.Sprintf("PDF uploaded to Google Drive: url=%s", uploadedURL))
	}

	// Step 5: Calculate period dates for payable
	// Period start: payday of the invoice month (work month)
	// Period end: payday of the next month
	// e.g., for month=2026-01 with payDay=15: periodStart=2026-01-15, periodEnd=2026-02-15
	payDay := invoiceData.PayDay
	if payDay == 0 {
		payDay = 15 // Default to 15
	}
	monthTime, _ := time.Parse("2006-01", req.Month)
	nextMonth := monthTime.AddDate(0, 1, 0)

	// Period start: payday of invoice month (work month)
	periodStart := time.Date(monthTime.Year(), monthTime.Month(), payDay, 0, 0, 0, 0, time.UTC)
	// Period end: payday of the next month
	periodEnd := time.Date(nextMonth.Year(), nextMonth.Month(), payDay, 0, 0, 0, 0, time.UTC)

	l.Debug(fmt.Sprintf("calculated period: start=%s end=%s", periodStart.Format("2006-01-02"), periodEnd.Format("2006-01-02")))

	// Step 6: Create or update payable in Notion
	l.Debug("step 6: creating/updating payable in Notion")
	payableInput := notion.CreatePayableInput{
		ContractorPageID: rateData.ContractorPageID,
		Total:            invoiceData.TotalUSD,
		Currency:         "USD",
		PeriodStart:      periodStart.Format("2006-01-02"),
		PeriodEnd:        periodEnd.Format("2006-01-02"),
		InvoiceDate:      invoiceData.Date.Format("2006-01-02"),
		InvoiceID:        invoiceData.InvoiceNumber,
		PayoutItemIDs:    invoiceData.PayoutPageIDs,
		ContractorType:   "Individual", // Default
		ExchangeRate:     invoiceData.ExchangeRate,
		PDFBytes:         pdfBytes,
	}

	payablePageID, err := h.service.Notion.ContractorPayables.CreatePayable(ctx, payableInput)
	if err != nil {
		l.Errorf(err, "failed to create/update payable in Notion")
		// Non-fatal: PDF is still available, just couldn't create payable record
		// Continue with sharing
	} else {
		l.Debug(fmt.Sprintf("payable created/updated: pageID=%s", payablePageID))
	}

	// Step 7: Share file with contractor's personal email (if Google Drive upload succeeded)
	var googleDriveURL string
	if uploadedURL != "" {
		googleDriveURL = uploadedURL // Already a full URL from UploadContractorInvoicePDF

		// Extract file ID from URL for sharing
		// URL format: https://drive.google.com/file/d/{FILE_ID}/view
		fileID := extractFileIDFromURL(uploadedURL)
		if fileID != "" {
			l.Debug("step 7: sharing file with contractor email")
			if err := h.service.GoogleDrive.ShareFileWithEmail(fileID, personalEmail); err != nil {
				l.Errorf(err, fmt.Sprintf("failed to share file with email: %s", personalEmail))
				// Non-fatal: file exists, just couldn't share
			} else {
				l.Debug(fmt.Sprintf("file shared with email: %s", personalEmail))
			}
		} else {
			l.Debug(fmt.Sprintf("could not extract file ID from URL: %s", uploadedURL))
		}
	}

	// Step 8: Update DM with success
	l.Debug("step 8: updating DM with success")
	if googleDriveURL != "" {
		h.updateDMWithInvoiceGenerated(l, req.DMChannelID, req.DMMessageID, req.Month,
			invoiceData.InvoiceNumber, invoiceData.TotalUSD, invoiceData.Currency, googleDriveURL, personalEmail)
	} else {
		// Fallback: no Google Drive URL, show partial success
		h.updateDMWithError(l, req.DMChannelID, req.DMMessageID,
			fmt.Sprintf("Invoice generated (ID: %s, Amount: %.2f %s) but failed to upload to Google Drive. Please try again later.",
				invoiceData.InvoiceNumber, invoiceData.TotalUSD, invoiceData.Currency))
	}

	l.Info(fmt.Sprintf("invoice generation completed: discord_username=%s month=%s invoiceNumber=%s total=%.2f",
		req.DiscordUsername, req.Month, invoiceData.InvoiceNumber, invoiceData.TotalUSD))
}
