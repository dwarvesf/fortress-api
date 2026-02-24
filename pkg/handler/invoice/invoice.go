package invoice

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	invoiceCtrl "github.com/dwarvesf/fortress-api/pkg/controller/invoice"
	"github.com/dwarvesf/fortress-api/pkg/handler/invoice/errs"
	"github.com/dwarvesf/fortress-api/pkg/handler/invoice/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	discordsvc "github.com/dwarvesf/fortress-api/pkg/service/discord"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils/authutils"
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/dwarvesf/fortress-api/pkg/worker"
)

type handler struct {
	controller *controller.Controller
	store      *store.Store
	service    *service.Service
	worker     *worker.Worker
	logger     logger.Logger
	repo       store.DBRepo
	config     *config.Config
}

// Concurrency constants for batch invoice processing
const (
	maxConcurrentInvoiceWorkers = 3 // Number of parallel workers for batch processing
)

// invoiceJob represents a job for the invoice worker pool
type invoiceJob struct {
	index      int
	contractor notion.ContractorRateData
}

// invoiceResult represents the result of processing a single contractor
type invoiceResult struct {
	index  int
	result view.BatchInvoiceResult
	total  float64
}

// New returns a handler
func New(ctrl *controller.Controller, store *store.Store, repo store.DBRepo, service *service.Service, worker *worker.Worker, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		controller: ctrl,
		store:      store,
		repo:       repo,
		service:    service,
		worker:     worker,
		logger:     logger,
		config:     cfg,
	}
}

// List godoc
// @Summary Get latest invoice by project id
// @Description Get latest invoice by project id
// @id getInvoiceList
// @Tags Invoice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param projectID query string false "projectID"
// @Param status query string false "status"
// @Param invoice_number query string false "invoice_number"
// @Param page query int false "page"
// @Param size query int false "size"
// @Param sort query string false "sort"
// @Success 200 {object} InvoiceListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /invoices [get]
func (h *handler) List(c *gin.Context) {
	var query request.GetListInvoiceInput
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "invoice",
		"method":  "GetLatestInvoice",
		"query":   query,
	})

	channelID := c.Query("channelId")

	pagination := query.StandardizeInput()

	if err := query.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	// Wire ProgressBar if channelID is provided
	var pb *discordsvc.ProgressBar
	if channelID != "" && h.service.Discord != nil {
		initEmbed := &discordgo.MessageEmbed{
			Title:       "⏳ Loading Invoices",
			Description: "Fetching invoices from Notion...",
			Color:       5793266,
		}
		msg, sendErr := h.service.Discord.SendChannelMessageComplex(channelID, "", []*discordgo.MessageEmbed{initEmbed}, nil)
		if sendErr != nil {
			l.Error(sendErr, "failed to send initial discord progress message")
		} else if msg != nil {
			reporter := discordsvc.NewChannelMessageReporter(h.service.Discord, channelID, msg.ID)
			pb = discordsvc.NewProgressBar(reporter, l)
		}
	}

	// Build progress callback that throttles updates (every 3 or at end)
	var onProgress func(completed, total int)
	if pb != nil {
		onProgress = func(completed, total int) {
			if completed%3 == 0 || completed == total {
				pb.Report(&discordgo.MessageEmbed{
					Title:       "⏳ Loading Invoices",
					Description: fmt.Sprintf("Fetching line items...\n\n%s", discordsvc.BuildBar(completed, total)),
					Color:       5793266,
				})
			}
		}
	}

	invoices, total, err := h.controller.Invoice.List(invoiceCtrl.GetListInvoiceInput{
		Pagination:    pagination,
		ProjectIDs:    query.ProjectID,
		Statuses:      query.Status,
		InvoiceNumber: query.InvoiceNumber,
		OnProgress:    onProgress,
	})
	if err != nil {
		l.Error(err, "failed to get latest invoice")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	rs, err := view.ToInvoiceListResponse(invoices)
	if err != nil {
		l.Error(err, "failed to parse invoice list response")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](rs, &view.PaginationResponse{Total: total, Pagination: view.Pagination{Page: pagination.Page, Size: pagination.Size, Sort: pagination.Sort}}, nil, nil, ""))
}

// GetTemplate godoc
// @Summary Get the latest invoice by project id
// @Description Get the latest invoice by project id
// @id getInvoiceTemplate
// @Tags Invoice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param projectID query string true "projectID"
// @Success 200 {object} InvoiceTemplateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /invoices/template [get]
func (h *handler) GetTemplate(c *gin.Context) {
	now := time.Now()
	var input request.GetInvoiceInput
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if input.ProjectID == "" || !model.IsUUIDFromString(input.ProjectID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidProjectID, input, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "invoice",
		"method":  "GetTemplate",
		"input":   input,
	})

	nextInvoiceNumber, lastInvoice, p, err := h.controller.Invoice.GetTemplate(invoiceCtrl.GetInvoiceInput{
		Now:       &now,
		ProjectID: input.ProjectID,
	})
	if err != nil {
		l.Error(err, "failed to get invoice template")
		errs.ConvertControllerErr(c, err)
		return
	}

	rs, err := view.ToInvoiceTemplateResponse(p, lastInvoice, nextInvoiceNumber)
	if err != nil {
		l.Error(err, "failed to parse invoice template response")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](rs, nil, nil, nil, ""))
}

