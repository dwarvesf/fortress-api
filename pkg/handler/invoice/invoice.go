package invoice

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler/invoice/errs"
	"github.com/dwarvesf/fortress-api/pkg/handler/invoice/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
	repo    store.DBRepo
	config  *config.Config
}

// New returns a handler
func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		store:   store,
		repo:    repo,
		service: service,
		logger:  logger,
		config:  cfg,
	}
}

// Create godoc
// @Summary Create new invoice
// @Description Create new invoice
// @Tags Invoice
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param Body body request.CreateInput true "body"
// @Success 200 {object} view.CreateInvoiceResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /invoices/ [post]
func (h *handler) Create(c *gin.Context) {
	senderID, err := utils.GetUserIDFromContext(c, h.config)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	var input request.CreateInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "invoice",
		"method":  "Create",
		"input":   input,
	})

	if err := input.Validate(); err != nil {
		l.Error(err, "invalid input")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	// check sender existence
	exists, err := h.store.Employee.IsExist(h.repo.DB(), senderID)
	if err != nil {
		l.Error(err, "failed to check sender existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if !exists {
		l.Error(errs.ErrSenderNotFound, "sender not exist")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrSenderNotFound, input, ""))
		return
	}

	senderUUID := model.MustGetUUIDFromString(senderID)

	// check bank account existence
	exists, err = h.store.BankAccount.IsExist(h.repo.DB(), input.BankID.String())
	if err != nil {
		l.Error(err, "failed to check bank account existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if !exists {
		l.Error(errs.ErrBankAccountNotFound, "bank account not exist")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrBankAccountNotFound, input, ""))
		return
	}

	invoice := &model.Invoice{
		Number:           input.Number,
		Status:           input.Status,
		Email:            input.Email,
		Description:      input.Description,
		Note:             input.Note,
		SubTotal:         input.SubTotal,
		Tax:              input.Tax,
		Discount:         input.Discount,
		Total:            input.Total,
		ConversionAmount: input.ConversionAmount,
		Month:            input.Month,
		Year:             input.Year,
		SentBy:           &senderUUID,
		ThreadID:         input.ThreadID,
		ConversionRate:   input.ConversionRate,
		BankID:           input.BankID,
	}

	if !input.ProjectID.IsZero() {
		// check project existence
		exists, err = h.store.Project.IsExist(h.repo.DB(), input.ProjectID.String())
		if err != nil {
			l.Error(err, "failed to check project existence")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
			return
		}

		if !exists {
			l.Error(errs.ErrProjectNotFound, "project not exist")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, input, ""))
			return
		}

		invoice.ProjectID = input.ProjectID
	}

	if strings.TrimSpace(input.DueAt) != "" {
		dueAt, err := time.Parse("2006-01-02", input.DueAt)
		if err != nil {
			l.Error(errs.ErrInvalidDueAt, "invalid left date")
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidDueAt, input, ""))
			return
		}
		invoice.DueAt = &dueAt
	}

	if strings.TrimSpace(input.PaidAt) != "" {
		paidAt, err := time.Parse("2006-01-02", input.PaidAt)
		if err != nil {
			l.Error(errs.ErrInvalidPaidAt, "invalid left date")
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidPaidAt, input, ""))
			return
		}
		invoice.PaidAt = &paidAt
	}

	if strings.TrimSpace(input.ScheduledDate) != "" {
		scheduledDate, err := time.Parse("2006-01-02", input.ScheduledDate)
		if err != nil {
			l.Error(errs.ErrInvalidScheduledDate, "invalid left date")
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidScheduledDate, input, ""))
			return
		}
		invoice.ScheduledDate = &scheduledDate
	}

	res, err := h.store.Invoice.Create(h.repo.DB(), invoice)
	if err != nil {
		l.Error(err, "failed to create invoice")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](res, nil, nil, nil, ""))
}

// Update godoc
// @Summary Update status for invoice
// @Description Update status for invoice
// @Tags Invoice
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.MessageResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /invoices/{id}/status [get]
func (h *handler) Update(c *gin.Context) {
	invoiceID := c.Param("id")
	if invoiceID == "" || !model.IsUUIDFromString(invoiceID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidInvoiceID, nil, ""))
		return
	}

	var input request.UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "invoice",
		"method":  "Update",
		"input":   input,
	})

	if err := input.Validate(); err != nil {
		l.Error(err, "invalid input")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	// check invoice existence
	invoice, err := h.store.Invoice.One(h.repo.DB(), invoiceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(errs.ErrInvoiceNotFound, "invoice not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrInvoiceNotFound, input, ""))
			return
		}

		l.Error(err, "failed to get invoice")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	invoice.Status = input.Status

	invoice, err = h.store.Invoice.UpdateSelectedFieldsByID(h.repo.DB(), invoice.ID.String(), *invoice, "status")
	if err != nil {
		l.Error(err, "failed to update invoice")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

// GetLatestInvoice godoc
// @Summary Get latest invoice by project id
// @Description Get latest invoice by project id
// @Tags Invoice
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param projectID query string true "projectID"
// @Success 200 {object} view.GetLatestInvoiceResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /invoices/latest [get]
func (h *handler) GetLatestInvoice(c *gin.Context) {
	var input request.GetLatestInvoiceInput
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
		"method":  "GetLatestInvoice",
		"input":   input,
	})

	// check project existence
	exists, err := h.store.Project.IsExist(h.repo.DB(), input.ProjectID)
	if err != nil {
		l.Error(err, "failed to check project existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if !exists {
		l.Error(errs.ErrProjectNotFound, "project not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, input, ""))
		return
	}

	invoice, err := h.store.Invoice.GetLatestInvoiceByProject(h.repo.DB(), input.ProjectID)
	if err != nil {
		l.Error(err, "failed to get latest invoice")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](invoice, nil, nil, nil, ""))
}
