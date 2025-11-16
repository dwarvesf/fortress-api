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

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"
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

	l.Debugf("starting invoice status update: invoiceID=%s targetStatus=%v", in.InvoiceID, in.Status)

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

	l.Debugf("invoice found: invoiceID=%s currentStatus=%v targetStatus=%v number=%s", invoice.ID, invoice.Status, in.Status, invoice.Number)

	if invoice.Status == in.Status {
		l.Debugf("invoice status already matches target status: currentStatus=%v targetStatus=%v", invoice.Status, in.Status)
		l.Error(ErrInvoiceStatusAlready, "invoice status already")
		return nil, ErrInvoiceStatusAlready
	}

	l.Debugf("processing status update: currentStatus=%v targetStatus=%v", invoice.Status, in.Status)

	switch in.Status {
	case model.InvoiceStatusError:
		l.Debug("marking invoice as error")
		_, err = c.MarkInvoiceAsError(invoice)
	case model.InvoiceStatusPaid:
		l.Debugf("marking invoice as paid: sendThankYouEmail=%v", in.SendThankYouEmail)
		_, err = c.MarkInvoiceAsPaid(invoice, in.SendThankYouEmail)
	default:
		l.Debugf("updating invoice status directly: newStatus=%v", in.Status)
		_, err = c.store.Invoice.UpdateSelectedFieldsByID(c.repo.DB(), invoice.ID.String(), *invoice, "status")
	}
	if err != nil {
		l.Error(err, "failed to update invoice")
		return nil, err
	}

	l.Debugf("invoice status updated successfully: invoiceID=%s newStatus=%v", invoice.ID, in.Status)

	return invoice, nil
}

