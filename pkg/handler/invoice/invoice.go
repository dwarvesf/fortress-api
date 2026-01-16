package invoice

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	invoiceCtrl "github.com/dwarvesf/fortress-api/pkg/controller/invoice"
	"github.com/dwarvesf/fortress-api/pkg/handler/invoice/errs"
	"github.com/dwarvesf/fortress-api/pkg/handler/invoice/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
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

	pagination := query.StandardizeInput()

	if err := query.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	invoices, total, err := h.controller.Invoice.List(invoiceCtrl.GetListInvoiceInput{
		Pagination:    pagination,
		ProjectIDs:    query.ProjectID,
		Statuses:      query.Status,
		InvoiceNumber: query.InvoiceNumber,
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

	l.Debug(fmt.Sprintf("received request: contractor=%s month=%s skipUpload=%v", req.Contractor, req.Month, req.SkipUpload))

	// 2. Validate month format (YYYY-MM) - only if month is provided
	if req.Month != "" && !isValidMonthFormat(req.Month) {
		l.Error(errs.ErrInvalidMonthFormat, fmt.Sprintf("month validation failed: month=%s", req.Month))
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidMonthFormat, req, ""))
		return
	}

	// 3. Generate invoice data
	l.Debug("calling controller to generate contractor invoice data")
	invoiceData, err := h.controller.Invoice.GenerateContractorInvoice(c.Request.Context(), req.Contractor, req.Month)
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

	if req.SkipUpload {
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
	}

	// 5.5 Create Contractor Payables record in Notion
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
		PDFBytes:         pdfBytes,     // Upload PDF to Notion
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
	c.JSON(http.StatusOK, view.CreateResponse(response, nil, nil, req, ""))
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

	l.Debugf("received request: invoiceNumber=%s", req.InvoiceNumber)

	// 2. Call controller
	result, err := h.controller.Invoice.MarkInvoiceAsPaidByNumber(req.InvoiceNumber)
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