// Send godoc
// @Summary Create new invoice and send to client
// @Description Create new invoice and send to client
// @id sendInvoice
// @Tags Invoice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Body body SendInvoiceRequest true "body"
// @Success 200 {object} MessageResponse
// @Failure 404 {object} ErrorResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /invoices/send [post]
func (h *handler) Send(c *gin.Context) {
	userID, err := authutils.GetUserIDFromContext(c, h.config)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	var req request.SendInvoiceRequest

	l := h.logger.Fields(logger.Fields{
		"handler": "invoice",
		"method":  "Send",
	})

	if err := req.ValidateAndMappingRequest(c, h.config); err != nil {
		l.Errorf(err, "failed to validating and mapping the quest", "input", req)
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	// DEBUG: Log frontend-provided values before server-side calculation
	var frontendSubTotal float64
	for _, item := range req.LineItems {
		frontendSubTotal += item.Cost
	}
	l.Debugf("Invoice totals - Frontend: SubTotal=%.2f Total=%.2f Tax=%.2f Discount=%.2f | LineItemsSum=%.2f",
		req.SubTotal, req.Total, req.Tax, req.Discount, frontendSubTotal)

	iv, err := req.ToInvoiceModel(userID)
	if err != nil {
		l.Error(err, "failed to parse request to invoice model")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	// DEBUG: Log server-calculated values after ToInvoiceModel
	l.Debugf("Invoice totals - Server-calculated: SubTotal=%.2f Total=%.2f | Expected Total=%.2f",
		iv.SubTotal, iv.Total, iv.SubTotal+iv.Tax-iv.Discount)

	_, err = h.controller.Invoice.Send(iv)
	if err != nil {
		l.Error(err, "failed to send invoice")
		errs.ConvertControllerErr(c, err)
		return
	}

	// send message to discord channel
	err = h.controller.Discord.Log(model.LogDiscordInput{
		Type: "invoice_send",
		Data: map[string]interface{}{
			"invoice_number": iv.Number,
			"employee_id":    userID,
		},
	})
	if err != nil {
		l.Error(err, "failed to log to discord")
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

// UpdateStatus godoc
// @Summary Update status for invoice
// @Description Update status for invoice
// @Tags Invoice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} MessageResponse
// @Failure 404 {object} ErrorResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /invoices/{id}/status [put]
func (h *handler) UpdateStatus(c *gin.Context) {
	invoiceID := c.Param("id")
	if invoiceID == "" || !model.IsUUIDFromString(invoiceID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidInvoiceID, nil, ""))
		return
	}

	var req request.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "invoice",
		"method":  "UpdateStatus",
		"req":     req,
	})

	l.Debugf("received update status request: invoiceID=%s targetStatus=%v sendThankYouEmail=%v", invoiceID, req.Status, req.SendThankYouEmail)

	if err := req.Validate(); err != nil {
		l.Error(err, "invalid request")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	l.Debug("request validation passed, calling controller")

	// check invoice existence
	_, err := h.controller.Invoice.UpdateStatus(invoiceCtrl.UpdateStatusInput{
		InvoiceID:         invoiceID,
		Status:            req.Status,
		SendThankYouEmail: req.SendThankYouEmail,
	})
	if err != nil {
		l.Error(err, "failed to update invoice status")
		errs.ConvertControllerErr(c, err)
		return
	}

	l.Debug("invoice status updated successfully, returning response")

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

// CalculateCommissions godoc
// @Summary Calculate commissions for an invoice
// @Description Calculate commissions for an invoice, with optional dry run
// @Tags Invoice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Invoice ID"
// @Param dry_run query bool false "Dry run (do not save, just return calculation)"
// @Success 200 {object} []model.EmployeeCommission
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /invoices/{id}/calculate-commissions [post]
func (h *handler) CalculateCommissions(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "invoice",
		"method":  "CalculateCommissions",
	})

	invoiceID := c.Param("id")
	if invoiceID == "" {
		l.Error(errors.New("missing invoice id"), "missing invoice id")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errors.New("missing invoice id"), nil, ""))
		return
	}

	dryRun := c.DefaultQuery("dry_run", "false") == "true"

	commissions, err := h.controller.Invoice.ProcessCommissions(invoiceID, dryRun, l)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](commissions, nil, nil, nil, ""))
}

