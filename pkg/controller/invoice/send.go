package invoice

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
)

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

	invoiceItems, err := model.GetInfoItems(iv.LineItems)
	if err != nil {
		l.Errorf(err, "failed to get info items", "invoice-lineItems", iv.LineItems)
		return nil, err
	}

	iv.Bonus = c.getInvoiceBonus(invoiceItems)
	iv.TotalWithoutBonus = iv.Total - iv.Bonus

	if err := c.generateInvoicePDF(l, iv, invoiceItems); err != nil {
		l.Error(err, "failed to generate Invoice PDF")
		return nil, err
	}

	conversionAmount, rate, err := c.service.Wise.Convert(iv.Total, iv.Bank.Currency.Name, "VND")
	if err != nil {
		l.Error(err, "failed to convert currency")
		return nil, err
	}

	am := model.NewVietnamDong(int64(conversionAmount))
	iv.ConversionAmount = float64(am)
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

			if err := c.dispatchInvoiceTask(iv, fn); err != nil {
				l.Error(err, "failed to dispatch invoice task")
				errsCh <- err
				return
			}

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

func (c *controller) dispatchInvoiceTask(iv *model.Invoice, fileName string) error {
	provider := c.service.TaskProvider
	if provider == nil {
		return errors.New("task provider is not configured")
	}

	attachmentRef, err := provider.UploadAttachment(context.Background(), nil, taskprovider.InvoiceAttachmentInput{
		FileName:    fileName,
		ContentType: "application/pdf",
		Content:     iv.InvoiceFileContent,
		URL:         iv.InvoiceFileURL,
	})
	if err != nil {
		return fmt.Errorf("create task attachment: %w", err)
	}
	if strings.TrimSpace(iv.InvoiceFileURL) == "" && attachmentRef != nil && attachmentRef.ExternalID != "" {
		iv.InvoiceFileURL = attachmentRef.ExternalID
	}

	if attachmentRef != nil {
		iv.TodoAttachment = attachmentRef.Markup
		if len(attachmentRef.Meta) > 0 {
			iv.InvoiceAttachmentMeta = attachmentRef.Meta
		}
	} else {
		iv.TodoAttachment = ""
	}

	ref, err := provider.EnsureTask(context.Background(), taskprovider.CreateInvoiceTaskInput{Invoice: iv})
	if err != nil {
		return fmt.Errorf("ensure invoice task: %w", err)
	}

	if err := c.syncAccountingTodoWithInvoice(context.Background(), iv, ref, attachmentRef); err != nil {
		c.logger.Fields(logger.Fields{
			"invoice":    iv.Number,
			"project_id": iv.ProjectID,
		}).Warnf("failed to sync accounting todo with invoice: %v", err)
	}

	msg := fmt.Sprintf(`#Invoice %v has been sent

	Confirm Command: Paid @Giang #%v`, iv.Number, iv.Number)

	taskJob := taskprovider.InvoiceCommentJob{
		Ref: ref,
		Input: taskprovider.InvoiceCommentInput{
			Message: msg,
		},
	}
	c.worker.Enqueue(taskprovider.WorkerMessageInvoiceComment, taskJob)

	return nil
}

