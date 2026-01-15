package worker

import (
	"context"
	"errors"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"
)

type Worker struct {
	ctx     context.Context
	service *service.Service
	queue   chan model.WorkerMessage
	logger  logger.Logger
}

func New(ctx context.Context, queue chan model.WorkerMessage, service *service.Service, logger logger.Logger) *Worker {
	return &Worker{
		ctx:     ctx,
		service: service,
		queue:   queue,
		logger:  logger,
	}
}

func (w *Worker) ProcessMessage() error {
	consumeErr := make(chan error, 1)
	go func() {
		for {
			if w.ctx.Err() != nil {
				consumeErr <- w.ctx.Err()
				return
			}
			message := <-w.queue
			switch message.Type {
			case bcModel.BasecampCommentMsg:
				_ = w.handleCommentMessage(w.logger, message.Payload)

			case bcModel.BasecampTodoMsg:
				_ = w.handleTodoMessage(w.logger, message.Payload)

			case taskprovider.WorkerMessageInvoiceComment:
				_ = w.handleInvoiceCommentJob(w.logger, message.Payload)

			case GenerateInvoiceSplitsMsg:
				_ = w.handleGenerateInvoiceSplits(w.logger, message.Payload)
			default:
				continue
			}
		}
	}()

	select {
	case err := <-consumeErr:
		return err
	case <-w.ctx.Done():
		return nil
	}
}

func (w *Worker) Enqueue(action string, msg interface{}) {
	w.queue <- model.WorkerMessage{Type: action, Payload: msg}
}

func (w *Worker) handleCommentMessage(l logger.Logger, payload interface{}) error {
	m := payload.(bcModel.BasecampCommentMessage)
	err := w.service.Basecamp.Comment.Create(m.ProjectID, m.RecordingID, m.Payload)
	if err != nil {
		l.Errorf(err, "failed to create basecamp comment", "payload", m.Payload.Content)
		return err
	}

	return nil
}

func (w *Worker) handleTodoMessage(l logger.Logger, payload interface{}) error {
	m := payload.(bcModel.BasecampTodoMessageModel)
	_, err := w.service.Basecamp.Todo.Create(m.ProjectID, m.ListID, m.Payload)
	if err != nil {
		l.Errorf(err, "failed to create basecamp todo", "payload", m.Payload.Content)
		return err
	}

	return nil
}

func (w *Worker) handleInvoiceCommentJob(l logger.Logger, payload interface{}) error {
	job, ok := payload.(taskprovider.InvoiceCommentJob)
	if !ok {
		return errors.New("invalid invoice comment job payload")
	}
	if w.service.TaskProvider == nil {
		return errors.New("task provider not configured")
	}
	return w.service.TaskProvider.PostComment(w.ctx, job.Ref, job.Input)
}

