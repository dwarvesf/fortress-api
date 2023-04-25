package invoice

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/Rhymond/go-money"
	toPdf "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
)

// InvoiceItem invoice item
type InvoiceItem struct {
	Quantity    float64 `json:"quantity"`
	UnitCost    int64   `json:"unitCost"`
	Discount    int64   `json:"discount"`
	Cost        int64   `json:"cost"`
	Description string  `json:"description"`
	IsExternal  bool    `json:"isExternal"`
}

type SendInvoiceInput struct {
	IsDraft     bool          `json:"isDraft"`
	ProjectID   model.UUID    `json:"projectID" binding:"required"`
	BankID      model.UUID    `json:"bankID" binding:"required"`
	Description string        `json:"description"`
	Note        string        `json:"note"`
	CC          []string      `json:"cc"`
	LineItems   []InvoiceItem `json:"lineItems"`
	Email       string        `json:"email" binding:"required,email"`
	Total       int           `json:"total" binding:"gte=0"`
	Discount    int           `json:"discount" binding:"gte=0"`
	Tax         int           `json:"tax" binding:"gte=0"`
	SubTotal    int           `json:"subtotal" binding:"gte=0"`
	InvoiceDate string        `json:"invoiceDate" binding:"required"`
	DueDate     string        `json:"dueDate" binding:"required"`
	Month       int           `json:"invoiceMonth" binding:"gte=0,lte=11"`
	Year        int           `json:"invoiceYear" binding:"gte=0"`
	SentByID    *model.UUID
	Number      string
}

func (c *controller) Send(iv *model.Invoice) (*model.Invoice, error) {
	now := time.Now()

	l := c.logger.Fields(logger.Fields{
		"controller": "invoice",
		"method":     "Send",
	})

	// check sender existence
	exists, err := c.store.Employee.IsExist(c.repo.DB(), iv.SentBy.String())
	if err != nil {
		l.Error(err, "failed to check sender existence")
		return nil, err
	}

	if !exists {
		l.Error(ErrSenderNotFound, "sender not exist")
		return nil, err
	}

	// check bank account existence
	b, err := c.store.BankAccount.One(c.repo.DB(), iv.BankID.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(ErrBankAccountNotFound, "project not found")
			return nil, err
		}

		l.Error(err, "failed to check bank account existence")
		return nil, err
	}

	iv.Bank = b

	// check project existence
	p, err := c.store.Project.One(c.repo.DB(), iv.ProjectID.String(), true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(ErrProjectNotFound, "project not found")
			return nil, err
		}

		l.Error(err, "failed to check project existence")
		return nil, err
	}
	iv.Project = p

	nextInvoiceNumber, err := c.store.Invoice.GetNextInvoiceNumber(c.repo.DB(), now.Year(), p.Code)
	if err != nil {
		l.Error(err, "failed to get next invoice Number")
		return nil, err
	}
	iv.Number = *nextInvoiceNumber

	if err := c.generateInvoicePDF(l, iv); err != nil {
		l.Error(err, "failed to generate Invoice PDF")
		return nil, err
	}

	temp, rate, err := c.service.Wise.Convert(float64(iv.Total), iv.Bank.Currency.Name, "VND")
	if err != nil {
		l.Error(err, "failed to convert currency")
		return nil, err
	}
	am := model.NewVietnamDong(int64(temp))
	iv.ConversionAmount = int64(am)
	iv.ConversionRate = rate

	savedInvoice, err := c.store.Invoice.Save(c.repo.DB(), iv)
	if err != nil {
		l.Errorf(err, "failed to create invoice", "invoice", iv.Number)
		return nil, err
	}
	iv.ID = savedInvoice.ID

	if err := c.store.InvoiceNumberCaching.UpdateInvoiceCachingNumber(c.repo.DB(), time.Now(), iv.Project.Code); err != nil {
		l.Errorf(err, "failed to update invoice caching number", "project", iv.Project.Code)
		return nil, err
	}

	errsCh := make(chan error)
	var amountGr = 0
	if iv.Status != model.InvoiceStatusDraft {
		amountGr += 2
		fn := strconv.FormatInt(rand.Int63(), 10) + "_" + iv.Number + ".pdf"

		invoiceFilePath := fmt.Sprintf("https://storage.googleapis.com/%s/invoices/%s", c.config.Google.GCSBucketName, fn)
		iv.InvoiceFileURL = invoiceFilePath

		go func() {
			err = c.service.GoogleDrive.UploadInvoicePDF(iv, "Sent")
			if err != nil {
				l.Errorf(err, "failed to upload invoice")
				errsCh <- err
				return
			}

			errsCh <- nil
		}()

		go func() {
			threadID, err := c.service.GoogleMail.SendInvoiceMail(iv)
			if err != nil {
				l.Errorf(err, "failed to send invoice mail")
				errsCh <- err
				return
			}

			iv.ThreadID = threadID
			_, err = c.store.Invoice.UpdateSelectedFieldsByID(c.repo.DB(), iv.ID.String(), *iv, "thread_id")
			if err != nil {
				l.Errorf(err, "failed to update invoice thread id", "thread_id", threadID)
				errsCh <- err
				return
			}

			attachmentSgID, err := c.service.Basecamp.Attachment.Create("application/pdf", fn, iv.InvoiceFileContent)
			if err != nil {
				l.Errorf(err, "failed to create Basecamp Attachment", "invoice", iv)
				errsCh <- err
				return
			}

			iv.TodoAttachment = fmt.Sprintf(`<bc-attachment sgid="%v" caption="My photo"></bc-attachment>`, attachmentSgID)

			bucketID, todoID, err := c.getInvoiceTodo(iv)
			if err != nil {
				l.Errorf(err, "failed to get invoice todo", "invoice", iv)
				errsCh <- err
				return
			}

			msg := fmt.Sprintf(`#Invoice %v has been sent

			Confirm Command: Paid @Giang #%v`, iv.Number, iv.Number)

			c.worker.Enqueue(bcModel.BasecampCommentMsg, c.service.Basecamp.BuildCommentMessage(bucketID, todoID, msg, ""))

			errsCh <- nil
		}()
	}

	var count int
	for e := range errsCh {
		if e != nil {
			close(errsCh)
			return nil, err
		}
		count++
		if count == amountGr {
			close(errsCh)
			return iv, nil
		}
	}

	return iv, nil
}

func (c *controller) generateInvoicePDF(l logger.Logger, invoice *model.Invoice) error {
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
	haveIntermediaryBankName := invoice.Bank.IntermediaryBankName != ""
	haveIntermediaryBankAddress := invoice.Bank.IntermediaryBankAddress != ""

	data := &struct {
		Path                        string
		Invoice                     *model.Invoice
		HaveRouting                 bool
		HaveUKSortCode              bool
		HaveSWIFTCode               bool
		HaveIntermediaryBankName    bool
		HaveIntermediaryBankAddress bool
		CompanyContactInfo          *model.CompanyContactInfo
		InvoiceItem                 []model.InvoiceItem
		IntermediaryBankName        string
	}{
		Path:                        c.config.Invoice.TemplatePath,
		Invoice:                     invoice,
		HaveRouting:                 haveRouting,
		HaveUKSortCode:              haveUKSortCode,
		HaveSWIFTCode:               haveSwiftCode,
		HaveIntermediaryBankName:    haveIntermediaryBankName,
		HaveIntermediaryBankAddress: haveIntermediaryBankAddress,
		CompanyContactInfo:          companyInfo,
		InvoiceItem:                 items,
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

	if c.config.Env == "local" {
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