// GenerateContractorInvoice godoc
// @Summary Generate a contractor invoice
// @Description Generate a contractor invoice PDF based on contractor discord and month
// @id generateContractorInvoice
// @Tags Invoice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Body body request.GenerateContractorInvoiceRequest true "body"
// @Success 200 {object} view.ContractorInvoiceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /invoices/contractor/generate [post]
func (h *handler) GenerateContractorInvoice(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "invoice",
		"method":  "GenerateContractorInvoice",
	})

	l.Debug("handling generate contractor invoice request")

	// 1. Parse request
	var req request.GenerateContractorInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		l.Error(err, "invalid request body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	l.Debug(fmt.Sprintf("received request: contractor=%s month=%s skipUpload=%v dryRun=%v invoiceType=%s batch=%d channelID=%s",
		req.Contractor, req.Month, req.SkipUpload, req.DryRun, req.InvoiceType, req.Batch, req.ChannelID))

	// 2. Validate month format (YYYY-MM) - only if month is provided
	if req.Month != "" && !isValidMonthFormat(req.Month) {
		l.Error(errs.ErrInvalidMonthFormat, fmt.Sprintf("month validation failed: month=%s", req.Month))
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidMonthFormat, req, ""))
		return
	}

	// 3. Handle batch dry-run for "all" contractors (synchronous)
	if req.Contractor == "all" && req.Batch > 0 && req.DryRun {
		l.Debug(fmt.Sprintf("batch dry-run mode: batch=%d month=%s", req.Batch, req.Month))
		h.processBatchDryRun(c, l, req)
		return
	}

	// 4. Handle batch processing for "all" contractors (async)
	if req.Contractor == "all" && req.Batch > 0 {
		l.Debug(fmt.Sprintf("batch processing mode: batch=%d channelID=%s", req.Batch, req.ChannelID))

		// Return 200 OK immediately and process asynchronously
		c.JSON(http.StatusOK, view.CreateResponse(view.BatchInvoiceResponse{
			Message: "Invoice generation started",
			Batch:   req.Batch,
			Month:   req.Month,
		}, nil, nil, req, ""))

		// Process batch asynchronously
		go h.processBatchInvoices(l, req)
		return
	}

	// 4. Build options from request (single contractor flow)
	opts := &invoiceCtrl.ContractorInvoiceOptions{
		GroupFeeByProject: false,           // default: Commission items displayed individually in Extra Payment section
		InvoiceType:       req.InvoiceType, // "service_and_refund" | "extra_payment" | "" (full)
	}
	if req.GroupFeeByProject != nil {
		opts.GroupFeeByProject = *req.GroupFeeByProject
	}

	// 5. Generate invoice data
	var invoiceData *invoiceCtrl.ContractorInvoiceData
	var err error
	if req.DryRun {
		// Dry-run: skip force sync to avoid creating payout records in Notion
		l.Debug("dry-run mode: generating invoice data without force sync")
		invoiceData, err = h.controller.Invoice.GenerateContractorInvoice(c.Request.Context(), req.Contractor, req.Month, opts)
	} else {
		l.Debug("calling controller to generate contractor invoice data with force sync")
		invoiceData, err = h.controller.Invoice.GenerateContractorInvoiceWithForceSync(c.Request.Context(), req.Contractor, req.Month, req.Batch, opts)
	}
	if err != nil {
		l.Error(err, "failed to generate contractor invoice")
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, req, ""))
			return
		}
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	l.Debug(fmt.Sprintf("invoice data generated: invoiceNumber=%s total=%.2f", invoiceData.InvoiceNumber, invoiceData.Total))

	// 4. Generate PDF
	l.Debug("generating PDF")
	pdfBytes, err := h.controller.Invoice.GenerateContractorInvoicePDF(l, invoiceData)
	if err != nil {
		l.Error(err, "failed to generate PDF")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	l.Debug(fmt.Sprintf("PDF generated: size=%d bytes", len(pdfBytes)))

	// 5. Upload to Google Drive or save locally
	var fileURL string
	fileName := fmt.Sprintf("%s.pdf", invoiceData.InvoiceNumber)

	if req.SkipUpload || req.DryRun {
		l.Debug("[DEBUG] skipping Google Drive upload, saving to local file")

		// Save to local file
		localDir := filepath.Join(os.TempDir(), "contractor-invoices")
		if err := os.MkdirAll(localDir, 0755); err != nil {
			l.Error(err, "failed to create local directory")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
			return
		}

		localPath := filepath.Join(localDir, fileName)
		if err := os.WriteFile(localPath, pdfBytes, 0644); err != nil {
			l.Error(err, "failed to write PDF to local file")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
			return
		}

		l.Debug(fmt.Sprintf("[DEBUG] PDF saved locally: path=%s", localPath))
		fileURL = localPath
	} else {
		l.Debug(fmt.Sprintf("uploading PDF to Google Drive: fileName=%s contractorName=%s", fileName, invoiceData.ContractorName))

		fileURL, err = h.service.GoogleDrive.UploadContractorInvoicePDF(invoiceData.ContractorName, fileName, pdfBytes)
		if err != nil {
			l.Error(err, "failed to upload PDF to Google Drive")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
			return
		}

		l.Debug(fmt.Sprintf("PDF uploaded: url=%s", fileURL))

		// Share with contractor if requested
		if req.Share && invoiceData.ContractorEmail != "" {
			l.Debug(fmt.Sprintf("sharing PDF with contractor: email=%s", invoiceData.ContractorEmail))

			// Extract file ID from URL (format: https://drive.google.com/file/d/{fileID}/view)
			fileID := extractFileIDFromURL(fileURL)
			if fileID != "" {
				if err := h.service.GoogleDrive.ShareFileWithEmail(fileID, invoiceData.ContractorEmail); err != nil {
					l.Error(err, fmt.Sprintf("failed to share PDF with contractor: email=%s", invoiceData.ContractorEmail))
					// Non-fatal: PDF was uploaded, just sharing failed
				} else {
					l.Info(fmt.Sprintf("PDF shared with contractor: email=%s fileID=%s", invoiceData.ContractorEmail, fileID))
				}
			} else {
				l.Warn(fmt.Sprintf("could not extract file ID from URL to share: %s", fileURL))
			}
		} else if req.Share && invoiceData.ContractorEmail == "" {
			l.Warn("share requested but contractor email is empty, skipping share")
		}
	}

	// 5.5 Create Contractor Payables record in Notion (skip in dry-run mode)
	if req.DryRun {
		l.Debug("[DEBUG] dry-run mode: skipping contractor payables record creation")
	} else {
		l.Debug("[DEBUG] creating contractor payables record in Notion")

		// Calculate period dates based on payday from invoice data
		// PeriodStart: payday of invoice month
		// PeriodEnd: payday of next month
		monthTime, _ := time.Parse("2006-01", invoiceData.Month)
		payday := invoiceData.PayDay
		if payday == 0 {
			payday = 15 // Default to 15 if not set
		}
		periodStart := time.Date(monthTime.Year(), monthTime.Month(), payday, 0, 0, 0, 0, time.UTC)
		nextMonth := monthTime.AddDate(0, 1, 0)
		periodEnd := time.Date(nextMonth.Year(), nextMonth.Month(), payday, 0, 0, 0, 0, time.UTC)

		l.Debug(fmt.Sprintf("[DEBUG] calculated period: payday=%d start=%s end=%s",
			payday, periodStart.Format("2006-01-02"), periodEnd.Format("2006-01-02")))

		payableInput := notion.CreatePayableInput{
			ContractorPageID: invoiceData.ContractorPageID,
			Total:            invoiceData.TotalUSD,
			Currency:         "USD",
			PeriodStart:      periodStart.Format("2006-01-02"),
			PeriodEnd:        periodEnd.Format("2006-01-02"),
			InvoiceDate:      time.Now().Format("2006-01-02"),
			InvoiceID:        invoiceData.InvoiceNumber,
			PayoutItemIDs:    invoiceData.PayoutPageIDs,
			ContractorType:   "Individual", // Default to Individual
			ExchangeRate:     invoiceData.ExchangeRate,
			PDFBytes:         pdfBytes, // Upload PDF to Notion
		}

		l.Debug(fmt.Sprintf("[DEBUG] payable input: contractor=%s total=%.2f payoutItems=%d periodStart=%s periodEnd=%s",
			payableInput.ContractorPageID, payableInput.Total, len(payableInput.PayoutItemIDs), payableInput.PeriodStart, payableInput.PeriodEnd))

		payablePageID, payableErr := h.service.Notion.ContractorPayables.CreatePayable(c.Request.Context(), payableInput)
		if payableErr != nil {
			l.Error(payableErr, "[DEBUG] failed to create contractor payables record - continuing with response")
			// Non-fatal: continue with response
		} else {
			l.Debug(fmt.Sprintf("[DEBUG] contractor payables record created: pageID=%s", payablePageID))
		}
	}

	// 6. Build response
	lineItems := make([]view.ContractorInvoiceLineItem, len(invoiceData.LineItems))
	for i, item := range invoiceData.LineItems {
		lineItems[i] = view.ContractorInvoiceLineItem{
			Title:       item.Title,
			Description: item.Description,
			Hours:       item.Hours,
			Rate:        item.Rate,
			Amount:      item.Amount,
		}
	}

	response := view.ContractorInvoiceResponse{
		InvoiceNumber:  invoiceData.InvoiceNumber,
		ContractorName: invoiceData.ContractorName,
		Month:          invoiceData.Month,
		BillingType:    invoiceData.BillingType,
		Currency:       invoiceData.Currency,
		Total:          invoiceData.Total,
		PDFFileURL:     fileURL,
		GeneratedAt:    time.Now().Format(time.RFC3339),
		LineItems:      lineItems,
	}

	l.Info(fmt.Sprintf("contractor invoice generated successfully: invoice_number=%s", response.InvoiceNumber))

	// Send Discord message with result if channelID is provided (API creates the message)
	if req.ChannelID != "" {
		l.Debug(fmt.Sprintf("sending Discord result message: channelID=%s", req.ChannelID))
		h.sendDiscordSingleInvoiceSuccess(l, req.ChannelID, invoiceData, fileURL)
	}

	c.JSON(http.StatusOK, view.CreateResponse(response, nil, nil, req, ""))
}

