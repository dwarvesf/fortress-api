package invoice

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/Rhymond/go-money"
	toPdf "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler/invoice/errs"
	"github.com/dwarvesf/fortress-api/pkg/handler/invoice/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils/authutils"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
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

// UpdateStatus godoc
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
		l.Error(err, "invalid req")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	// check invoice existence
	invoice, err := h.store.Invoice.One(h.repo.DB(), invoiceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(errs.ErrInvoiceNotFound, "invoice not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrInvoiceNotFound, req, ""))
			return
		}

		l.Error(err, "failed to get invoice")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	if invoice.Status == req.Status {
		l.Error(errs.ErrInvoiceStatusAlready, "invoice status already")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvoiceStatusAlready, req, ""))
		return
	}

	// Start Transaction
	tx, done := h.repo.NewTransaction()

	switch req.Status {
	case model.InvoiceStatusError:
		_, err = h.markInvoiceAsError(tx.DB(), l, *invoice)
	case model.InvoiceStatusPaid:
		_, err = h.markInvoiceAsPaid(tx.DB(), l, *invoice, req.SendThankYouEmail)
	default:
		_, err = h.store.Invoice.UpdateSelectedFieldsByID(tx.DB(), invoice.ID.String(), *invoice, "status")
	}
	if err != nil {
		l.Error(err, "failed to update invoice")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), req, ""))
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

// GetTemplate godoc
// @Summary Get latest invoice by project id
// @Description Get latest invoice by project id
// @Tags Invoice
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param projectID query string true "projectID"
// @Success 200 {object} view.InvoiceTemplateResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
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

	// check project existence
	p, err := h.store.Project.One(h.repo.DB(), input.ProjectID, true)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(errs.ErrProjectNotFound, "project not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, nil, ""))
		return
	}

	if err != nil {
		l.Error(err, "failed to get project")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	nextInvoiceNumber, err := h.store.Invoice.GetNextInvoiceNumber(h.repo.DB(), now.Year(), p.Code)
	if err != nil {
		l.Error(err, "failed to get next invoice Number")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	lastInvoice, err := h.store.Invoice.GetLatestInvoiceByProject(h.repo.DB(), input.ProjectID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(err, "failed to get latest invoice")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		lastInvoice = nil
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToInvoiceTemplateResponse(p, lastInvoice, *nextInvoiceNumber), nil, nil, nil, ""))
}

// Send godoc
// @Summary Create new invoice and send to client
// @Description Create new invoice and send to clientm
// @Tags Invoice
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param Body body request.SendInvoiceRequest true "body"
// @Success 200 {object} view.MessageResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /invoices/send [post]
func (h *handler) Send(c *gin.Context) {
	now := time.Now()
	userID, err := authutils.GetUserIDFromContext(c, h.config)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	var input request.SendInvoiceRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "invoice",
		"method":  "Send",
		"input":   input,
	})

	senderID, err := model.UUIDFromString(userID)
	if err != nil {
		l.Error(err, "failed to parse sender id")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
	}

	input.SentByID = &senderID

	// check sender existence
	exists, err := h.store.Employee.IsExist(h.repo.DB(), senderID.String())
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

	iv, err := input.ToInvoiceModel()
	if err != nil {
		l.Error(err, "invalid input")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	dueAt, err := time.Parse("2006-01-02", input.DueDate)
	if err != nil {
		l.Error(errs.ErrInvalidDueAt, "invalid invoice due date")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidDueAt, input, ""))
		return
	}
	iv.DueAt = &dueAt

	invoiceAt, err := time.Parse("2006-01-02", input.InvoiceDate)
	if err != nil {
		l.Error(errs.ErrInvalidPaidAt, "invalid invoice date")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidPaidAt, input, ""))
		return
	}
	iv.InvoicedAt = &invoiceAt

	// check bank account existence
	b, err := h.store.BankAccount.One(h.repo.DB(), input.BankID.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(errs.ErrBankAccountNotFound, "project not found")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, errs.ErrBankAccountNotFound, input, ""))
			return
		}

		l.Error(err, "failed to check bank account existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	iv.Bank = b

	// check project existence
	p, err := h.store.Project.One(h.repo.DB(), input.ProjectID.String(), true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(errs.ErrProjectNotFound, "project not found")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, input, ""))
			return
		}

		l.Error(err, "failed to check project existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	iv.Project = p

	nextInvoiceNumber, err := h.store.Invoice.GetNextInvoiceNumber(h.repo.DB(), now.Year(), p.Code)
	if err != nil {
		l.Error(err, "failed to get next invoice Number")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}
	iv.Number = *nextInvoiceNumber

	if err := h.generateInvoicePDF(l, iv); err != nil {
		l.Error(err, "failed to get next invoice Number")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if input.IsDraft {
		iv.Status = model.InvoiceStatusDraft
	}

	errsCh := make(chan error)
	var amountGr = 1

	go func() {
		temp, rate, err := h.service.Wise.Convert(float64(iv.Total), iv.Project.BankAccount.Currency.Name, "VND")
		if err != nil {
			errsCh <- err
			return
		}
		am := model.NewVietnamDong(int64(temp))
		iv.ConversionAmount = int64(am)
		iv.ConversionRate = rate
		// store invoice to db

		_, err = h.store.Invoice.Save(h.repo.DB(), iv)
		if err != nil {
			l.Errorf(err, "failed to create invoice", "invoice", iv.Number)
			errsCh <- err
			return
		}

		if err := h.store.InvoiceNumberCaching.UpdateInvoiceCachingNumber(h.repo.DB(), time.Now(), iv.Project.Code); err != nil {
			l.Errorf(err, "failed to update invoice caching number", "project", iv.Project.Code)
			errsCh <- err
			return
		}
		errsCh <- nil
	}()

	if !input.IsDraft {
		amountGr += 2
		fn := strconv.FormatInt(rand.Int63(), 10) + "_" + iv.Number + ".pdf"
		iv.InvoiceFileURL = h.config.Google.GDSBucketURL + fn

		go func() {
			errsCh <- h.service.GoogleDrive.UploadInvoicePDF(iv, "Sent")
		}()

		go func() {
			threadID, err := h.service.GoogleMail.SendInvoiceMail(iv)
			if err != nil {
				l.Errorf(err, "failed to send invoice mail")
				errsCh <- err
				return
			}

			_, err = h.store.Invoice.UpdateSelectedFieldsByID(h.repo.DB(), iv.ID.String(), *iv, "thread_id")
			if err != nil {
				l.Errorf(err, "failed to update invoice thread id", "thread_id", threadID)
				errsCh <- err
				return
			}
		}()
	}

	var count int
	for e := range errsCh {
		if e != nil {
			close(errsCh)
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, e, input, ""))
			return
		}
		count++
		if count == amountGr {
			close(errsCh)
			c.JSON(http.StatusOK, view.CreateResponse[any](view.MessageResponse{Message: "ok"}, nil, nil, nil, ""))
			return
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.MessageResponse{Message: "ok"}, nil, nil, nil, ""))
}

