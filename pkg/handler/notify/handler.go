package notify

import (
	"fmt"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bwmarrin/discordgo"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	discordsvc "github.com/dwarvesf/fortress-api/pkg/service/discord"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	service *service.Service
	logger  logger.Logger
	config  *config.Config
}

// New creates a new notify handler
func New(service *service.Service, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		service: service,
		logger:  logger,
		config:  cfg,
	}
}

// ExtraPaymentContractor represents a contractor to be notified
type ExtraPaymentContractor struct {
	PageID  string  `json:"pageId"`
	Name    string  `json:"name"`
	Discord string  `json:"discord"`
	Email   string  `json:"email"`
	Amount  float64 `json:"amount"`
	Reason  string  `json:"reason"`
}

// PreviewExtraPaymentNotificationResponse represents the preview response
type PreviewExtraPaymentNotificationResponse struct {
	Count       int                      `json:"count"`
	TotalAmount float64                  `json:"totalAmount"`
	Contractors []ExtraPaymentContractor `json:"contractors"`
}

// SendExtraPaymentNotificationRequest represents the request body for sending notifications
type SendExtraPaymentNotificationRequest struct {
	Reasons []string `json:"reasons,omitempty"` // Optional custom reasons (override Notion descriptions)
}

// SendExtraPaymentNotificationResponse represents the send response
type SendExtraPaymentNotificationResponse struct {
	Sent   int      `json:"sent"`
	Failed int      `json:"failed"`
	Errors []string `json:"errors,omitempty"`
}