// sendDiscordSingleInvoiceSuccess sends a Discord embed with single invoice success (API creates a new message)
func (h *handler) sendDiscordSingleInvoiceSuccess(l logger.Logger, channelID string, invoiceData interface{}, fileURL string) {
	// Type assert to get invoice data
	data, ok := invoiceData.(*invoiceCtrl.ContractorInvoiceData)
	if !ok {
		l.Error(nil, "failed to cast invoice data for Discord update")
		return
	}

	displayMonth := formatMonthDisplay(data.Month)

	description := fmt.Sprintf("**Month:** %s\n", displayMonth)
	description += fmt.Sprintf("**Contractor:** %s\n", data.ContractorName)
	description += fmt.Sprintf("**Invoice Number:** %s\n", data.InvoiceNumber)
	description += fmt.Sprintf("**Total:** USD %.2f\n", data.TotalUSD)
	description += fmt.Sprintf("**Line Items:** %d\n", len(data.LineItems))

	if fileURL != "" {
		description += fmt.Sprintf("\n[View Invoice](%s)", fileURL)
	}

	embed := &discordgo.MessageEmbed{
		Title:       "✅ Invoice Generation Successful",
		Description: description,
		Color:       5763719, // Green
	}

	_, err := h.service.Discord.SendChannelMessageComplex(channelID, "", []*discordgo.MessageEmbed{embed}, nil)
	if err != nil {
		l.Error(err, "failed to send discord single invoice success message")
	} else {
		l.Debug("discord single invoice success message sent")
	}
}

