package invoice

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	invoiceCtrl "github.com/dwarvesf/fortress-api/pkg/controller/invoice"
	"github.com/dwarvesf/fortress-api/pkg/handler/invoice/errs"
	"github.com/dwarvesf/fortress-api/pkg/handler/invoice/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
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
		Pagination: pagination,
		ProjectIDs: query.ProjectID,
		Statuses:   query.Status,
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

	iv, err := req.ToInvoiceModel(userID)
	if err != nil {
		l.Error(err, "failed to parse request to invoice model")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

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

	if err := req.Validate(); err != nil {
		l.Error(err, "invalid request")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

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

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}