func (w *Worker) handleGenerateInvoiceSplits(l logger.Logger, payload interface{}) error {
	l = l.Fields(logger.Fields{
		"worker": "generateInvoiceSplits",
	})

	// Extract payload
	p, ok := payload.(GenerateInvoiceSplitsPayload)
	if !ok {
		l.Error(errors.New("invalid payload"), "failed to cast GenerateInvoiceSplitsPayload")
		return errors.New("invalid generate invoice splits payload")
	}

	l = l.Fields(logger.Fields{
		"invoicePageID": p.InvoicePageID,
	})
	l.Info("processing invoice splits generation")

	// Check if Notion service is available
	if w.service.Notion == nil {
		l.Error(errors.New("notion service not configured"), "cannot process invoice splits")
		return errors.New("notion service not configured")
	}

	// Check if splits already generated (idempotency)
	splitsGenerated, err := w.service.Notion.IsSplitsGenerated(p.InvoicePageID)
	if err != nil {
		l.Error(err, "failed to check if splits generated")
		return err
	}
	if splitsGenerated {
		l.Info("splits already generated, skipping")
		return nil
	}

	// Query line items with commission data
	lineItems, err := w.service.Notion.QueryLineItemsWithCommissions(p.InvoicePageID)
	if err != nil {
		l.Error(err, "failed to query line items with commissions")
		return err
	}

	if len(lineItems) == 0 {
		l.Info("no line items found for invoice")
		// Still mark as generated to avoid re-processing
		if markErr := w.service.Notion.MarkSplitsGenerated(p.InvoicePageID); markErr != nil {
			l.Error(markErr, "failed to mark splits generated")
			return markErr
		}
		return nil
	}

	l.Infof("found %d line items with commissions", len(lineItems))

	// Check if InvoiceSplit service is available
	if w.service.Notion.InvoiceSplit == nil {
		l.Error(errors.New("invoice split service not configured"), "cannot create splits")
		return errors.New("invoice split service not configured")
	}

	// Process each line item
	var createdCount int
	for _, item := range lineItems {
		// Create splits for each role that has an amount > 0
		roles := []struct {
			name      string
			amount    float64
			personIDs []string
		}{
			{"Sales", item.SalesAmount, item.SalesPersonIDs},
			{"Account Manager", item.AccountMgrAmount, item.AccountMgrIDs},
			{"Delivery Lead", item.DeliveryLeadAmount, item.DeliveryLeadIDs},
			{"Hiring Referral", item.HiringRefAmount, item.HiringRefIDs},
		}

		for _, role := range roles {
			if role.amount <= 0 {
				continue
			}

			// Create a split for each person in this role
			for _, personID := range role.personIDs {
				// Build split name: "Sales Commission - ProjectCode Dec 2025"
				splitName := buildSplitName(role.name, item.ProjectCode, item.Month)

				input := notion.CreateCommissionSplitInput{
					Name:              splitName,
					Amount:            role.amount / float64(len(role.personIDs)), // Split equally among people
					Currency:          item.Currency,
					Month:             item.Month,
					Role:              role.name,
					Type:              "Commission",
					Status:            "Pending",
					ContractorPageID:  personID,
					DeploymentPageID:  item.DeploymentPageID,
					InvoiceItemPageID: item.PageID,
					InvoicePageID:     p.InvoicePageID,
				}

				split, err := w.service.Notion.InvoiceSplit.CreateCommissionSplit(w.ctx, input)
				if err != nil {
					l.Errorf(err, "failed to create commission split for role=%s person=%s", role.name, personID)
					// Continue with other splits even if one fails
					continue
				}
				createdCount++

				// Create corresponding payout for this split
				if w.service.Notion.ContractorPayouts != nil && split != nil {
					payoutInput := notion.CreateCommissionPayoutInput{
						Name:             splitName,
						ContractorPageID: personID,
						InvoiceSplitID:   split.PageID,
						Amount:           split.Amount,
						Currency:         split.Currency,
						Date:             item.Month.Format("2006-01-02"),
						Description:      "", // Empty for now, can be populated if needed
					}

					_, payoutErr := w.service.Notion.ContractorPayouts.CreateCommissionPayout(w.ctx, payoutInput)
					if payoutErr != nil {
						l.Errorf(payoutErr, "failed to create payout for split=%s role=%s person=%s", split.PageID, role.name, personID)
						// Continue even if payout creation fails
					} else {
						l.Debugf("created payout for split=%s", split.PageID)
					}
				}
			}
		}
	}

	l.Infof("created %d commission splits", createdCount)

	// Mark splits as generated
	if err := w.service.Notion.MarkSplitsGenerated(p.InvoicePageID); err != nil {
		l.Error(err, "failed to mark splits generated")
		return err
	}

	l.Info("invoice splits generation completed successfully")
	return nil
}

// buildSplitName creates a descriptive name for the split
func buildSplitName(role, projectCode string, month time.Time) string {
	monthStr := month.Format("Jan 2006")
	if projectCode != "" {
		return role + " Commission - " + projectCode + " " + monthStr
	}
	return role + " Commission - " + monthStr
}