// processBatchInvoices handles async batch invoice generation for all contractors in a batch
// Uses a worker pool pattern for concurrent processing
func (h *handler) processBatchInvoices(l logger.Logger, req request.GenerateContractorInvoiceRequest) {
	ctx := context.Background()

	l.Debug(fmt.Sprintf("starting batch invoice processing: batch=%d month=%s workers=%d",
		req.Batch, req.Month, maxConcurrentInvoiceWorkers))

	// 0. Create initial Discord progress message if channelID is provided (API owns the message lifecycle)
	var messageID string
	if req.ChannelID != "" {
		displayMonth := formatMonthDisplay(req.Month)
		initEmbed := &discordgo.MessageEmbed{
			Title:       "⏳ Generating Contractor Invoices",
			Description: fmt.Sprintf("Processing invoices for **%s** (batch %d)...\n\nThis may take a few moments.", displayMonth, req.Batch),
			Color:       5793266, // Discord Blurple
			Footer:      &discordgo.MessageEmbedFooter{Text: "Processing..."},
		}
		msg, err := h.service.Discord.SendChannelMessageComplex(req.ChannelID, "", []*discordgo.MessageEmbed{initEmbed}, nil)
		if err != nil {
			l.Error(err, "failed to send initial discord progress message")
		} else if msg != nil {
			messageID = msg.ID
			l.Debug(fmt.Sprintf("created discord progress message: messageID=%s channelID=%s", messageID, req.ChannelID))
		}
	}

	// 1. Get list of contractors for this batch
	ratesService := notion.NewContractorRatesService(h.config, h.logger)
	if ratesService == nil {
		l.Error(nil, "failed to create contractor rates service")
		h.updateDiscordWithError(l, req.ChannelID, messageID, "Failed to initialize contractor rates service")
		return
	}

	contractors, err := ratesService.ListActiveContractorsByBatch(ctx, req.Month, req.Batch)
	if err != nil {
		l.Error(err, fmt.Sprintf("failed to list contractors for batch %d", req.Batch))
		h.updateDiscordWithError(l, req.ChannelID, messageID, fmt.Sprintf("Failed to list contractors: %v", err))
		return
	}

	if len(contractors) == 0 {
		l.Debug(fmt.Sprintf("no contractors found for batch %d", req.Batch))
		h.updateDiscordWithNoContractors(l, req.ChannelID, messageID, req.Month, req.Batch)
		return
	}

	l.Debug(fmt.Sprintf("found %d contractors for batch %d", len(contractors), req.Batch))

	// 2. Process contractors concurrently using worker pool
	opts := &invoiceCtrl.ContractorInvoiceOptions{
		GroupFeeByProject: false,
		InvoiceType:       req.InvoiceType,
	}

	// Create channels for job distribution and result collection
	jobs := make(chan invoiceJob, len(contractors))
	resultsChan := make(chan invoiceResult, len(contractors))

	// Start worker pool
	var wg sync.WaitGroup
	numWorkers := maxConcurrentInvoiceWorkers
	if len(contractors) < numWorkers {
		numWorkers = len(contractors) // Don't spawn more workers than jobs
	}

	l.Debug(fmt.Sprintf("starting %d invoice workers", numWorkers))

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go h.invoiceWorker(l, ctx, req.Month, opts, &wg, jobs, resultsChan)
	}

	// Send jobs to workers
	for i, contractor := range contractors {
		jobs <- invoiceJob{
			index:      i,
			contractor: contractor,
		}
	}
	close(jobs)

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Build a ProgressBar for Discord updates (nil-safe: pb.Report is a no-op when pb is nil)
	var pb *discordsvc.ProgressBar
	if req.ChannelID != "" && messageID != "" {
		reporter := discordsvc.NewChannelMessageReporter(h.service.Discord, req.ChannelID, messageID)
		pb = discordsvc.NewProgressBar(reporter, l)
	}

	// Collect results and update Discord progressively
	allResults := make([]view.BatchInvoiceResult, len(contractors))
	var successCount int
	var totalAmount float64
	completed := 0

	for r := range resultsChan {
		allResults[r.index] = r.result
		completed++

		if r.result.Success {
			successCount++
			totalAmount += r.total
		}

		// Update Discord progress every 3 completions or at the end
		if completed%3 == 0 || completed == len(contractors) {
			// Build current results slice (only completed ones)
			var currentResults []view.BatchInvoiceResult
			for _, res := range allResults {
				if res.Contractor != "" {
					currentResults = append(currentResults, res)
				}
			}
			h.updateDiscordWithProgress(pb, req.Month, req.Batch,
				completed, len(contractors), fmt.Sprintf("%d completed", completed), currentResults)
		}

		l.Debug(fmt.Sprintf("batch progress: %d/%d completed, %d success, total=%.2f",
			completed, len(contractors), successCount, totalAmount))
	}

	// Convert to final results slice
	results := make([]view.BatchInvoiceResult, 0, len(contractors))
	for _, r := range allResults {
		if r.Contractor != "" {
			results = append(results, r)
		}
	}

	// 3. Send final summary
	h.updateDiscordWithBatchComplete(pb, req.Month, req.Batch,
		successCount, len(contractors), totalAmount, results)

	l.Info(fmt.Sprintf("batch invoice processing completed: batch=%d success=%d/%d total=%.2f workers=%d",
		req.Batch, successCount, len(contractors), totalAmount, numWorkers))
}

