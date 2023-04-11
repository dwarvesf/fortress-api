package invoice

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/consts"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	bcConst "github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	sInvoice "github.com/dwarvesf/fortress-api/pkg/store/invoice"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
)

type UpdateStatusInput struct {
	InvoiceID         string              `json:"invoiceID"`
	Status            model.InvoiceStatus `json:"status"`
	SendThankYouEmail bool                `json:"sendThankYouEmail"`
}

func (c *controller) UpdateStatus(in UpdateStatusInput) (*model.Invoice, error) {
	l := c.logger.Fields(logger.Fields{
		"controller": "invoice",
		"method":     "UpdateStatus",
		"req":        in,
	})

	// check invoice existence
	invoice, err := c.store.Invoice.One(c.repo.DB(), &sInvoice.Query{ID: in.InvoiceID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(ErrInvoiceNotFound, "invoice not found")
			return nil, ErrInvoiceNotFound
		}

		l.Error(err, "failed to get invoice")
		return nil, err
	}

	if invoice.Status == in.Status {
		l.Error(ErrInvoiceStatusAlready, "invoice status already")
		return nil, ErrInvoiceStatusAlready
	}

	switch in.Status {
	case model.InvoiceStatusError:
		_, err = c.MarkInvoiceAsError(invoice)
	case model.InvoiceStatusPaid:
		_, err = c.MarkInvoiceAsPaid(invoice, in.SendThankYouEmail)
	default:
		_, err = c.store.Invoice.UpdateSelectedFieldsByID(c.repo.DB(), invoice.ID.String(), *invoice, "status")
	}
	if err != nil {
		l.Error(err, "failed to update invoice")
		return nil, err
	}

	return invoice, nil
}

func (c *controller) MarkInvoiceAsError(invoice *model.Invoice) (*model.Invoice, error) {
	l := c.logger.Fields(logger.Fields{
		"controller": "invoice",
		"method":     "MarkInvoiceAsError",
		"req":        invoice,
	})

	tx, done := c.repo.NewTransaction()
	invoice.Status = model.InvoiceStatusError
	iv, err := c.store.Invoice.UpdateSelectedFieldsByID(tx.DB(), invoice.ID.String(), *invoice, "status")
	if err != nil {
		l.Errorf(err, "failed to update invoice status to error")
		return nil, done(err)
	}

	err = c.store.InvoiceNumberCaching.UnCountErrorInvoice(tx.DB(), *invoice.InvoicedAt)
	if err != nil {
		l.Errorf(err, "failed to un-count error invoice")
		return nil, done(err)
	}

	if err := c.markInvoiceTodoAsError(invoice); err != nil {
		return nil, done(err)
	}

	if err := c.service.GoogleDrive.MoveInvoicePDF(invoice, "Sent", "Error"); err != nil {
		l.Errorf(err, "failed to upload invoice pdf to google drive")
		return nil, done(err)
	}

	return iv, done(nil)
}

func (c *controller) markInvoiceTodoAsError(invoice *model.Invoice) error {
	if invoice.Project == nil {
		return fmt.Errorf(`missing project info`)
	}

	bucketID, todoID, err := c.getInvoiceTodo(invoice)
	if err != nil {
		return err
	}

	c.worker.Enqueue(bcModel.BasecampCommentMsg, c.service.Basecamp.BuildCommentMessage(bucketID, todoID, "Invoice has been mark as error", "failed"))

	return c.service.Basecamp.Recording.Archive(bucketID, todoID)
}

type processPaidInvoiceRequest struct {
	Invoice          *model.Invoice
	InvoiceTodoID    int
	InvoiceBucketID  int
	SentThankYouMail bool
}

func (c *controller) MarkInvoiceAsPaid(invoice *model.Invoice, sendThankYouEmail bool) (*model.Invoice, error) {
	l := c.logger.Fields(logger.Fields{
		"controller": "invoice",
		"method":     "MarkInvoiceAsPaid",
		"req":        invoice,
	})

	if invoice.Status != model.InvoiceStatusSent && invoice.Status != model.InvoiceStatusOverdue {
		err := fmt.Errorf(`unable to update invoice status, invoice have status %v`, invoice.Status)
		l.Errorf(err, "failed to update invoice", "invoiceID", invoice.ID.String())
		return nil, err
	}
	invoice.Status = model.InvoiceStatusPaid

	bucketID, todoID, err := c.getInvoiceTodo(invoice)
	if err != nil {
		l.Errorf(err, "failed to get invoice todo", "invoiceID", invoice.ID.String())
		return nil, err
	}

	err = c.service.Basecamp.Todo.Complete(bucketID, todoID)
	if err != nil {
		l.Errorf(err, "failed to complete invoice todo", "invoiceID", invoice.ID.String())
	}

	c.processPaidInvoice(l, &processPaidInvoiceRequest{
		Invoice:          invoice,
		InvoiceTodoID:    todoID,
		InvoiceBucketID:  bucketID,
		SentThankYouMail: sendThankYouEmail,
	})

	return invoice, nil
}