func (c *controller) generateInvoicePDF(l logger.Logger, invoice *model.Invoice, items []model.InvoiceItem) error {
	// Use USD currency code for USDC to display "$" symbol instead of "USDC"
	currencyCode := invoice.Project.BankAccount.Currency.Name
	if strings.ToUpper(currencyCode) == "USDC" {
		currencyCode = "USD"
		l.Debug("detected USDC currency, using USD for money formatting to display '$' symbol")
	}
	pound := money.New(1, currencyCode)

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
		"formatMoney": func(money float64) string {
			var result string
			tmpValue := money * math.Pow(10, float64(pound.Currency().Fraction))
			result = pound.Multiply(int64(tmpValue)).Display()

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
		"debugValue": func(label string, value interface{}) string {
			l.Debug(fmt.Sprintf("[TEMPLATE DEBUG] %s = %v", label, value))
			return fmt.Sprintf("%v", value)
		},
		"formatDiscount": func(discountValue float64, discountType string) string {
			// Handle empty or None discount type
			if discountType == "" || discountType == "None" || discountValue == 0 {
				return "0%"
			}

			// Format based on discount type
			if discountType == "Percentage" {
				return fmt.Sprintf("%.0f%%", discountValue)
			}

			// For Fixed Amount and all other types, format as money
			tmpValue := discountValue * math.Pow(10, float64(pound.Currency().Fraction))
			return pound.Multiply(int64(tmpValue)).Display()
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

func (c *controller) syncAccountingTodoWithInvoice(ctx context.Context, iv *model.Invoice, invoiceTaskRef *taskprovider.InvoiceTaskRef, attachmentRef *taskprovider.InvoiceAttachmentRef) error {
	if c.service == nil || c.service.AccountingProvider == nil || c.service.AccountingProvider.Type() != taskprovider.ProviderNocoDB {
		return nil
	}
	if c.service.NocoDB == nil {
		return errors.New("nocodb service not configured")
	}
	if iv == nil || iv.ProjectID.IsZero() {
		return errors.New("missing invoice or project context")
	}
	refs, err := c.store.AccountingTaskRef.FindByProjectMonthYear(
		c.repo.DB(),
		iv.ProjectID.String(),
		iv.Month,
		iv.Year,
		string(taskprovider.AccountingGroupIn),
	)
	if err != nil {
		return err
	}
	if len(refs) == 0 {
		return fmt.Errorf("no accounting todo found for project %s %d/%d", iv.ProjectID, iv.Month, iv.Year)
	}
	target := refs[0]
	row, err := c.service.NocoDB.GetAccountingTodo(ctx, target.TaskRef)
	if err != nil {
		return err
	}
	meta := extractAccountingMetadata(row)
	meta["invoice_id"] = iv.ID.String()
	meta["invoice_number"] = iv.Number
	if invoiceTaskRef != nil && invoiceTaskRef.ExternalID != "" {
		meta["invoice_task_id"] = invoiceTaskRef.ExternalID
	}
	if attachmentRef != nil && attachmentRef.ExternalID != "" {
		meta["attachment_url"] = attachmentRef.ExternalID
	}
	if attachmentRef != nil && len(attachmentRef.Meta) > 0 {
		meta["attachment"] = attachmentRef.Meta
	}
	payload := map[string]interface{}{
		"description": iv.Number,
		"metadata":    meta,
	}
	return c.service.NocoDB.UpdateAccountingTodo(ctx, target.TaskRef, payload)
}

func extractAccountingMetadata(row map[string]interface{}) map[string]interface{} {
	meta := map[string]interface{}{}
	if row == nil {
		return meta
	}
	if raw, ok := row["metadata"]; ok && raw != nil {
		switch val := raw.(type) {
		case map[string]interface{}:
			for k, v := range val {
				meta[k] = v
			}
		case string:
			var decoded map[string]interface{}
			if err := json.Unmarshal([]byte(val), &decoded); err == nil {
				for k, v := range decoded {
					meta[k] = v
				}
			}
		}
	}
	return meta
}

func (c *controller) getInvoiceBonus(items []model.InvoiceItem) float64 {
	var bonus float64
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Description), "bonus") {
			bonus += item.Cost
		}
	}
	return bonus
}

// GenerateInvoicePDFForTest is a public method for testing PDF generation
func (c *controller) GenerateInvoicePDFForTest(l logger.Logger, invoice *model.Invoice, items []model.InvoiceItem) error {
	return c.generateInvoicePDF(l, invoice, items)
}

// GenerateInvoicePDFForNotion is a public method for Notion webhook invoice generation
func (c *controller) GenerateInvoicePDFForNotion(l logger.Logger, invoice *model.Invoice, items []model.InvoiceItem) error {
	return c.generateInvoicePDF(l, invoice, items)
}