// processBatchDryRun handles synchronous dry-run for all contractors in a batch.
// It lists contractors, generates invoice data for each (no PDF/upload/payables),
// and returns a summary response compatible with the Discord bot's GenerateContractorInvoicesData.
func (h *handler) processBatchDryRun(c *gin.Context, l logger.Logger, req request.GenerateContractorInvoiceRequest) {
	ctx := c.Request.Context()

	// 1. Get list of contractors for this batch
	ratesService := notion.NewContractorRatesService(h.config, h.logger)
	if ratesService == nil {
		l.Error(nil, "failed to create contractor rates service")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, fmt.Errorf("failed to initialize contractor rates service"), req, ""))
		return
	}

	contractors, err := ratesService.ListActiveContractorsByBatch(ctx, req.Month, req.Batch)
	if err != nil {
		l.Error(err, fmt.Sprintf("failed to list contractors for batch %d", req.Batch))
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, fmt.Errorf("failed to list contractors: %w", err), req, ""))
		return
	}

	if len(contractors) == 0 {
		l.Debug(fmt.Sprintf("no contractors found for batch %d", req.Batch))
		c.JSON(http.StatusOK, view.CreateResponse(view.BatchDryRunResponse{
			Month: req.Month,
			Batch: req.Batch,
		}, nil, nil, req, fmt.Sprintf("No contractors found for batch %d", req.Batch)))
		return
	}

	l.Debug(fmt.Sprintf("batch dry-run: found %d contractors for batch %d", len(contractors), req.Batch))

	// 2. Generate invoice data for each contractor (no PDF, no upload, no payables)
	opts := &invoiceCtrl.ContractorInvoiceOptions{
		GroupFeeByProject: false,
		InvoiceType:       req.InvoiceType,
	}

	var contractorNames []string
	var totalAmount float64

	for _, contractor := range contractors {
		l.Debug(fmt.Sprintf("batch dry-run: processing contractor %s", contractor.Discord))

		// Dry-run: skip force sync to avoid creating payout records in Notion.
		// Only preview what would happen with the current state.
		invoiceData, err := h.controller.Invoice.GenerateContractorInvoice(ctx, contractor.Discord, req.Month, opts)
		if err != nil {
			if strings.Contains(err.Error(), "no pending payouts") {
				l.Debug(fmt.Sprintf("batch dry-run: skipping %s - no pending payouts", contractor.Discord))
				contractorNames = append(contractorNames, fmt.Sprintf("%s (skipped: no pending payouts)", contractor.Discord))
			} else {
				l.Error(err, fmt.Sprintf("batch dry-run: failed for %s", contractor.Discord))
				contractorNames = append(contractorNames, fmt.Sprintf("%s (error)", contractor.Discord))
			}
			continue
		}

		totalAmount += invoiceData.TotalUSD
		contractorNames = append(contractorNames, fmt.Sprintf("%s (USD %.2f, %d items)", contractor.Discord, invoiceData.TotalUSD, len(invoiceData.LineItems)))
	}

	l.Debug(fmt.Sprintf("batch dry-run completed: %d contractors processed, total=%.2f", len(contractorNames), totalAmount))

	response := view.BatchDryRunResponse{
		Month:       req.Month,
		Batch:       req.Batch,
		Total:       totalAmount,
		Currency:    "USD",
		Contractors: contractorNames,
	}

	c.JSON(http.StatusOK, view.CreateResponse(response, nil, nil, req,
		fmt.Sprintf("Dry run completed for %d contractors in batch %d. Total: USD %.2f", len(contractors), req.Batch, totalAmount)))
}

// invoiceWorker processes invoice jobs from the jobs channel and sends results to resultsChan
func (h *handler) invoiceWorker(l logger.Logger, ctx context.Context, month string,
	opts *invoiceCtrl.ContractorInvoiceOptions, wg *sync.WaitGroup, jobs <-chan invoiceJob, resultsChan chan<- invoiceResult) {
	defer wg.Done()

	for job := range jobs {
		result := h.processContractorInvoice(l, ctx, month, opts, job.contractor)
		resultsChan <- invoiceResult{
			index:  job.index,
			result: result,
			total:  result.Total,
		}
	}
}

// processContractorInvoice handles the complete invoice generation for a single contractor
func (h *handler) processContractorInvoice(l logger.Logger, ctx context.Context, month string,
	opts *invoiceCtrl.ContractorInvoiceOptions, contractor notion.ContractorRateData) view.BatchInvoiceResult {
	l.Debug(fmt.Sprintf("processing contractor: %s", contractor.Discord))

	// Generate invoice data
	invoiceData, err := h.controller.Invoice.GenerateContractorInvoice(ctx, contractor.Discord, month, opts)
	if err != nil {
		// Check if this is a "no pending payouts" case - treat as skipped, not error
		if strings.Contains(err.Error(), "no pending payouts") {
			l.Debug(fmt.Sprintf("skipping contractor %s: no pending payouts", contractor.Discord))
			return view.BatchInvoiceResult{
				Contractor: contractor.Discord,
				Success:    false,
				Skipped:    true,
				SkipReason: "no pending payouts",
			}
		}

		l.Error(err, fmt.Sprintf("failed to generate invoice for contractor %s", contractor.Discord))
		return view.BatchInvoiceResult{
			Contractor: contractor.Discord,
			Success:    false,
			Error:      err.Error(),
		}
	}

	// Generate PDF
	pdfBytes, err := h.controller.Invoice.GenerateContractorInvoicePDF(l, invoiceData)
	if err != nil {
		l.Error(err, fmt.Sprintf("failed to generate PDF for contractor %s", contractor.Discord))
		return view.BatchInvoiceResult{
			Contractor: contractor.Discord,
			Success:    false,
			Error:      fmt.Sprintf("PDF generation failed: %v", err),
		}
	}

	// Upload to Google Drive
	fileName := fmt.Sprintf("%s.pdf", invoiceData.InvoiceNumber)
	fileURL, err := h.service.GoogleDrive.UploadContractorInvoicePDF(invoiceData.ContractorName, fileName, pdfBytes)
	if err != nil {
		l.Error(err, fmt.Sprintf("failed to upload PDF for contractor %s", contractor.Discord))
		return view.BatchInvoiceResult{
			Contractor: contractor.Discord,
			Success:    false,
			Error:      fmt.Sprintf("Upload failed: %v", err),
		}
	}

	l.Debug(fmt.Sprintf("uploaded PDF for %s: %s", contractor.Discord, fileURL))

	// Create Contractor Payables record in Notion
	monthTime, _ := time.Parse("2006-01", invoiceData.Month)
	payday := invoiceData.PayDay
	if payday == 0 {
		payday = 15
	}
	periodStart := time.Date(monthTime.Year(), monthTime.Month(), payday, 0, 0, 0, 0, time.UTC)
	nextMonth := monthTime.AddDate(0, 1, 0)
	periodEnd := time.Date(nextMonth.Year(), nextMonth.Month(), payday, 0, 0, 0, 0, time.UTC)

	payableInput := notion.CreatePayableInput{
		ContractorPageID: invoiceData.ContractorPageID,
		Total:            invoiceData.TotalUSD,
		Currency:         "USD",
		PeriodStart:      periodStart.Format("2006-01-02"),
		PeriodEnd:        periodEnd.Format("2006-01-02"),
		InvoiceDate:      time.Now().Format("2006-01-02"),
		InvoiceID:        invoiceData.InvoiceNumber,
		PayoutItemIDs:    invoiceData.PayoutPageIDs,
		ContractorType:   "Individual",
		ExchangeRate:     invoiceData.ExchangeRate,
		PDFBytes:         pdfBytes,
	}

	_, payableErr := h.service.Notion.ContractorPayables.CreatePayable(ctx, payableInput)
	if payableErr != nil {
		l.Error(payableErr, fmt.Sprintf("failed to create payable record for %s - continuing", contractor.Discord))
		// Non-fatal: continue with success
	}

	l.Debug(fmt.Sprintf("invoice generation completed for %s: total=%.2f", contractor.Discord, invoiceData.TotalUSD))

	return view.BatchInvoiceResult{
		Contractor: contractor.Discord,
		Success:    true,
		Total:      invoiceData.TotalUSD,
		Currency:   "USD",
	}
}

