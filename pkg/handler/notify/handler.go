package notify

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
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

	// Build response (use AmountUSD for consistent currency)
	var contractors []ExtraPaymentContractor
	var totalAmount float64

	for _, entry := range entries {
		contractors = append(contractors, ExtraPaymentContractor{
			Name:    entry.ContractorName,
			Discord: entry.Discord,
			Email:   entry.ContractorEmail,
			Amount:  entry.AmountUSD, // Use USD amount for display
			Reason:  entry.Description,
		})
		totalAmount += entry.AmountUSD
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

	// Parse request body for custom reasons
	var req SendExtraPaymentNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Ignore JSON parsing errors - reasons are optional
		l.Debugf("no custom reasons provided: %v", err)
	}

	l.Debugf("sending extra payment notifications for month=%s discord=%s testEmail=%s customReasons=%d",
		month, discordUsername, testEmail, len(req.Reasons))

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

	// Format month for display (e.g., "January 2025")
	formattedMonth := monthTime.Format("January 2006")

	// Send emails
	gmailService := h.service.GoogleMail
	if gmailService == nil {
		l.Debug("gmail service is not initialized")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, fmt.Errorf("gmail service not initialized"), nil, ""))
		return
	}

	// Use concurrency to send emails faster
	const maxConcurrent = 5 // Limit concurrent email sends to avoid rate limiting
	semaphore := make(chan struct{}, maxConcurrent)

	var (
		sent   int
		failed int
		errors []string
		mu     sync.Mutex
		wg     sync.WaitGroup
	)

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
				mu.Unlock()
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

			// Format amount (use USD)
			amountFormatted := fmt.Sprintf("$%.0f", entry.AmountUSD)
			if entry.AmountUSD != float64(int(entry.AmountUSD)) {
				amountFormatted = fmt.Sprintf("$%.2f", entry.AmountUSD)
			}

			// Build email data
			emailData := &model.ExtraPaymentNotificationEmail{
				ContractorName:  entry.ContractorName,
				ContractorEmail: recipientEmail,
				Month:           formattedMonth,
				Amount:          entry.AmountUSD,
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
				mu.Unlock()
				return
			}

			l.Debugf("sent extra payment notification to %s (%s)", entry.ContractorName, recipientEmail)
			mu.Lock()
			sent++
			mu.Unlock()
		}(entry)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	response := SendExtraPaymentNotificationResponse{
		Sent:   sent,
		Failed: failed,
		Errors: errors,
	}

	l.Debugf("send complete: sent=%d failed=%d", sent, failed)

	c.JSON(http.StatusOK, view.CreateResponse[any](response, nil, nil, nil, ""))
}
