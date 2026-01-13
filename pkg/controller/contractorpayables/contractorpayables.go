package contractorpayables

import (
	"context"
	"fmt"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
)

type controller struct {
	config  *config.Config
	logger  logger.Logger
	service *service.Service
}

// New creates a new contractor payables controller
func New(service *service.Service, logger logger.Logger, cfg *config.Config) IController {
	return &controller{
		config:  cfg,
		logger:  logger,
		service: service,
	}
}

// PreviewCommit queries pending payables and returns preview data
func (c *controller) PreviewCommit(ctx context.Context, month string, batch int, contractor string) (*PreviewCommitResponse, error) {
	l := c.logger.Fields(logger.Fields{
		"controller": "contractorpayables",
		"method":     "PreviewCommit",
		"month":      month,
		"batch":      batch,
		"contractor": contractor,
	})

	l.Debug("querying pending payables")

	// Convert month to period date (YYYY-MM-01)
	period := month + "-01"

	// Query all pending payables for the period
	payables, err := c.service.Notion.ContractorPayables.QueryPendingPayablesByPeriod(ctx, period)
	if err != nil {
		l.Error(err, "failed to query pending payables")
		return nil, fmt.Errorf("failed to query pending payables: %w", err)
	}

	l.Debug(fmt.Sprintf("found %d pending payables", len(payables)))

	// Filter by PayDay (batch) and optionally by contractor (Discord username)
	var filtered []ContractorPreview
	var totalAmount float64

	for _, payable := range payables {
		// Filter by contractor Discord username if specified
		if contractor != "" && payable.Discord != contractor {
			l.Debug(fmt.Sprintf("skipping payable %s: discord %s != %s", payable.PageID, payable.Discord, contractor))
			continue
		}

		// Get contractor's PayDay from Service Rate
		payDay, err := c.service.Notion.ContractorPayables.GetContractorPayDay(ctx, payable.ContractorPageID)
		if err != nil {
			l.Debug(fmt.Sprintf("failed to get PayDay for contractor %s: %v", payable.ContractorPageID, err))
			continue
		}

		// Filter by batch
		if payDay != batch {
			l.Debug(fmt.Sprintf("skipping payable %s: PayDay %d != batch %d", payable.PageID, payDay, batch))
			continue
		}

		preview := ContractorPreview{
			Name:      payable.ContractorName,
			Amount:    payable.Total,
			Currency:  payable.Currency,
			PayableID: payable.PageID,
		}
		filtered = append(filtered, preview)
		totalAmount += payable.Total
	}

	l.Debug(fmt.Sprintf("filtered to %d payables for batch %d contractor=%s", len(filtered), batch, contractor))

	// Return empty list instead of nil if no payables found
	if filtered == nil {
		filtered = []ContractorPreview{}
	}

	return &PreviewCommitResponse{
		Month:       month,
		Batch:       batch,
		Count:       len(filtered),
		TotalAmount: totalAmount,
		Contractors: filtered,
	}, nil
}

// CommitPayables executes the cascade status update for all matching payables
func (c *controller) CommitPayables(ctx context.Context, month string, batch int, contractor string) (*CommitResponse, error) {
	l := c.logger.Fields(logger.Fields{
		"controller": "contractorpayables",
		"method":     "CommitPayables",
		"month":      month,
		"batch":      batch,
		"contractor": contractor,
	})

	l.Debug("starting commit operation")

	// Convert month to period date (YYYY-MM-01)
	period := month + "-01"

	// Query all pending payables for the period
	payables, err := c.service.Notion.ContractorPayables.QueryPendingPayablesByPeriod(ctx, period)
	if err != nil {
		l.Error(err, "failed to query pending payables")
		return nil, fmt.Errorf("failed to query pending payables: %w", err)
	}

	if len(payables) == 0 {
		return nil, fmt.Errorf("no pending payables found for month %s", month)
	}

	// Filter by PayDay and optionally by contractor (Discord username)
	var toCommit []payableToCommit
	for _, payable := range payables {
		// Filter by contractor Discord username if specified
		if contractor != "" && payable.Discord != contractor {
			l.Debug(fmt.Sprintf("skipping payable %s: discord %s != %s", payable.PageID, payable.Discord, contractor))
			continue
		}

		payDay, err := c.service.Notion.ContractorPayables.GetContractorPayDay(ctx, payable.ContractorPageID)
		if err != nil {
			l.Debug(fmt.Sprintf("failed to get PayDay for contractor %s: %v", payable.ContractorPageID, err))
			continue
		}

		if payDay == batch {
			toCommit = append(toCommit, payableToCommit{
				PageID:            payable.PageID,
				ContractorPageID:  payable.ContractorPageID,
				PayoutItemPageIDs: payable.PayoutItemPageIDs,
			})
		}
	}

	if len(toCommit) == 0 {
		return nil, fmt.Errorf("no pending payables found for month %s batch %d contractor=%s", month, batch, contractor)
	}

	l.Debug(fmt.Sprintf("committing %d payables for contractor=%s", len(toCommit), contractor))

	// Execute cascade updates for each payable
	var successCount, failCount int
	var errors []CommitError

	for _, payable := range toCommit {
		if err := c.commitSinglePayable(ctx, payable); err != nil {
			l.Error(err, fmt.Sprintf("failed to commit payable %s", payable.PageID))
			failCount++
			errors = append(errors, CommitError{
				PayableID: payable.PageID,
				Error:     err.Error(),
			})
		} else {
			successCount++
		}
	}

	l.Info(fmt.Sprintf("commit complete: %d succeeded, %d failed", successCount, failCount))

	return &CommitResponse{
		Month:   month,
		Batch:   batch,
		Updated: successCount,
		Failed:  failCount,
		Errors:  errors,
	}, nil
}