func (c *controller) processPaidInvoice(l logger.Logger, req *processPaidInvoiceRequest) {
	wg := &sync.WaitGroup{}
	wg.Add(3)

	go func() {
		_ = c.processPaidInvoiceData(l, wg, req)
	}()

	go c.sendThankYouEmail(l, wg, req)
	go c.movePaidInvoiceGDrive(l, wg, req)

	wg.Wait()
}

func (c *controller) processPaidInvoiceData(l logger.Logger, wg *sync.WaitGroup, req *processPaidInvoiceRequest) error {
	// Start Transaction
	tx, done := c.repo.NewTransaction()

	msg := bcConst.CommentUpdateInvoiceFailed
	msgType := bcModel.CommentMsgTypeFailed
	defer func() {
		wg.Done()
		c.worker.Enqueue(bcModel.BasecampCommentMsg, c.service.Basecamp.BuildCommentMessage(req.InvoiceBucketID, req.InvoiceTodoID, msg, msgType))
	}()

	now := time.Now()
	req.Invoice.PaidAt = &now
	_, err := c.store.Invoice.UpdateSelectedFieldsByID(tx.DB(), req.Invoice.ID.String(), *req.Invoice, "status", "paid_at")
	if err != nil {
		l.Errorf(err, "failed to update invoice status to paid", "invoice", req.Invoice)
		return done(err)
	}

	_, err = c.storeCommission(tx.DB(), l, req.Invoice)
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

	err = c.store.Accounting.CreateTransaction(tx.DB(), accountingTxn)
	if err != nil {
		l.Errorf(err, "failed to create accounting transaction", "Accounting Transaction", accountingTxn)
		return done(err)
	}

	msg = bcConst.CommentUpdateInvoiceSuccessfully
	msgType = bcModel.CommentMsgTypeCompleted

	return done(nil)
}

func (c *controller) sendThankYouEmail(l logger.Logger, wg *sync.WaitGroup, req *processPaidInvoiceRequest) {
	msg := c.service.Basecamp.BuildCommentMessage(req.InvoiceBucketID, req.InvoiceTodoID, bcConst.CommentThankYouEmailSent, bcModel.CommentMsgTypeCompleted)

	defer func() {
		c.worker.Enqueue(bcModel.BasecampCommentMsg, msg)
		wg.Done()
	}()

	err := c.service.GoogleMail.SendInvoiceThankYouMail(req.Invoice)
	if err != nil {
		l.Errorf(err, "failed to send invoice thank you mail", "invoice", req.Invoice)
		msg = c.service.Basecamp.BuildCommentMessage(req.InvoiceBucketID, req.InvoiceTodoID, bcConst.CommentThankYouEmailFailed, bcModel.CommentMsgTypeFailed)
		return
	}
}

func (c *controller) getInvoiceTodo(iv *model.Invoice) (bucketID, todoID int, err error) {
	if iv.Project == nil {
		return 0, 0, fmt.Errorf(`missing project info`)
	}

	accountingID := consts.AccountingID
	accountingTodoID := consts.AccountingTodoID

	if c.config.Env != "prod" {
		accountingID = consts.PlaygroundID
		accountingTodoID = consts.PlaygroundTodoID
	}

	re := regexp.MustCompile(`Accounting \| ([A-Za-z]+) ([0-9]{4})`)

	todoLists, err := c.service.Basecamp.Todo.GetLists(accountingID, accountingTodoID)
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
		todoList, err = c.service.Basecamp.Todo.CreateList(
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

	todoGroup, err := c.service.Basecamp.Todo.FirstOrCreateGroup(
		accountingID,
		todoList.ID,
		`In`)
	if err != nil {
		return 0, 0, err
	}

	todo, err := c.service.Basecamp.Todo.FirstOrCreateInvoiceTodo(
		accountingID,
		todoGroup.ID,
		iv)
	if err != nil {
		return 0, 0, err
	}

	return accountingID, todo.ID, nil
}
