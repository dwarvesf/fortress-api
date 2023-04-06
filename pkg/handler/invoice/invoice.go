package invoice

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
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
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	bcConst "github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils/authutils"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/dwarvesf/fortress-api/pkg/worker"
)

type handler struct {
	store   *store.Store
	service *service.Service
	worker  *worker.Worker
	logger  logger.Logger
	repo    store.DBRepo
	config  *config.Config
}

// New returns a handler
func New(store *store.Store, repo store.DBRepo, service *service.Service, worker *worker.Worker, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		store:   store,
		repo:    repo,
		service: service,
		worker:  worker,
		logger:  logger,
		config:  cfg,
	}
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
// @Summary Get the latest invoice by project id
// @Description Get the latest invoice by project id
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

	rs, err := view.ToInvoiceTemplateResponse(p, lastInvoice, *nextInvoiceNumber)
	if err != nil {
		l.Error(err, "failed to parse invoice template response")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](rs, nil, nil, nil, ""))
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

	senderID, err := model.UUIDFromString(userID)
	if err != nil {
		l.Error(err, "failed to parse sender id")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
	}

	req.SentByID = &senderID

	// check sender existence
	exists, err := h.store.Employee.IsExist(h.repo.DB(), senderID.String())
	if err != nil {
		l.Error(err, "failed to check sender existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	if !exists {
		l.Error(errs.ErrSenderNotFound, "sender not exist")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrSenderNotFound, req, ""))
		return
	}

	iv, err := req.ToInvoiceModel()
	if err != nil {
		l.Error(err, "invalid req")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	dueAt, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		l.Error(errs.ErrInvalidDueAt, "invalid invoice due date")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidDueAt, req, ""))
		return
	}
	iv.DueAt = &dueAt

	invoiceAt, err := time.Parse("2006-01-02", req.InvoiceDate)
	if err != nil {
		l.Error(errs.ErrInvalidPaidAt, "invalid invoice date")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidPaidAt, req, ""))
		return
	}
	iv.InvoicedAt = &invoiceAt

	// check bank account existence
	b, err := h.store.BankAccount.One(h.repo.DB(), req.BankID.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(errs.ErrBankAccountNotFound, "project not found")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, errs.ErrBankAccountNotFound, req, ""))
			return
		}

		l.Error(err, "failed to check bank account existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	iv.Bank = b

	// check project existence
	p, err := h.store.Project.One(h.repo.DB(), req.ProjectID.String(), true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(errs.ErrProjectNotFound, "project not found")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, req, ""))
			return
		}

		l.Error(err, "failed to check project existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}
	iv.Project = p

	nextInvoiceNumber, err := h.store.Invoice.GetNextInvoiceNumber(h.repo.DB(), now.Year(), p.Code)
	if err != nil {
		l.Error(err, "failed to get next invoice Number")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}
	iv.Number = *nextInvoiceNumber

	if err := h.generateInvoicePDF(l, iv); err != nil {
		l.Error(err, "failed to get next invoice Number")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	if req.IsDraft {
		iv.Status = model.InvoiceStatusDraft
	}

	errsCh := make(chan error)
	var amountGr = 1

	go func() {
		temp, rate, err := h.service.Wise.Convert(float64(iv.Total), iv.Bank.Currency.Name, "VND")
		if err != nil {
			errsCh <- err
			return
		}
		am := model.NewVietnamDong(int64(temp))
		iv.ConversionAmount = int64(am)
		iv.ConversionRate = rate

		invrs, err := h.store.Invoice.Save(h.repo.DB(), iv)
		if err != nil {
			l.Errorf(err, "failed to create invoice", "invoice", iv.Number)
			errsCh <- err
			return
		}
		iv.ID = invrs.ID

		if err := h.store.InvoiceNumberCaching.UpdateInvoiceCachingNumber(h.repo.DB(), time.Now(), iv.Project.Code); err != nil {
			l.Errorf(err, "failed to update invoice caching number", "project", iv.Project.Code)
			errsCh <- err
			return
		}
		errsCh <- nil
	}()

	if !req.IsDraft {
		amountGr += 2
		fn := strconv.FormatInt(rand.Int63(), 10) + "_" + iv.Number + ".pdf"

		invoiceFilePath := fmt.Sprintf("https://storage.googleapis.com/%s/invoices/%s", h.config.Google.GCSBucketName, fn)
		iv.InvoiceFileURL = invoiceFilePath

		go func() {
			err = h.service.GoogleDrive.UploadInvoicePDF(iv, "Sent")
			if err != nil {
				l.Errorf(err, "failed to upload invoice")
				errsCh <- err
				return
			}
			errsCh <- nil
		}()

		go func() {
			threadID, err := h.service.GoogleMail.SendInvoiceMail(iv)
			if err != nil {
				l.Errorf(err, "failed to send invoice mail")
				errsCh <- err
				return
			}
			iv.ThreadID = threadID

			_, err = h.store.Invoice.UpdateSelectedFieldsByID(h.repo.DB(), iv.ID.String(), *iv, "thread_id")
			if err != nil {
				l.Errorf(err, "failed to update invoice thread id", "thread_id", threadID)
				errsCh <- err
				return
			}

			attachmentSgID, err := h.service.Basecamp.Attachment.Create("application/pdf", fn, iv.InvoiceFileContent)
			if err != nil {
				l.Errorf(err, "failed to create Basecamp Attachment", "invoice", iv)
				errsCh <- err
				return
			}

			iv.TodoAttachment = fmt.Sprintf(`<bc-attachment sgid="%v" caption="My photo"></bc-attachment>`, attachmentSgID)

			bucketID, todoID, err := h.getInvoiceTodo(iv)
			if err != nil {
				l.Errorf(err, "failed to get invoice todo", "invoice", iv)
				errsCh <- err
				return
			}

			msg := fmt.Sprintf(`#Invoice %v has been sent

			Confirm Command: Paid @Giang #%v`, iv.Number, iv.Number)

			h.worker.Enqueue(bcModel.BasecampCommentMsg, h.service.Basecamp.BuildCommentMessage(bucketID, todoID, msg, ""))

			errsCh <- nil
		}()
	}

	var count int
	for e := range errsCh {
		if e != nil {
			close(errsCh)
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, e, req, ""))
			return
		}
		count++
		if count == amountGr {
			close(errsCh)
			c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
			return
		}
	}
}