// SendOneExtraPaymentNotificationResponse represents the single send response
type SendOneExtraPaymentNotificationResponse struct {
	Success bool   `json:"success"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Error   string `json:"error,omitempty"`
}

// PreviewExtraPaymentNotification godoc
// @Summary Preview contractors to notify for extra payments
// @Description Returns a list of contractors with pending Commission or Extra Payment payouts for the specified month
// @Tags Notify
// @Accept json
// @Produce json
// @Param month query string true "Month in YYYY-MM format"
// @Param discord query string false "Optional Discord username filter"
// @Success 200 {object} view.Response{data=PreviewExtraPaymentNotificationResponse}
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /api/v1/notify/extra-payment/preview [post]
func (h *handler) PreviewExtraPaymentNotification(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "notify",
		"method":  "PreviewExtraPaymentNotification",
	})

	// Parse query params
	month := c.Query("month")
	if month == "" {
		l.Debug("month parameter is required")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, fmt.Errorf("month parameter is required (format: YYYY-MM)"), nil, ""))
		return
	}

	// Validate month format
	if _, err := time.Parse("2006-01", month); err != nil {
		l.Debug("invalid month format")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, fmt.Errorf("invalid month format (expected: YYYY-MM)"), nil, ""))
		return
	}

	discordUsername := c.Query("discord")

	l.Debugf("previewing extra payment notifications for month=%s discord=%s", month, discordUsername)

	// Query pending extra payments from Notion
	notionService := h.service.Notion.ContractorPayouts
	if notionService == nil {
		l.Debug("contractor payouts service is not initialized")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, fmt.Errorf("contractor payouts service not initialized"), nil, ""))
		return
	}

	entries, err := notionService.QueryPendingExtraPayments(c.Request.Context(), month, discordUsername)
	if err != nil {
		l.Error(err, "failed to query pending extra payments")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Build response - convert amounts to USD using Wise service for accurate exchange rates
	var contractors []ExtraPaymentContractor
	var totalAmount float64

	for _, entry := range entries {
		// Convert to USD using Wise service (same as invoice generation)
		amountUSD := entry.Amount
		if strings.ToUpper(entry.Currency) != "USD" && h.service.Wise != nil {
			convertedAmount, _, convErr := h.service.Wise.Convert(entry.Amount, entry.Currency, "USD")
			if convErr != nil {
				l.Error(convErr, fmt.Sprintf("failed to convert %s to USD for entry %s, using fallback", entry.Currency, entry.PageID))
				// Fallback to the pre-calculated AmountUSD (uses hardcoded rate)
				amountUSD = entry.AmountUSD
			} else {
				// Round to 2 decimal places
				amountUSD = math.Round(convertedAmount*100) / 100
				l.Debugf("converted %.2f %s to %.2f USD for entry %s", entry.Amount, entry.Currency, amountUSD, entry.PageID)
			}
		}

		contractors = append(contractors, ExtraPaymentContractor{
			PageID:  entry.PageID,
			Name:    entry.ContractorName,
			Discord: entry.Discord,
			Email:   entry.ContractorEmail,
			Amount:  amountUSD,
			Reason:  entry.Description,
		})
		totalAmount += amountUSD
	}

	response := PreviewExtraPaymentNotificationResponse{
		Count:       len(contractors),
		TotalAmount: totalAmount,
		Contractors: contractors,
	}

	l.Debugf("preview complete: count=%d totalAmount=%.2f", len(contractors), totalAmount)

	c.JSON(http.StatusOK, view.CreateResponse[any](response, nil, nil, nil, ""))
}

// SendExtraPaymentNotification godoc
// @Summary Send extra payment notification emails
// @Description Sends email notifications to contractors with pending Commission or Extra Payment payouts
// @Tags Notify
// @Accept json
// @Produce json
// @Param month query string true "Month in YYYY-MM format"
// @Param discord query string false "Optional Discord username filter"
// @Param test_email query string false "Optional test email to send all notifications to"
// @Param body body SendExtraPaymentNotificationRequest false "Optional custom reasons"
// @Success 200 {object} view.Response{data=SendExtraPaymentNotificationResponse}
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /api/v1/notify/extra-payment/send [post]
func (h *handler) SendExtraPaymentNotification(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "notify",
		"method":  "SendExtraPaymentNotification",
	})

	// Parse query params
	month := c.Query("month")
	if month == "" {
		l.Debug("month parameter is required")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, fmt.Errorf("month parameter is required (format: YYYY-MM)"), nil, ""))
		return
	}

	// Validate month format
	monthTime, err := time.Parse("2006-01", month)
	if err != nil {
		l.Debug("invalid month format")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, fmt.Errorf("invalid month format (expected: YYYY-MM)"), nil, ""))
		return
	}

	discordUsername := c.Query("discord")
	testEmail := c.Query("test_email")
	channelID := c.Query("channel_id")

	// Parse request body for custom reasons
	var req SendExtraPaymentNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Ignore JSON parsing errors - reasons are optional
		l.Debugf("no custom reasons provided: %v", err)
	}

	l.Debugf("sending extra payment notifications for month=%s discord=%s testEmail=%s channelID=%s customReasons=%d",
		month, discordUsername, testEmail, channelID, len(req.Reasons))

	// Query pending extra payments from Notion
	notionService := h.service.Notion.ContractorPayouts
	if notionService == nil {
		l.Debug("contractor payouts service is not initialized")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, fmt.Errorf("contractor payouts service not initialized"), nil, ""))
		return
	}

	entries, err := notionService.QueryPendingExtraPayments(c.Request.Context(), month, discordUsername)
	if err != nil {
		l.Error(err, "failed to query pending extra payments")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	if len(entries) == 0 {
		l.Debug("no pending extra payments found")
		c.JSON(http.StatusOK, view.CreateResponse[any](SendExtraPaymentNotificationResponse{
			Sent:   0,
			Failed: 0,
		}, nil, nil, nil, "no pending extra payments found"))
		return
	}

	// Wire ProgressBar if channelID is provided
	var pb *discordsvc.ProgressBar
	if channelID != "" && h.service.Discord != nil {
		initEmbed := buildNotifyProgressEmbed(0, len(entries))
		msg, discordErr := h.service.Discord.SendChannelMessageComplex(channelID, "", []*discordgo.MessageEmbed{initEmbed}, nil)
		if discordErr != nil {
			l.Error(discordErr, "failed to send initial discord progress message")
		} else if msg != nil {
			reporter := discordsvc.NewChannelMessageReporter(h.service.Discord, channelID, msg.ID)
			pb = discordsvc.NewProgressBar(reporter, l)
		}
	}

	// Format month for display (e.g., "January 2025")
	formattedMonth := monthTime.Format("January 2006")

	// Send emails
	gmailService := h.service.GoogleMail
	if gmailService == nil {
		l.Debug("gmail service is not initialized")
		if pb != nil {
			pb.Report(&discordgo.MessageEmbed{
				Title:       "❌ Failed",
				Description: "Gmail service not initialized.",
				Color:       15548997,
			})
		}
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, fmt.Errorf("gmail service not initialized"), nil, ""))
		return
	}

	// Use concurrency to send emails faster
	const maxConcurrent = 5 // Limit concurrent email sends to avoid rate limiting
	semaphore := make(chan struct{}, maxConcurrent)

	var (
		sent      int
		failed    int
		errors    []string
		completed int
		mu        sync.Mutex
		wg        sync.WaitGroup
	)

	total := len(entries)

	for _, entry := range entries {
		wg.Add(1)

		go func(entry notion.ExtraPaymentEntry) {
			defer wg.Done()

			// Acquire semaphore slot
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Determine email recipient
			recipientEmail := entry.ContractorEmail
			if testEmail != "" {
				recipientEmail = testEmail
			}

			if recipientEmail == "" {
				errMsg := fmt.Sprintf("no email for contractor %s (discord: %s)", entry.ContractorName, entry.Discord)
				l.Debug(errMsg)
				mu.Lock()
				errors = append(errors, errMsg)
				failed++
				completed++
				cur, tot := completed, total
				mu.Unlock()
				if pb != nil && (cur%3 == 0 || cur == tot) {
					pb.Report(buildNotifyProgressEmbed(cur, tot))
				}
				return
			}

			// Determine reasons to use
			var reasons []string
			if len(req.Reasons) > 0 {
				// Use custom reasons from request
				reasons = req.Reasons
			} else if entry.Description != "" {
				// Use Notion description as single reason
				reasons = []string{entry.Description}
			}

			// Convert to USD using Wise service for accurate exchange rates
			amountUSD := entry.Amount
			if strings.ToUpper(entry.Currency) != "USD" && h.service.Wise != nil {
				convertedAmount, _, convErr := h.service.Wise.Convert(entry.Amount, entry.Currency, "USD")
				if convErr != nil {
					l.Error(convErr, fmt.Sprintf("failed to convert %s to USD for entry %s, using fallback", entry.Currency, entry.PageID))
					// Fallback to the pre-calculated AmountUSD (uses hardcoded rate)
					amountUSD = entry.AmountUSD
				} else {
					// Round to 2 decimal places
					amountUSD = math.Round(convertedAmount*100) / 100
					l.Debugf("converted %.2f %s to %.2f USD for entry %s", entry.Amount, entry.Currency, amountUSD, entry.PageID)
				}
			}

			// Format amount (use USD)
			amountFormatted := fmt.Sprintf("$%.0f", amountUSD)
			if amountUSD != float64(int(amountUSD)) {
				amountFormatted = fmt.Sprintf("$%.2f", amountUSD)
			}

			// Build email data
			emailData := &model.ExtraPaymentNotificationEmail{
				ContractorName:  entry.ContractorName,
				ContractorEmail: recipientEmail,
				Month:           formattedMonth,
				Amount:          amountUSD,
				AmountFormatted: amountFormatted,
				Reasons:         reasons,
				SenderName:      "Team Dwarves",
			}

			// Send email
			if err := gmailService.SendExtraPaymentNotificationMail(emailData); err != nil {
				errMsg := fmt.Sprintf("failed to send email to %s: %v", recipientEmail, err)
				l.Error(err, errMsg)
				mu.Lock()
				errors = append(errors, errMsg)
				failed++
				completed++
				cur, tot := completed, total
				mu.Unlock()
				if pb != nil && (cur%3 == 0 || cur == tot) {
					pb.Report(buildNotifyProgressEmbed(cur, tot))
				}
				return
			}

			l.Debugf("sent extra payment notification to %s (%s)", entry.ContractorName, recipientEmail)
			mu.Lock()
			sent++
			completed++
			cur, tot := completed, total
			mu.Unlock()
			if pb != nil && (cur%3 == 0 || cur == tot) {
				pb.Report(buildNotifyProgressEmbed(cur, tot))
			}
		}(entry)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	if pb != nil {
		pb.Report(buildNotifyCompleteEmbed(sent, failed, errors))
	}

	response := SendExtraPaymentNotificationResponse{
		Sent:   sent,
		Failed: failed,
		Errors: errors,
	}

	l.Debugf("send complete: sent=%d failed=%d", sent, failed)

	c.JSON(http.StatusOK, view.CreateResponse[any](response, nil, nil, nil, ""))
}

// SendOneExtraPaymentNotification godoc
// @Summary Send extra payment notification to a single contractor
// @Description Sends email notification to a single contractor identified by page ID
// @Tags Notify
// @Accept json
// @Produce json
// @Param month query string true "Month in YYYY-MM format"
// @Param page_id query string true "Notion payout page ID"
// @Param test_email query string false "Optional test email to send notification to"
// @Param body body SendExtraPaymentNotificationRequest false "Optional custom reasons"
// @Success 200 {object} view.Response{data=SendOneExtraPaymentNotificationResponse}
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /api/v1/notify/extra-payment/send-one [post]
func (h *handler) SendOneExtraPaymentNotification(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "notify",
		"method":  "SendOneExtraPaymentNotification",
	})

	// Parse query params
	month := c.Query("month")
	if month == "" {
		l.Debug("month parameter is required")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, fmt.Errorf("month parameter is required (format: YYYY-MM)"), nil, ""))
		return
	}

	// Validate month format
	monthTime, err := time.Parse("2006-01", month)
	if err != nil {
		l.Debug("invalid month format")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, fmt.Errorf("invalid month format (expected: YYYY-MM)"), nil, ""))
		return
	}

	pageID := c.Query("page_id")
	if pageID == "" {
		l.Debug("page_id parameter is required")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, fmt.Errorf("page_id parameter is required"), nil, ""))
		return
	}

	testEmail := c.Query("test_email")

	// Parse request body for custom reasons
	var req SendExtraPaymentNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Ignore JSON parsing errors - reasons are optional
		l.Debugf("no custom reasons provided: %v", err)
	}

	l.Debugf("sending single extra payment notification pageID=%s month=%s testEmail=%s customReasons=%d",
		pageID, month, testEmail, len(req.Reasons))

	// Get the entry by page ID
	notionService := h.service.Notion.ContractorPayouts
	if notionService == nil {
		l.Debug("contractor payouts service is not initialized")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, fmt.Errorf("contractor payouts service not initialized"), nil, ""))
		return
	}

	entry, err := notionService.GetExtraPaymentEntryByPageID(c.Request.Context(), pageID)
	if err != nil {
		l.Error(err, "failed to get extra payment entry")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Format month for display (e.g., "January 2025")
	formattedMonth := monthTime.Format("January 2006")

	// Check Gmail service
	gmailService := h.service.GoogleMail
	if gmailService == nil {
		l.Debug("gmail service is not initialized")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, fmt.Errorf("gmail service not initialized"), nil, ""))
		return
	}

	// Determine email recipient
	recipientEmail := entry.ContractorEmail
	if testEmail != "" {
		recipientEmail = testEmail
	}

	response := SendOneExtraPaymentNotificationResponse{
		Name:  entry.ContractorName,
		Email: recipientEmail,
	}

	if recipientEmail == "" {
		errMsg := fmt.Sprintf("no email for contractor %s (discord: %s)", entry.ContractorName, entry.Discord)
		l.Debug(errMsg)
		response.Success = false
		response.Error = errMsg
		c.JSON(http.StatusOK, view.CreateResponse[any](response, nil, nil, nil, ""))
		return
	}

	// Determine reasons to use
	var reasons []string
	if len(req.Reasons) > 0 {
		// Use custom reasons from request
		reasons = req.Reasons
	} else if entry.Description != "" {
		// Use Notion description as single reason
		reasons = []string{entry.Description}
	}

	// Convert to USD using Wise service for accurate exchange rates
	amountUSD := entry.Amount
	if strings.ToUpper(entry.Currency) != "USD" && h.service.Wise != nil {
		convertedAmount, _, convErr := h.service.Wise.Convert(entry.Amount, entry.Currency, "USD")
		if convErr != nil {
			l.Error(convErr, fmt.Sprintf("failed to convert %s to USD for entry %s, using fallback", entry.Currency, entry.PageID))
			// Fallback to the pre-calculated AmountUSD (uses hardcoded rate)
			amountUSD = entry.AmountUSD
		} else {
			// Round to 2 decimal places
			amountUSD = math.Round(convertedAmount*100) / 100
			l.Debugf("converted %.2f %s to %.2f USD for entry %s", entry.Amount, entry.Currency, amountUSD, entry.PageID)
		}
	}

	// Format amount (use USD)
	amountFormatted := fmt.Sprintf("$%.0f", amountUSD)
	if amountUSD != float64(int(amountUSD)) {
		amountFormatted = fmt.Sprintf("$%.2f", amountUSD)
	}

	// Build email data
	emailData := &model.ExtraPaymentNotificationEmail{
		ContractorName:  entry.ContractorName,
		ContractorEmail: recipientEmail,
		Month:           formattedMonth,
		Amount:          amountUSD,
		AmountFormatted: amountFormatted,
		Reasons:         reasons,
		SenderName:      "Team Dwarves",
	}

	// Send email
	if err := gmailService.SendExtraPaymentNotificationMail(emailData); err != nil {
		errMsg := fmt.Sprintf("failed to send email: %v", err)
		l.Error(err, errMsg)
		response.Success = false
		response.Error = errMsg
		c.JSON(http.StatusOK, view.CreateResponse[any](response, nil, nil, nil, ""))
		return
	}

	l.Debugf("sent extra payment notification to %s (%s)", entry.ContractorName, recipientEmail)
	response.Success = true

	c.JSON(http.StatusOK, view.CreateResponse[any](response, nil, nil, nil, ""))
}

// buildNotifyProgressEmbed creates a Discord embed showing email sending progress.
func buildNotifyProgressEmbed(completed, total int) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "Sending Extra Payment Notifications",
		Description: fmt.Sprintf("Sending notifications...\n\n%s", discordsvc.BuildBar(completed, total)),
		Color:       5793266, // Discord Blurple
	}
}

// buildNotifyCompleteEmbed creates a Discord embed summarizing the completed send.
func buildNotifyCompleteEmbed(sent, failed int, errors []string) *discordgo.MessageEmbed {
	if failed == 0 {
		return &discordgo.MessageEmbed{
			Title:       "✅ Notifications Sent",
			Description: fmt.Sprintf("Successfully sent **%d** extra payment notification(s).", sent),
			Color:       3066993, // Green
		}
	}
	desc := fmt.Sprintf("Sent: **%d** | Failed: **%d**", sent, failed)
	if len(errors) > 0 && len(errors) <= 5 {
		desc += "\n\nErrors:\n• " + strings.Join(errors, "\n• ")
	}
	return &discordgo.MessageEmbed{
		Title:       "⚠️ Notifications Complete (with errors)",
		Description: desc,
		Color:       15105570, // Orange
	}
}