// updateDiscordWithProgress updates the Discord embed with current progress via ProgressBar.
func (h *handler) updateDiscordWithProgress(pb *discordsvc.ProgressBar, month string, batch int,
	current, total int, currentContractor string, results []view.BatchInvoiceResult) {
	if pb == nil {
		return
	}

	displayMonth := formatMonthDisplay(month)

	// Build progress description
	description := fmt.Sprintf("Processing invoices for **%s**...\n\n", displayMonth)
	description += fmt.Sprintf("**Progress:** %d/%d contractors\n", current, total)
	description += fmt.Sprintf("**Current:** %s\n\n", currentContractor)

	// Show recent results (last 5)
	startIdx := 0
	if len(results) > 5 {
		startIdx = len(results) - 5
	}
	for _, r := range results[startIdx:] {
		if r.Success {
			description += fmt.Sprintf("✓ %s - USD %.2f\n", r.Contractor, r.Total)
		} else if r.Skipped {
			description += fmt.Sprintf("⊘ %s - skipped\n", r.Contractor)
		} else {
			description += fmt.Sprintf("✗ %s - %s\n", r.Contractor, r.Error)
		}
	}
	description += fmt.Sprintf("→ %s (processing...)", currentContractor)

	embed := &discordgo.MessageEmbed{
		Title:       "⏳ Generating Contractor Invoices",
		Description: description,
		Color:       5793266, // Blue
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Processing contractor %d of %d...", current, total),
		},
	}

	pb.Report(embed)
}

// updateDiscordWithBatchComplete sends the final success/partial failure embed via ProgressBar.
func (h *handler) updateDiscordWithBatchComplete(pb *discordsvc.ProgressBar, month string, batch int,
	successCount, totalCount int, totalAmount float64, results []view.BatchInvoiceResult) {
	if pb == nil {
		return
	}

	displayMonth := formatMonthDisplay(month)

	// Count skipped and failed separately
	var skippedCount, failedCount int
	var skippedContractors []string
	for _, r := range results {
		if r.Skipped {
			skippedCount++
			skippedContractors = append(skippedContractors, r.Contractor)
		} else if !r.Success {
			failedCount++
		}
	}

	var title string
	var color int

	// Determine title and color based on actual errors (not skipped)
	if failedCount == 0 {
		title = "✅ Invoice Generation Completed"
		color = 5763719 // Green
	} else if successCount > 0 {
		title = "⚠️ Invoice Generation Completed with Errors"
		color = 16776960 // Orange
	} else {
		title = "❌ Invoice Generation Failed"
		color = 15548997 // Red
	}

	description := fmt.Sprintf("**Month:** %s\n", displayMonth)
	description += fmt.Sprintf("**Batch:** %d (payday %dth)\n\n", batch, batch)
	description += "**Summary:**\n"
	description += fmt.Sprintf("• Generated: %d invoices\n", successCount)
	if skippedCount > 0 {
		description += fmt.Sprintf("• Skipped: %d contractors\n", skippedCount)
	}
	if failedCount > 0 {
		description += fmt.Sprintf("• Failed: %d invoices\n", failedCount)
	}
	description += fmt.Sprintf("• Total Amount: USD %.2f\n\n", totalAmount)

	// List successful contractors (limit to 15)
	if successCount > 0 {
		var successContractors []string
		for _, r := range results {
			if r.Success {
				successContractors = append(successContractors, r.Contractor)
			}
		}
		description += "**Contractors:**\n"
		displayCount := len(successContractors)
		if displayCount > 15 {
			displayCount = 15
		}
		description += strings.Join(successContractors[:displayCount], ", ")
		if len(successContractors) > 15 {
			description += fmt.Sprintf("\n... and %d more", len(successContractors)-15)
		}
	}

	// List skipped contractors (informational, not errors)
	if skippedCount > 0 {
		description += "\n\n**Skipped** (no pending payouts):\n"
		displayCount := len(skippedContractors)
		if displayCount > 15 {
			displayCount = 15
		}
		description += strings.Join(skippedContractors[:displayCount], ", ")
		if len(skippedContractors) > 15 {
			description += fmt.Sprintf("\n... and %d more", len(skippedContractors)-15)
		}
	}

	// List actual errors if any
	if failedCount > 0 {
		description += "\n\n**Errors:**\n"
		for _, r := range results {
			if !r.Success && !r.Skipped {
				description += fmt.Sprintf("• %s: %s\n", r.Contractor, r.Error)
			}
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       color,
	}

	pb.Report(embed)
}

// updateDiscordWithError updates the Discord embed with an error message
func (h *handler) updateDiscordWithError(l logger.Logger, channelID, messageID, errorMsg string) {
	if channelID == "" || messageID == "" {
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "❌ Invoice Generation Failed",
		Description: fmt.Sprintf("**Error:** %s", errorMsg),
		Color:       15548997, // Red
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Please try again or contact support",
		},
	}

	_, err := h.service.Discord.UpdateChannelMessage(channelID, messageID, "", []*discordgo.MessageEmbed{embed}, nil)
	if err != nil {
		l.Error(err, "failed to update discord with error")
	}
}