func (h *handler) generateInvoicePDF(l logger.Logger, invoice *model.Invoice) error {
	pound := money.New(1, invoice.Project.BankAccount.Currency.Name)

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
		l.Errorf(err, "failed to create invoice pdf", "invoice", invoice)
		return err
	}

	invoice.InvoiceFileContent = pdfg.Buffer().Bytes()

	return nil
}

func (h *handler) getInvoiceTodo(iv *model.Invoice) (bucketID, todoID int, err error) {
	if iv.Project == nil {
		return 0, 0, fmt.Errorf(`missing project info`)
	}

	accountingID := consts.AccountingID
	accountingTodoID := consts.AccountingTodoID

	if h.config.Env != "prod" {
		accountingID = consts.PlaygroundID
		accountingTodoID = consts.PlaygroundTodoID
	}

	re := regexp.MustCompile(`Accounting \| ([A-Za-z]+) ([0-9]{4})`)

	todoLists, err := h.service.Basecamp.Todo.GetLists(accountingID, accountingTodoID)
	if err != nil {
		return 0, 0, err
	}

	var todoList *bcModel.TodoList
	var latestListDate time.Time

	for i := range todoLists {
		info := re.FindStringSubmatch(todoLists[i].Title)
		if len(info) == 3 {
			month, err := timeutil.GetMonthFromString(info[1])
			if err != nil {
				continue
			}
			year, _ := strconv.Atoi(info[2])
			listDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
			if listDate.After(latestListDate) {
				todoList = &todoLists[i]
				latestListDate = listDate
			}
		}
	}

	if todoList == nil {
		month := iv.Month + 1
		if month > 12 {
			month = 1
		}
		todoList, err = h.service.Basecamp.Todo.CreateList(
			accountingID,
			accountingTodoID,
			bcModel.TodoList{Name: fmt.Sprintf(
				`Accounting | %v %v`, time.Month(month).String(),
				iv.Year)},
		)
		if err != nil {
			return 0, 0, err
		}
	}

	todoGroup, err := h.service.Basecamp.Todo.FirstOrCreateGroup(
		accountingID,
		todoList.ID,
		`In`)
	if err != nil {
		return 0, 0, err
	}

	todo, err := h.service.Basecamp.Todo.FirstOrCreateInvoiceTodo(
		accountingID,
		todoGroup.ID,
		iv)
	if err != nil {
		return 0, 0, err
	}

	return accountingID, todo.ID, nil
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

	switch req.Status {
	case model.InvoiceStatusError:
		_, err = h.markInvoiceAsError(l, invoice)
	case model.InvoiceStatusPaid:
		_, err = h.markInvoiceAsPaid(l, invoice, req.SendThankYouEmail)
	default:
		_, err = h.store.Invoice.UpdateSelectedFieldsByID(h.repo.DB(), invoice.ID.String(), *invoice, "status")
	}
	if err != nil {
		l.Error(err, "failed to update invoice")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

func (h *handler) markInvoiceAsError(l logger.Logger, invoice *model.Invoice) (*model.Invoice, error) {
	tx, done := h.repo.NewTransaction()
	invoice.Status = model.InvoiceStatusError
	iv, err := h.store.Invoice.UpdateSelectedFieldsByID(tx.DB(), invoice.ID.String(), *invoice, "status")
	if err != nil {
		l.Errorf(err, "failed to update invoice status to error")
		return nil, done(err)
	}

	err = h.store.InvoiceNumberCaching.UnCountErrorInvoice(tx.DB(), *invoice.InvoicedAt)
	if err != nil {
		l.Errorf(err, "failed to un-count error invoice")
		return nil, done(err)
	}

	if err := h.markInvoiceTodoAsError(invoice); err != nil {
		return nil, done(err)
	}

	if err := h.service.GoogleDrive.MoveInvoicePDF(invoice, "Sent", "Error"); err != nil {
		l.Errorf(err, "failed to upload invoice pdf to google drive")
		return nil, done(err)
	}

	return iv, done(nil)
}

func (h *handler) markInvoiceTodoAsError(invoice *model.Invoice) error {
	if invoice.Project == nil {
		return fmt.Errorf(`missing project info`)
	}

	bucketID, todoID, err := h.getInvoiceTodo(invoice)
	if err != nil {
		return err
	}

	h.worker.Enqueue(bcModel.BasecampCommentMsg, h.service.Basecamp.BuildCommentMessage(bucketID, todoID, "Invoice has been mark as error", "failed"))

	return h.service.Basecamp.Recording.Archive(bucketID, todoID)
}

type processPaidInvoiceRequest struct {
	Invoice          *model.Invoice
	InvoiceTodoID    int
	InvoiceBucketID  int
	SentThankYouMail bool
}

func (h *handler) markInvoiceAsPaid(l logger.Logger, invoice *model.Invoice, sendThankYouEmail bool) (*model.Invoice, error) {
	if invoice.Status != model.InvoiceStatusSent && invoice.Status != model.InvoiceStatusOverdue {
		err := fmt.Errorf(`unable to update invoice status, invoice have status %v`, invoice.Status)
		l.Errorf(err, "failed to update invoice", "invoiceID", invoice.ID.String())
		return nil, err
	}
	invoice.Status = model.InvoiceStatusPaid

	bucketID, todoID, err := h.getInvoiceTodo(invoice)
	if err != nil {
		l.Errorf(err, "failed to get invoice todo", "invoiceID", invoice.ID.String())
		return nil, err
	}

	err = h.service.Basecamp.Todo.Complete(bucketID, todoID)
	if err != nil {
		l.Errorf(err, "failed to complete invoice todo", "invoiceID", invoice.ID.String())
	}

	h.processPaidInvoice(l, &processPaidInvoiceRequest{
		Invoice:          invoice,
		InvoiceTodoID:    todoID,
		InvoiceBucketID:  bucketID,
		SentThankYouMail: sendThankYouEmail,
	})

	return invoice, nil
}

func (h *handler) processPaidInvoice(l logger.Logger, req *processPaidInvoiceRequest) {
	wg := &sync.WaitGroup{}
	wg.Add(3)

	go func() {
		_ = h.processPaidInvoiceData(l, wg, req)
	}()

	go h.sendThankYouEmail(l, wg, req)
	go h.movePaidInvoiceGDrive(l, wg, req)

	wg.Wait()
}

func (h *handler) processPaidInvoiceData(l logger.Logger, wg *sync.WaitGroup, req *processPaidInvoiceRequest) error {
	// Start Transaction
	tx, done := h.repo.NewTransaction()

	msg := bcConst.CommentUpdateInvoiceFailed
	msgType := bcModel.CommentMsgTypeFailed
	defer func() {
		wg.Done()
		h.worker.Enqueue(bcModel.BasecampCommentMsg, h.service.Basecamp.BuildCommentMessage(req.InvoiceBucketID, req.InvoiceTodoID, msg, msgType))
	}()

	now := time.Now()
	req.Invoice.PaidAt = &now
	_, err := h.store.Invoice.UpdateSelectedFieldsByID(tx.DB(), req.Invoice.ID.String(), *req.Invoice, "status", "paid_at")
	if err != nil {
		l.Errorf(err, "failed to update invoice status to paid", "invoice", req.Invoice)
		return done(err)
	}

	_, err = h.storeCommission(tx.DB(), l, req.Invoice)
	if err != nil {
		l.Errorf(err, "failed to store invoice commission", "invoice", req.Invoice)
		return done(err)
	}

	m := model.AccountingMetadata{
		Source: "invoice",
		ID:     req.Invoice.ID.String(),
	}

	bonusBytes, err := json.Marshal(&m)
	if err != nil {
		l.Errorf(err, "failed to process invoice accounting metadata", "invoiceNumber", req.Invoice.Number)
		return done(err)
	}

	projectOrg := ""
	if req.Invoice.Project.Organization != nil {
		projectOrg = req.Invoice.Project.Organization.Name
	}

	currencyName := "VND"
	currencyID := model.UUID{}
	if req.Invoice.Project.BankAccount.Currency != nil {
		currencyName = req.Invoice.Project.BankAccount.Currency.Name
		currencyID = req.Invoice.Project.BankAccount.Currency.ID
	}

	accountingTxn := &model.AccountingTransaction{
		Name:             req.Invoice.Number,
		Amount:           float64(req.Invoice.Total),
		Date:             &now,
		ConversionAmount: model.VietnamDong(req.Invoice.ConversionAmount),
		Organization:     projectOrg,
		Category:         model.AccountingIn,
		Type:             model.AccountingIncome,
		Currency:         currencyName,
		CurrencyID:       &currencyID,
		ConversionRate:   req.Invoice.ConversionRate,
		Metadata:         bonusBytes,
	}

	err = h.store.Accounting.CreateTransaction(tx.DB(), accountingTxn)
	if err != nil {
		l.Errorf(err, "failed to create accounting transaction", "Accounting Transaction", accountingTxn)
		return done(err)
	}

	msg = bcConst.CommentUpdateInvoiceSuccessfully
	msgType = bcModel.CommentMsgTypeCompleted

	return done(nil)
}

func (h *handler) sendThankYouEmail(l logger.Logger, wg *sync.WaitGroup, req *processPaidInvoiceRequest) {
	msg := h.service.Basecamp.BuildCommentMessage(req.InvoiceBucketID, req.InvoiceTodoID, bcConst.CommentThankYouEmailSent, bcModel.CommentMsgTypeCompleted)

	defer func() {
		h.worker.Enqueue(bcModel.BasecampCommentMsg, msg)
		wg.Done()
	}()

	err := h.service.GoogleMail.SendInvoiceThankYouMail(req.Invoice)
	if err != nil {
		l.Errorf(err, "failed to send invoice thank you mail", "invoice", req.Invoice)
		msg = h.service.Basecamp.BuildCommentMessage(req.InvoiceBucketID, req.InvoiceTodoID, bcConst.CommentThankYouEmailFailed, bcModel.CommentMsgTypeFailed)
		return
	}
}