func (c *controller) MarkInvoiceAsError(invoice *model.Invoice) (*model.Invoice, error) {
	l := c.logger.Fields(logger.Fields{
		"controller": "invoice",
		"method":     "MarkInvoiceAsError",
		"req":        invoice,
	})

	l.Debugf("marking invoice as error: invoiceID=%s currentStatus=%v number=%s", invoice.ID, invoice.Status, invoice.Number)

	tx, done := c.repo.NewTransaction()
	invoice.Status = model.InvoiceStatusError
	l.Debugf("updating invoice status to error in database: invoiceID=%s", invoice.ID)
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

	// Skip Basecamp todo management when using NocoDB task provider
	// or when Basecamp service is not available
	if c.config.TaskProvider == "nocodb" || c.service.Basecamp == nil {
		c.logger.Info("skipping Basecamp todo error handling - using NocoDB provider or Basecamp unavailable")
		return nil
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
	TaskRef          *taskprovider.InvoiceTaskRef
	InvoiceTodoID    int
	InvoiceBucketID  int
	SentThankYouMail bool
}

func (c *controller) MarkInvoiceAsPaid(invoice *model.Invoice, sendThankYouEmail bool) (*model.Invoice, error) {
	l := c.logger.Fields(logger.Fields{
		"controller": "invoice",
		"method":     "MarkInvoiceAsPaid",
		"invoiceID":  invoice.ID,
		"number":     invoice.Number,
		"status":     invoice.Status,
	})

	l.Debugf("attempting to mark invoice as paid: invoiceID=%s currentStatus=%v sendThankYouEmail=%v", invoice.ID, invoice.Status, sendThankYouEmail)

	if invoice.Status != model.InvoiceStatusSent && invoice.Status != model.InvoiceStatusOverdue {
		err := fmt.Errorf(`unable to update invoice status, invoice have status %v`, invoice.Status)
		l.Debugf("invoice status validation failed: currentStatus=%v allowedStatuses=%v", invoice.Status, []string{"sent", "overdue"})
		l.Errorf(err, "failed to update invoice", "invoiceID", invoice.ID.String())
		return nil, err
	}

	l.Debug("invoice status validation passed, proceeding to mark as paid")

	// Skip Basecamp todo management when using NocoDB task provider
	// or when Basecamp service is not available (e.g., fallback for migrated invoices)
	// This condition ensures compatibility with NocoDB task provider when some invoices are still linked to Basecamp todos.
	if c.config.TaskProvider == "nocodb" || c.service.Basecamp == nil {
		l.Info("skipping Basecamp todo management - using NocoDB provider or Basecamp unavailable")
		return c.MarkInvoiceAsPaidWithTaskRef(invoice, nil, sendThankYouEmail)
	}

	bucketID, todoID, err := c.getInvoiceTodo(invoice)
	if err != nil {
		l.Errorf(err, "failed to get invoice todo", "invoiceID", invoice.ID.String())
		return nil, err
	}

	err = c.service.Basecamp.Todo.Complete(bucketID, todoID)
	if err != nil {
		l.Errorf(err, "failed to complete invoice todo", "invoiceID", invoice.ID.String())
	}

	ref := &taskprovider.InvoiceTaskRef{
		Provider:   taskprovider.ProviderBasecamp,
		ExternalID: strconv.Itoa(todoID),
		BucketID:   bucketID,
		TodoID:     todoID,
	}

	return c.MarkInvoiceAsPaidWithTaskRef(invoice, ref, sendThankYouEmail)
}

func (c *controller) MarkInvoiceAsPaidByBasecampWebhookMessage(invoice *model.Invoice, msg *model.BasecampWebhookMessage) (*model.Invoice, error) {
	ref := &taskprovider.InvoiceTaskRef{
		Provider:   taskprovider.ProviderBasecamp,
		ExternalID: strconv.Itoa(msg.Recording.ID),
		BucketID:   msg.Recording.Bucket.ID,
		TodoID:     msg.Recording.ID,
	}

	return c.MarkInvoiceAsPaidWithTaskRef(invoice, ref, true)
}

func (c *controller) MarkInvoiceAsPaidWithTaskRef(invoice *model.Invoice, ref *taskprovider.InvoiceTaskRef, sendThankYouEmail bool) (*model.Invoice, error) {
	l := c.logger.Fields(logger.Fields{
		"controller": "invoice",
		"method":     "MarkInvoiceAsPaidWithTaskRef",
		"invoiceID":  invoice.ID,
		"number":     invoice.Number,
		"status":     invoice.Status,
	})

	l.Debugf("attempting to mark invoice as paid with task ref: invoiceID=%s currentStatus=%v hasTaskRef=%v sendThankYouEmail=%v", invoice.ID, invoice.Status, ref != nil, sendThankYouEmail)

	if invoice.Status != model.InvoiceStatusSent && invoice.Status != model.InvoiceStatusOverdue {
		err := fmt.Errorf(`unable to update invoice status, invoice have status %v`, invoice.Status)
		l.Debugf("invoice status validation failed in MarkInvoiceAsPaidWithTaskRef: currentStatus=%v allowedStatuses=%v", invoice.Status, []string{"sent", "overdue"})
		l.Errorf(err, "failed to update invoice", "invoiceID", invoice.ID.String())
		return nil, err
	}

	l.Debug("invoice status validation passed in MarkInvoiceAsPaidWithTaskRef, proceeding to process paid invoice")
	invoice.Status = model.InvoiceStatusPaid

	var todoID, bucketID int
	if ref != nil {
		todoID = ref.TodoID
		bucketID = ref.BucketID
	}

	c.processPaidInvoice(l, &processPaidInvoiceRequest{
		Invoice:          invoice,
		TaskRef:          ref,
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
	l.Debugf("starting processPaidInvoiceData: invoiceID=%s number=%s status=%v", req.Invoice.ID, req.Invoice.Number, req.Invoice.Status)

	// Start Transaction
	tx, done := c.repo.NewTransaction()

	msg := consts.CommentUpdateInvoiceFailed
	msgType := bcModel.CommentMsgTypeFailed
	defer func() {
		wg.Done()
		c.enqueueInvoiceComment(req.TaskRef, req.InvoiceBucketID, req.InvoiceTodoID, msg, msgType)
	}()

	now := time.Now()
	req.Invoice.PaidAt = &now
	l.Debugf("updating invoice status and paid_at in database: invoiceID=%s newStatus=%v paidAt=%v", req.Invoice.ID, req.Invoice.Status, now)
	_, err := c.store.Invoice.UpdateSelectedFieldsByID(tx.DB(), req.Invoice.ID.String(), *req.Invoice, "status", "paid_at")
	if err != nil {
		l.Errorf(err, "failed to update invoice status to paid", "invoice", req.Invoice)
		return done(err)
	}

	l.Debug("invoice status and paid_at updated successfully in database")

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
		Amount:           req.Invoice.Total,
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

	msg = consts.CommentUpdateInvoiceSuccessfully
	msgType = bcModel.CommentMsgTypeCompleted

	return done(nil)
}

func (c *controller) sendThankYouEmail(l logger.Logger, wg *sync.WaitGroup, req *processPaidInvoiceRequest) {
	msgType := bcModel.CommentMsgTypeCompleted
	message := consts.CommentThankYouEmailSent

	defer func() {
		c.enqueueInvoiceComment(req.TaskRef, req.InvoiceBucketID, req.InvoiceTodoID, message, msgType)
		wg.Done()
	}()

	err := c.service.GoogleMail.SendInvoiceThankYouMail(req.Invoice)
	if err != nil {
		l.Errorf(err, "failed to send invoice thank you mail", "invoice", req.Invoice)
		message = consts.CommentThankYouEmailFailed
		msgType = bcModel.CommentMsgTypeFailed
		return
	}
}

func (c *controller) getInvoiceTodo(iv *model.Invoice) (bucketID, todoID int, err error) {
	if iv.Project == nil {
		return 0, 0, fmt.Errorf(`missing project info`)
	}

	accountingID := c.config.Basecamp.AccountingProjectID
	accountingTodoID := c.config.Basecamp.AccountingTodoSetID
	playgroundProjectID := c.config.Basecamp.PlaygroundProjectID
	playgroundTodoID := c.config.Basecamp.PlaygroundTodoSetID

	if accountingID == 0 {
		accountingID = consts.AccountingID
	}
	if accountingTodoID == 0 {
		accountingTodoID = consts.AccountingTodoID
	}
	if playgroundProjectID == 0 {
		playgroundProjectID = consts.PlaygroundID
	}
	if playgroundTodoID == 0 {
		playgroundTodoID = consts.PlaygroundTodoID
	}

	if c.config.Env != "prod" {
		accountingID = playgroundProjectID
		accountingTodoID = playgroundTodoID
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

func (c *controller) enqueueInvoiceComment(ref *taskprovider.InvoiceTaskRef, bucketID, todoID int, message, msgType string) {
	if ref != nil && c.service.TaskProvider != nil {
		c.worker.Enqueue(taskprovider.WorkerMessageInvoiceComment, taskprovider.InvoiceCommentJob{
			Ref: ref,
			Input: taskprovider.InvoiceCommentInput{
				Message: message,
				Type:    msgType,
			},
		})
		return
	}

	// Skip Basecamp comment when Basecamp service is not available (e.g., using NocoDB provider)
	if c.service.Basecamp == nil {
		c.logger.Debugf("skipping Basecamp comment enqueue - Basecamp service unavailable: message=%s", message)
		return
	}

	c.worker.Enqueue(bcModel.BasecampCommentMsg, c.service.Basecamp.BuildCommentMessage(bucketID, todoID, message, msgType))
}
