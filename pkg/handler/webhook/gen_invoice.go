package webhook

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
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

// processGenInvoice handles the async invoice generation process
func (h *handler) processGenInvoice(l logger.Logger, req GenInvoiceRequest) {
	ctx := context.Background()

	l.Debug(fmt.Sprintf("starting async invoice generation: discord_username=%s month=%s", req.DiscordUsername, req.Month))

	// Step 1: Generate invoice data
	l.Debug("step 1: generating invoice data")
	invoiceData, err := h.controller.Invoice.GenerateContractorInvoice(ctx, req.DiscordUsername, req.Month)
	if err != nil {
		l.Errorf(err, "failed to generate contractor invoice data")
		h.updateDMWithError(l, req.DMChannelID, req.DMMessageID, fmt.Sprintf("Failed to generate invoice: %v", err))
		return
	}

	l.Debug(fmt.Sprintf("invoice data generated: contractor=%s invoiceNumber=%s total=%.2f",
		invoiceData.ContractorName, invoiceData.InvoiceNumber, invoiceData.TotalUSD))

	// Step 2: Generate PDF
	l.Debug("step 2: generating invoice PDF")
	pdfBytes, err := h.controller.Invoice.GenerateContractorInvoicePDF(l, invoiceData)
	if err != nil {
		l.Errorf(err, "failed to generate contractor invoice PDF")
		h.updateDMWithError(l, req.DMChannelID, req.DMMessageID, fmt.Sprintf("Failed to generate PDF: %v", err))
		return
	}

	l.Debug(fmt.Sprintf("PDF generated: size=%d bytes", len(pdfBytes)))

	// Step 3: Upload PDF to Google Drive
	l.Debug("step 3: uploading PDF to Google Drive")
	fileName := fmt.Sprintf("%s.pdf", invoiceData.InvoiceNumber)
	fileURL, err := h.service.GoogleDrive.UploadContractorInvoicePDF(invoiceData.ContractorName, fileName, pdfBytes)
	if err != nil {
		l.Errorf(err, "failed to upload PDF to Google Drive")
		h.updateDMWithError(l, req.DMChannelID, req.DMMessageID, fmt.Sprintf("Failed to upload PDF: %v", err))
		return
	}

	l.Debug(fmt.Sprintf("PDF uploaded to Google Drive: url=%s", fileURL))

	// Step 3.5: Create Contractor Payables record in Notion
	l.Debug("step 3.5: creating contractor payables record in Notion")

	payableInput := notion.CreatePayableInput{
		ContractorPageID: invoiceData.ContractorPageID,
		Total:            invoiceData.TotalUSD,
		Currency:         "USD",
		Period:           invoiceData.Month + "-01",
		InvoiceDate:      time.Now().Format("2006-01-02"),
		InvoiceID:        invoiceData.InvoiceNumber,
		PayoutItemIDs:    invoiceData.PayoutPageIDs,
		ContractorType:   "Individual", // Default to Individual
		PDFBytes:         pdfBytes,     // Upload PDF to Notion
	}

	l.Debug(fmt.Sprintf("[DEBUG] payable input: contractor=%s total=%.2f payoutItems=%d",
		payableInput.ContractorPageID, payableInput.Total, len(payableInput.PayoutItemIDs)))

	payablePageID, payableErr := h.service.Notion.ContractorPayables.CreatePayable(ctx, payableInput)
	if payableErr != nil {
		l.Errorf(payableErr, "[DEBUG] failed to create contractor payables record - continuing with response")
		// Non-fatal: continue with response
	} else {
		l.Debug(fmt.Sprintf("[DEBUG] contractor payables record created: pageID=%s", payablePageID))
	}

	// Extract file ID from URL for sharing
	fileID := extractFileIDFromURL(fileURL)
	if fileID == "" {
		l.Error(nil, "failed to extract file ID from URL")
		h.updateDMWithError(l, req.DMChannelID, req.DMMessageID, "Failed to process uploaded file")
		return
	}

	l.Debug(fmt.Sprintf("extracted file ID: %s", fileID))

	// Step 4: Get contractor's personal email from Notion
	l.Debug("step 4: fetching contractor personal email")
	ratesService := notion.NewContractorRatesService(h.config, h.logger)
	if ratesService == nil {
		l.Error(nil, "failed to create contractor rates service")
		h.updateDMWithError(l, req.DMChannelID, req.DMMessageID, "Failed to lookup contractor information")
		return
	}

	rateData, err := ratesService.QueryRatesByDiscordAndMonth(ctx, req.DiscordUsername, req.Month)
	if err != nil {
		l.Errorf(err, "failed to query contractor rates")
		h.updateDMWithError(l, req.DMChannelID, req.DMMessageID, fmt.Sprintf("Failed to find contractor: %v", err))
		return
	}

	taskOrderLogService := h.service.Notion.TaskOrderLog
	personalEmail := taskOrderLogService.GetContractorPersonalEmail(ctx, rateData.ContractorPageID)
	if personalEmail == "" {
		l.Debug(fmt.Sprintf("personal email not found for contractor: %s", rateData.ContractorPageID))
		h.updateDMWithError(l, req.DMChannelID, req.DMMessageID, "Personal email not found. Please contact HR to update your profile.")
		return
	}

	l.Debug(fmt.Sprintf("found personal email: %s", personalEmail))

	// Step 5: Share file with contractor's personal email
	l.Debug("step 5: sharing file with contractor email")
	if err := h.service.GoogleDrive.ShareFileWithEmail(fileID, personalEmail); err != nil {
		l.Errorf(err, "failed to share file with email")
		// Non-fatal: file is uploaded, just couldn't share. Update with partial success
		h.updateDMWithPartialSuccess(l, req.DMChannelID, req.DMMessageID, invoiceData, fileURL, personalEmail, err)
		return
	}

	l.Debug(fmt.Sprintf("file shared with email: %s", personalEmail))

	// Step 6: Update DM with success
	l.Debug("step 6: updating Discord DM with success")
	h.updateDMWithSuccess(l, req.DMChannelID, req.DMMessageID, invoiceData, fileURL, personalEmail)

	l.Info(fmt.Sprintf("invoice generation completed successfully: discord_username=%s month=%s invoice=%s",
		req.DiscordUsername, req.Month, invoiceData.InvoiceNumber))
}