// updateDiscordWithNoContractors updates the Discord embed when no contractors found
func (h *handler) updateDiscordWithNoContractors(l logger.Logger, channelID, messageID, month string, batch int) {
	if channelID == "" || messageID == "" {
		return
	}

	displayMonth := formatMonthDisplay(month)

	embed := &discordgo.MessageEmbed{
		Title:       "ℹ️ No Contractors Found",
		Description: fmt.Sprintf("No active contractors found for **%s** with payday batch **%d**.", displayMonth, batch),
		Color:       5793266, // Blue
	}

	_, err := h.service.Discord.UpdateChannelMessage(channelID, messageID, "", []*discordgo.MessageEmbed{embed}, nil)
	if err != nil {
		l.Error(err, "failed to update discord with no contractors")
	}
}

// formatMonthDisplay converts "2026-01" to "January 2026"
func formatMonthDisplay(month string) string {
	parts := strings.Split(month, "-")
	if len(parts) != 2 {
		return month
	}

	monthNames := []string{
		"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}

	monthNum := 0
	if _, err := fmt.Sscanf(parts[1], "%d", &monthNum); err != nil {
		return month
	}
	if monthNum < 1 || monthNum > 12 {
		return month
	}

	return fmt.Sprintf("%s %s", monthNames[monthNum-1], parts[0])
}

// isValidMonthFormat validates that the month string is in YYYY-MM format
func isValidMonthFormat(month string) bool {
	matched, _ := regexp.MatchString(`^\d{4}-\d{2}$`, month)
	return matched
}

// MarkPaidResponse is the response for mark paid endpoint
type MarkPaidResponse struct {
	InvoiceNumber   string    `json:"invoice_number"`
	Source          string    `json:"source"`
	PaidAt          time.Time `json:"paid_at"`
	PostgresUpdated bool      `json:"postgres_updated"`
	NotionUpdated   bool      `json:"notion_updated"`
} // @name MarkPaidResponse

// MarkPaid godoc
// @Summary Mark invoice as paid by invoice number
// @Description Mark invoice as paid by searching both PostgreSQL and Notion, updating where found
// @id markInvoiceAsPaid
// @Tags Invoice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.MarkPaidRequest true "Mark paid request"
// @Success 200 {object} MarkPaidResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /invoices/mark-paid [post]
func (h *handler) MarkPaid(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "invoice",
		"method":  "MarkPaid",
	})

	l.Debug("handling mark invoice as paid request")

	// 1. Parse request
	var req request.MarkPaidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		l.Error(err, "invalid request body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	l.Debugf("received request: invoiceNumber=%s resendOnly=%v", req.InvoiceNumber, req.ResendOnly)

	// 2. Call controller
	result, err := h.controller.Invoice.MarkInvoiceAsPaidByNumber(req.InvoiceNumber, req.ResendOnly)
	if err != nil {
		l.Error(err, "failed to mark invoice as paid")
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, req, ""))
			return
		}
		if strings.Contains(err.Error(), "cannot mark as paid") {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, ""))
			return
		}
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	// 3. Build response
	response := MarkPaidResponse{
		InvoiceNumber:   result.InvoiceNumber,
		Source:          result.Source,
		PaidAt:          result.PaidAt,
		PostgresUpdated: result.PostgresUpdated,
		NotionUpdated:   result.NotionUpdated,
	}

	l.Infof("invoice marked as paid successfully: invoiceNumber=%s source=%s", result.InvoiceNumber, result.Source)
	c.JSON(http.StatusOK, view.CreateResponse(response, nil, nil, req, ""))
}

// GenerateSplits godoc
// @Summary Generate invoice splits by Legacy Number
// @Description Query Notion Client Invoices database and enqueue worker job to generate splits
// @id generateInvoiceSplits
// @Tags Invoice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.GenerateSplitsRequest true "Generate Splits Request"
// @Success 200 {object} view.GenerateSplitsResponse
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 404 {object} ErrorResponse "Invoice not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /invoices/generate-splits [post]
func (h *handler) GenerateSplits(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "invoice",
		"method":  "GenerateSplits",
	})

	l.Debug("handling generate invoice splits request")

	// 1. Parse and validate request
	var req request.GenerateSplitsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		l.Error(err, "invalid request body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	l.Debugf("received request: legacyNumber=%s", req.LegacyNumber)

	// 2. Call controller to generate splits
	resp, err := h.controller.Invoice.GenerateInvoiceSplitsByLegacyNumber(req.LegacyNumber)
	if err != nil {
		l.Error(err, "failed to generate invoice splits")

		// Handle not found error
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, req, ""))
			return
		}

		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	// 3. Return success response
	l.Infof("invoice splits generation job enqueued successfully: legacyNumber=%s pageID=%s", req.LegacyNumber, resp.InvoicePageID)
	c.JSON(http.StatusOK, view.CreateResponse(resp, nil, nil, req, ""))
}

// extractFileIDFromURL extracts the Google Drive file ID from a URL
// Expected format: https://drive.google.com/file/d/{fileID}/view
func extractFileIDFromURL(url string) string {
	// Format: https://drive.google.com/file/d/{fileID}/view
	prefix := "https://drive.google.com/file/d/"
	suffix := "/view"

	if !strings.HasPrefix(url, prefix) {
		return ""
	}

	url = strings.TrimPrefix(url, prefix)
	url = strings.TrimSuffix(url, suffix)

	return url
}