func (h *handler) generateInvoicePDF(l logger.Logger, invoice *model.Invoice) error {
	pound := money.New(1, invoice.Bank.Currency.Name)

	items, err := model.GetInfoItems(invoice.LineItems)
	if err != nil {
		l.Errorf(err, "failed to get info items", "invoice-lineItems", invoice.LineItems)
		return err
	}

	companyInfo, err := invoice.Project.GetCompanyContactInfo()
	if err != nil {
		l.Errorf(err, "failed to get company contact info", "project", invoice.Project)
		return err
	}

	var haveDiscountColumn bool
	for i := range items {
		if items[i].Discount != 0 {
			haveDiscountColumn = true
		}
	}

	haveRouting := invoice.Bank.RoutingNumber != ""
	haveSwiftCode := invoice.Bank.SwiftCode != ""
	haveUKSortCode := invoice.Bank.UKSortCode != ""

	data := &struct {
		Path               string
		Invoice            *model.Invoice
		HaveRouting        bool
		HaveUKSortCode     bool
		HaveSWIFTCode      bool
		CompanyContactInfo *model.CompanyContactInfo
		InvoiceItem        []model.InvoiceItem
	}{
		Path:               h.config.Invoice.TemplatePath,
		Invoice:            invoice,
		HaveRouting:        haveRouting,
		HaveUKSortCode:     haveUKSortCode,
		HaveSWIFTCode:      haveSwiftCode,
		CompanyContactInfo: companyInfo,
		InvoiceItem:        items,
	}

	funcMap := template.FuncMap{
		"toString": func(month int) string {
			return time.Month(month).String()
		},
		"formatDate": func(t *time.Time) string {
			return timeutil.FormatDatetime(*t)
		},
		"lastDayOfMonth": func() string {
			return timeutil.
				FormatDatetime(timeutil.LastDayOfMonth(invoice.Month, invoice.Year))
		},
		"formatMoney": func(money int64) string {
			var result string
			formatted := pound.
				Multiply(money * int64(math.Pow(10, float64(pound.Currency().Fraction)))).
				Display()
			result = formatted
			parts := strings.Split(formatted, ".00")
			if len(parts) > 1 {
				result = parts[0]
			}
			return result
		},
		"haveDescription": func(description string) bool {
			return description != ""
		},
		"haveNote": func(note string) bool {
			return note != ""
		},
		"haveDiscountColumn": func() bool {
			return haveDiscountColumn
		},
		"float": func(n float64) string {
			return fmt.Sprintf("%.2f", n)
		},
	}

	if h.config.Env == "local" {
		data.Path = os.Getenv("GOPATH") + "/src/github.com/dwarvesf/fortress-api/pkg/templates"
	}

	tmpl, err := template.New("invoicePDF").Funcs(funcMap).ParseFiles(filepath.Join(data.Path, "invoice.html"))
	if err != nil {
		l.Errorf(err, "failed to parse template", "path", data.Path, "filename", "invoice.html")
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.Funcs(funcMap).ExecuteTemplate(&buf, "invoice.html", data); err != nil {
		l.Errorf(err, "failed to execute template", "data", data, "path", data.Path, "filename", "invoice.html")
		return err
	}

	pdfg, err := toPdf.NewPDFGenerator()
	if err != nil {
		l.Errorf(err, "failed to create pdf generator")
		return err
	}

	t := toPdf.NewPageReader(&buf)
	t.Zoom.Set(1.45)
	t.EnableLocalFileAccess.Set(true)
	pdfg.AddPage(t)
	pdfg.Dpi.Set(600)
	pdfg.PageSize.Set("A4")

	if err := pdfg.Create(); err != nil {
		l.Errorf(err, "failed to create pdf", "invoice", invoice)
		return err
	}

	invoice.InvoiceFileContent = pdfg.Buffer().Bytes()

	return nil
}

func (h *handler) markInvoiceAsError(db *gorm.DB, l logger.Logger, invoice model.Invoice) (*model.Invoice, error) {
	invoice.Status = model.InvoiceStatusError
	iv, err := h.store.Invoice.UpdateSelectedFieldsByID(db, invoice.ID.String(), invoice, "status")
	if err != nil {
		l.Errorf(err, "failed to update invoice status to error")
		return nil, err
	}

	err = h.store.InvoiceNumberCaching.UnCountErrorInvoice(db, *iv.InvoicedAt)
	if err != nil {
		l.Errorf(err, "failed to un-count error invoice")
		return nil, err
	}

	//TODO: mark Invoice as Error in Basecamp
	//if err := markInvoiceTodoAsError(cfg, i); err != nil {
	//	return nil, err
	//}

	if err := h.service.GoogleDrive.MoveInvoicePDF(iv, "Sent", "Error"); err != nil {
		l.Errorf(err, "failed to upload invoice pdf to google drive")
		return nil, err
	}

	return iv, nil
}

type processPaidInvoiceRequest struct {
	Invoice          *model.Invoice
	InvoiceTodoID    int
	InvoiceBucketID  int
	SentThankYouMail bool
}

func (h *handler) markInvoiceAsPaid(db *gorm.DB, l logger.Logger, invoice model.Invoice, sendThankYouEmail bool) (*model.Invoice, error) {
	if invoice.Status != model.InvoiceStatusSent && invoice.Status != model.InvoiceStatusOverdue {
		err := fmt.Errorf(`unable to update invoice status, invoice have status %v`, invoice.Status)
		l.Errorf(err, "failed to update invoice", "invoiceID", invoice.ID.String())
		return nil, err
	}
	invoice.Status = model.InvoiceStatusPaid

	//TODO: mark Invoice as Paid in Basecamp
	//bucketID, todoID, err := handler.GetInvoiceTodo(cfg, i)
	//if err != nil {
	//	log.E("Get invoice todo failed", err, log.M{"invoice": i})
	//	return nil, err
	//}
	//
	//err = cfg.Service().Basecamp.Todo.Complete(bucketID, todoID)
	//if err != nil {
	//	log.E("complete invoice todo failed", err, log.M{"bucketID": bucketID, "todoID": todoID})
	//}

	h.processPaidInvoice(db, l, &processPaidInvoiceRequest{
		Invoice:          &invoice,
		InvoiceTodoID:    0,
		InvoiceBucketID:  0,
		SentThankYouMail: sendThankYouEmail,
	})

	return &invoice, nil

}

func (h *handler) processPaidInvoice(db *gorm.DB, l logger.Logger, req *processPaidInvoiceRequest) {
	wg := &sync.WaitGroup{}
	wg.Add(3)
	go h.processPaidInvoiceData(db, l, wg, req)
	go h.sendThankYouEmail(db, l, wg, req)
	go h.movePaidInvoiceGDrive(db, l, wg, req)
	wg.Wait()
}

func (h *handler) processPaidInvoiceData(db *gorm.DB, l logger.Logger, wg *sync.WaitGroup, req *processPaidInvoiceRequest) {
	defer wg.Done()
	//TODO: mark Invoice as Paid in Basecamp
	//msg := BuildFailedComment(domain.CommentUpdateInvoiceFailed)
	//defer func() {
	//	CommentResult(req.InvoiceBucketID, req.InvoiceTodoID, msg)
	//}()

	now := time.Now()
	req.Invoice.PaidAt = &now
	_, err := h.store.Invoice.UpdateSelectedFieldsByID(db, req.Invoice.ID.String(), *req.Invoice, "status", "paid_at")
	if err != nil {
		l.Errorf(err, "failed to update invoice status to paid", "invoice", req.Invoice)
		return
	}
}

func (h *handler) sendThankYouEmail(db *gorm.DB, l logger.Logger, wg *sync.WaitGroup, req *processPaidInvoiceRequest) {
	defer wg.Done()
}

func (h *handler) movePaidInvoiceGDrive(db *gorm.DB, l logger.Logger, wg *sync.WaitGroup, req *processPaidInvoiceRequest) {
	defer wg.Done()
}