// updateDMWithSuccess updates the Discord DM with success status
func (h *handler) updateDMWithSuccess(l logger.Logger, channelID, messageID string, invoiceData interface{}, fileURL, email string) {
	l.Debug(fmt.Sprintf("updating DM with success: channelID=%s messageID=%s", channelID, messageID))

	// Check if Discord service is available
	if h.service == nil || h.service.Discord == nil {
		l.Error(nil, "Discord service not available, cannot update DM")
		return
	}

	// Build success embed
	successEmbed := &discordgo.MessageEmbed{
		Title:       "✅ Invoice Generated Successfully",
		Description: "Your invoice has been generated and shared with your personal email.",
		Color:       3066993, // Green
		Fields: []*discordgo.MessageEmbedField{
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

// updateDMWithPartialSuccess updates the Discord DM with partial success (file uploaded but not shared)
func (h *handler) updateDMWithPartialSuccess(l logger.Logger, channelID, messageID string, invoiceData interface{}, fileURL, email string, shareErr error) {
	l.Debug(fmt.Sprintf("updating DM with partial success: channelID=%s messageID=%s", channelID, messageID))

	// Check if Discord service is available
	if h.service == nil || h.service.Discord == nil {
		l.Error(nil, "Discord service not available, cannot update DM")
		return
	}

	warningEmbed := &discordgo.MessageEmbed{
		Title:       "⚠️ Invoice Generated (Partial Success)",
		Description: "Your invoice has been generated but we couldn't send the email notification.",
		Color:       16776960, // Yellow
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "File",
				Value:  fmt.Sprintf("[View Invoice](%s)", fileURL),
				Inline: false,
			},
			{
				Name:   "Issue",
				Value:  fmt.Sprintf("Failed to share with %s: %v", email, shareErr),
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

// isValidMonthFormat validates that the month is in YYYY-MM format
func isValidMonthFormat(month string) bool {
	matched, _ := regexp.MatchString(`^\d{4}-\d{2}$`, month)
	return matched
}

// extractFileIDFromURL extracts the file ID from a Google Drive URL
// URL format: https://drive.google.com/file/d/{fileID}/view
func extractFileIDFromURL(url string) string {
	// Remove the base URL and /view suffix
	s := strings.Replace(url, "https://drive.google.com/file/d/", "", 1)
	s = strings.Replace(s, "/view", "", 1)
	return s
}