// payableToCommit contains the data needed to commit a single payable
type payableToCommit struct {
	PageID            string
	ContractorPageID  string
	PayoutItemPageIDs []string
}

// commitSinglePayable performs the cascade update for a single payable
func (c *controller) commitSinglePayable(ctx context.Context, payable payableToCommit) error {
	l := c.logger.Fields(logger.Fields{
		"controller": "contractorpayables",
		"method":     "commitSinglePayable",
		"payable_id": payable.PageID,
	})

	l.Debug("starting cascade update")

	// Step 1: Update each Payout Item and its related Invoice Split/Refund
	for _, payoutPageID := range payable.PayoutItemPageIDs {
		if err := c.commitPayoutItem(ctx, payoutPageID); err != nil {
			l.Error(err, fmt.Sprintf("failed to commit payout item %s", payoutPageID))
			// Continue with other payouts (best-effort)
		}
	}

	// Step 2: Update the Contractor Payable itself
	paymentDate := time.Now().Format("2006-01-02")
	if err := c.service.Notion.ContractorPayables.UpdatePayableStatus(ctx, payable.PageID, "Paid", paymentDate); err != nil {
		l.Error(err, "failed to update payable status")
		return fmt.Errorf("failed to update payable status: %w", err)
	}

	l.Debug("cascade update complete")
	return nil
}

// commitPayoutItem updates a payout item and its related records
func (c *controller) commitPayoutItem(ctx context.Context, payoutPageID string) error {
	l := c.logger.Fields(logger.Fields{
		"controller":     "contractorpayables",
		"method":         "commitPayoutItem",
		"payout_page_id": payoutPageID,
	})

	// Get payout with relations (Invoice Split, Refund)
	payout, err := c.service.Notion.ContractorPayouts.GetPayoutWithRelations(ctx, payoutPageID)
	if err != nil {
		l.Error(err, "failed to get payout with relations")
		return fmt.Errorf("failed to get payout with relations: %w", err)
	}

	// Update Invoice Split if exists
	if payout.InvoiceSplitID != "" {
		l.Debug(fmt.Sprintf("updating invoice split %s", payout.InvoiceSplitID))
		if err := c.service.Notion.InvoiceSplit.UpdateInvoiceSplitStatus(ctx, payout.InvoiceSplitID, "Paid"); err != nil {
			l.Error(err, "failed to update invoice split")
			// Continue (best-effort)
		}
	}

	// Update Refund Request if exists
	if payout.RefundRequestID != "" {
		l.Debug(fmt.Sprintf("updating refund request %s", payout.RefundRequestID))
		if err := c.service.Notion.RefundRequests.UpdateRefundRequestStatus(ctx, payout.RefundRequestID, "Paid"); err != nil {
			l.Error(err, "failed to update refund request")
			// Continue (best-effort)
		}
	}

	// Update Payout Item status
	l.Debug("updating payout status")
	if err := c.service.Notion.ContractorPayouts.UpdatePayoutStatus(ctx, payoutPageID, "Paid"); err != nil {
		l.Error(err, "failed to update payout status")
		return fmt.Errorf("failed to update payout status: %w", err)
	}

	return nil
}
