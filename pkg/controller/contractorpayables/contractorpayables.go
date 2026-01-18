package contractorpayables

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
)

const (
	maxRetries    = 3
	retryInterval = 500 * time.Millisecond
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

	// Query all pending payables for the month and batch
	payables, err := c.service.Notion.ContractorPayables.QueryPendingPayablesByPeriod(ctx, month, batch)
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

	// Query all pending payables for the month and batch
	payables, err := c.service.Notion.ContractorPayables.QueryPendingPayablesByPeriod(ctx, month, batch)
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

	l.Debug(fmt.Sprintf("committing %d payables in parallel for contractor=%s", len(toCommit), contractor))

	// Execute cascade updates for each payable in parallel
	var successCount, failCount int
	var errors []CommitError
	var mu sync.Mutex

	g := new(errgroup.Group)

	for _, payable := range toCommit {
		g.Go(func() error {
			if err := c.commitSinglePayable(ctx, payable); err != nil {
				l.Error(err, fmt.Sprintf("failed to commit payable %s", payable.PageID))
				mu.Lock()
				failCount++
				errors = append(errors, CommitError{
					PayableID: payable.PageID,
					Error:     err.Error(),
				})
				mu.Unlock()
			} else {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
			return nil // continue with others even on error
		})
	}

	// Wait for all payables to be processed
	_ = g.Wait()

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

	l.Debug(fmt.Sprintf("starting cascade update for %d payout items", len(payable.PayoutItemPageIDs)))

	// Step 1: Update all Payout Items in parallel (best-effort)
	g := new(errgroup.Group)

	for _, payoutPageID := range payable.PayoutItemPageIDs {
		g.Go(func() error {
			if err := c.commitPayoutItem(ctx, payoutPageID); err != nil {
				l.Error(err, fmt.Sprintf("failed to commit payout item %s", payoutPageID))
			}
			return nil // best-effort, don't fail the group
		})
	}

	// Wait for all payout items to be processed
	_ = g.Wait()

	// Step 2: Update the Contractor Payable itself (with retry)
	paymentDate := time.Now().Format("2006-01-02")
	if err := c.retryUpdatePayableStatus(ctx, payable.PageID, paymentDate, l); err != nil {
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

	// Update Invoice Split and Refund Request in parallel (best-effort)
	g := new(errgroup.Group)

	if payout.InvoiceSplitID != "" {
		invoiceSplitID := payout.InvoiceSplitID
		g.Go(func() error {
			l.Debug(fmt.Sprintf("updating invoice split %s", invoiceSplitID))
			if err := c.service.Notion.InvoiceSplit.UpdateInvoiceSplitStatus(ctx, invoiceSplitID, "Paid"); err != nil {
				l.Error(err, "failed to update invoice split")
			}
			return nil // best-effort, don't fail the group
		})
	}

	if payout.RefundRequestID != "" {
		refundRequestID := payout.RefundRequestID
		g.Go(func() error {
			l.Debug(fmt.Sprintf("updating refund request %s", refundRequestID))
			if err := c.service.Notion.RefundRequests.UpdateRefundRequestStatus(ctx, refundRequestID, "Paid"); err != nil {
				l.Error(err, "failed to update refund request")
			}
			return nil // best-effort, don't fail the group
		})
	}

	// Wait for related updates to complete
	_ = g.Wait()

	// Update Payout Item status (with retry)
	if err := c.retryUpdatePayoutStatus(ctx, payoutPageID, l); err != nil {
		return fmt.Errorf("failed to update payout status: %w", err)
	}

	return nil
}

// retryUpdatePayableStatus retries UpdatePayableStatus on transient errors
func (c *controller) retryUpdatePayableStatus(ctx context.Context, pageID, paymentDate string, l logger.Logger) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			l.Debug(fmt.Sprintf("retrying UpdatePayableStatus attempt %d/%d for %s", i+1, maxRetries, pageID))
			time.Sleep(retryInterval)
		}

		err := c.service.Notion.ContractorPayables.UpdatePayableStatus(ctx, pageID, "Paid", paymentDate)
		if err == nil {
			return nil
		}

		lastErr = err
		// Only retry on transient errors (context canceled, network issues)
		if !isRetryableError(err) {
			l.Error(err, "non-retryable error updating payable status")
			return err
		}
		l.Debug(fmt.Sprintf("retryable error updating payable status: %v", err))
	}

	l.Error(lastErr, fmt.Sprintf("failed to update payable status after %d retries", maxRetries))
	return lastErr
}

// retryUpdatePayoutStatus retries UpdatePayoutStatus on transient errors
func (c *controller) retryUpdatePayoutStatus(ctx context.Context, pageID string, l logger.Logger) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			l.Debug(fmt.Sprintf("retrying UpdatePayoutStatus attempt %d/%d for %s", i+1, maxRetries, pageID))
			time.Sleep(retryInterval)
		}

		err := c.service.Notion.ContractorPayouts.UpdatePayoutStatus(ctx, pageID, "Paid")
		if err == nil {
			return nil
		}

		lastErr = err
		if !isRetryableError(err) {
			l.Error(err, "non-retryable error updating payout status")
			return err
		}
		l.Debug(fmt.Sprintf("retryable error updating payout status: %v", err))
	}

	l.Error(lastErr, fmt.Sprintf("failed to update payout status after %d retries", maxRetries))
	return lastErr
}

// isRetryableError checks if error is transient and can be retried
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "context canceled") ||
		strings.Contains(errStr, "context deadline exceeded") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "EOF") ||
		strings.Contains(errStr, "timeout")
}
