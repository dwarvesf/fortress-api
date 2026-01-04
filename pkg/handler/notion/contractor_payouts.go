package notion

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	notionsvc "github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

// formatMonthYear converts YYYY-MM to "Month, Year" format
func formatMonthYear(month string) string {
	if month == "" {
		return ""
	}
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return month // Return as-is if parsing fails
	}
	return t.Format("January, 2006")
}

var (
	PayoutType = map[string]string{
		"contractor_payroll": "Service Fee",
		"bonus":              "Bonus",
		"commission":         "Commission",
		"refund":             "Refund",
	}
)

// CreateContractorPayouts godoc
// @Summary Create contractor payouts from approved orders or other sources
// @Description Processes approved Task Order Log entries (type=contractor_payroll) or other payout sources and creates payout entries
// @Tags Cronjobs
// @Accept json
// @Produce json
// @Param type query string false "Payout type (default: contractor_payroll)"
// @Param contractor query string false "Filter by contractor (discord username, name, or page ID)"
// @Security BearerAuth
// @Success 200 {object} view.Response
// @Failure 500 {object} view.Response
// @Router /cronjobs/contractor-payouts [post]
func (h *handler) CreateContractorPayouts(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "Notion",
		"method":  "CreateContractorPayouts",
	})

	// Get optional type parameter (default: contractor_payroll)
	payoutTypeKey := c.Query("type")
	if payoutTypeKey == "" {
		payoutTypeKey = "contractor_payroll"
	}

	payoutType, ok := PayoutType[payoutTypeKey]
	if !ok {
		err := fmt.Errorf("invalid payout type: %s", payoutTypeKey)
		l.Error(err, "invalid payout type provided")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Get optional contractor filter (discord username or page ID)
	contractorFilter := c.Query("contractor")

	// Get optional pay_day filter (1-31)
	payDayFilter := 0
	if payDayStr := c.Query("pay_day"); payDayStr != "" {
		if pd, err := fmt.Sscanf(payDayStr, "%d", &payDayFilter); err != nil || pd != 1 {
			l.Debug(fmt.Sprintf("invalid pay_day parameter: %s", payDayStr))
			payDayFilter = 0
		}
	}

	l.Debug(fmt.Sprintf("payout type key: %s, value: %s, contractor: %s, pay_day: %d", payoutTypeKey, payoutType, contractorFilter, payDayFilter))
	l.Info("starting CreateContractorPayouts cronjob")

	// Process based on payout type
	switch payoutTypeKey {
	case "contractor_payroll":
		h.processContractorPayrollPayouts(c, l, payoutType, contractorFilter, payDayFilter)
	case "bonus":
		h.processBonusPayouts(c, l, payoutType)
	case "commission":
		h.processCommissionPayouts(c, l, payoutType)
	case "refund":
		h.processRefundPayouts(c, l, payoutType)
	default:
		err := fmt.Errorf("unknown payout type: %s", payoutTypeKey)
		l.Error(err, "unknown payout type")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
	}
}

// processContractorPayrollPayouts processes approved Task Order Log entries
// and creates payout entries of type "Service Fee"
// contractorFilter: optional filter by contractor discord username or page ID
// payDayFilter: optional filter by pay day (1-31), 0 means no filter
func (h *handler) processContractorPayrollPayouts(c *gin.Context, l logger.Logger, payoutType string, contractorFilter string, payDayFilter int) {
	ctx := c.Request.Context()

	l.Debug(fmt.Sprintf("processContractorPayrollPayouts: contractorFilter=%s payDayFilter=%d", contractorFilter, payDayFilter))

	// Get services
	taskOrderLogService := h.service.Notion.TaskOrderLog
	if taskOrderLogService == nil {
		err := fmt.Errorf("task order log service not configured")
		l.Error(err, "task order log service is nil")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	contractorRatesService := h.service.Notion.ContractorRates
	if contractorRatesService == nil {
		err := fmt.Errorf("contractor rates service not configured")
		l.Error(err, "contractor rates service is nil")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	contractorPayoutsService := h.service.Notion.ContractorPayouts
	if contractorPayoutsService == nil {
		err := fmt.Errorf("contractor payouts service not configured")
		l.Error(err, "contractor payouts service is nil")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Query approved orders (Type=Order, Status=Approved)
	l.Debug("querying Task Order Log with Type=Order, Status=Approved")
	approvedOrders, err := taskOrderLogService.QueryApprovedOrders(ctx)
	if err != nil {
		l.Error(err, "failed to query approved orders")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Info(fmt.Sprintf("found %d approved orders", len(approvedOrders)))

	// Filter by contractor if specified
	if contractorFilter != "" {
		l.Debug(fmt.Sprintf("filtering orders by contractor: %s", contractorFilter))
		var filteredOrders []*notionsvc.ApprovedOrderData
		for _, order := range approvedOrders {
			// Match by discord username or page ID
			if order.ContractorDiscord == contractorFilter ||
				order.ContractorPageID == contractorFilter ||
				order.ContractorName == contractorFilter {
				l.Debug(fmt.Sprintf("order %s matches contractor filter", order.PageID))
				filteredOrders = append(filteredOrders, order)
			}
		}
		l.Debug(fmt.Sprintf("filtered from %d to %d orders", len(approvedOrders), len(filteredOrders)))
		approvedOrders = filteredOrders
	}

	if len(approvedOrders) == 0 {
		l.Info("no approved orders found (after filtering), returning success with zero counts")
		c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
			"payouts_created":  0,
			"orders_processed": 0,
			"orders_skipped":   0,
			"errors":           0,
			"details":          []any{},
			"type":             payoutType,
		}, nil, nil, nil, "ok"))
		return
	}

	// Process orders concurrently with worker pool
	const maxWorkers = 5 // Limit concurrent Notion API calls
	l.Debug(fmt.Sprintf("processing %d orders with %d workers", len(approvedOrders), maxWorkers))

	type orderResult struct {
		detail         map[string]any
		payoutCreated  bool
		skipped        bool
		hasError       bool
	}

	// Channels for work distribution
	ordersChan := make(chan *notionsvc.ApprovedOrderData, len(approvedOrders))
	resultsChan := make(chan orderResult, len(approvedOrders))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for order := range ordersChan {
				l.Debug(fmt.Sprintf("[worker-%d] processing order: %s contractor: %s", workerID, order.PageID, order.ContractorName))

				// Extract month from date
				month := order.Date.Format("2006-01")

				detail := map[string]any{
					"order_page_id":   order.PageID,
					"contractor_name": order.ContractorName,
					"contractor_id":   order.ContractorPageID,
					"month":           month,
					"hours":           order.FinalHoursWorked,
					"payout_page_id":  nil,
					"status":          "",
					"reason":          nil,
				}

				// Validate contractor
				if order.ContractorPageID == "" {
					l.Warn(fmt.Sprintf("[worker-%d] skipping order %s: no contractor found", workerID, order.PageID))
					detail["status"] = "skipped"
					detail["reason"] = "contractor not found in relation"
					resultsChan <- orderResult{detail: detail, skipped: true}
					continue
				}

				// Get contractor rate
				l.Debug(fmt.Sprintf("[worker-%d] fetching contractor rate: contractor=%s date=%s", workerID, order.ContractorPageID, order.Date.Format("2006-01-02")))
				rate, err := contractorRatesService.FindActiveRateByContractor(ctx, order.ContractorPageID, order.Date)
				if err != nil {
					l.Error(err, fmt.Sprintf("[worker-%d] failed to get contractor rate for order %s", workerID, order.PageID))
					detail["status"] = "error"
					detail["reason"] = fmt.Sprintf("failed to get contractor rate: %v", err)
					resultsChan <- orderResult{detail: detail, hasError: true}
					continue
				}

				// Filter by pay day if specified
				if payDayFilter > 0 && rate.PayDay != payDayFilter {
					l.Debug(fmt.Sprintf("[worker-%d] skipping order %s: pay day mismatch (rate=%d, filter=%d)", workerID, order.PageID, rate.PayDay, payDayFilter))
					detail["status"] = "skipped"
					detail["reason"] = fmt.Sprintf("pay day mismatch: rate has %d, filter is %d", rate.PayDay, payDayFilter)
					resultsChan <- orderResult{detail: detail, skipped: true}
					continue
				}

				// Calculate amount based on billing type
				var amount float64
				if rate.BillingType == "Monthly Fixed" {
					amount = rate.MonthlyFixed
					l.Debug(fmt.Sprintf("[worker-%d] using monthly fixed rate: %.2f", workerID, amount))
				} else {
					amount = rate.HourlyRate * order.FinalHoursWorked
					l.Debug(fmt.Sprintf("[worker-%d] using hourly rate: %.2f * %.2f = %.2f", workerID, rate.HourlyRate, order.FinalHoursWorked, amount))
				}

				detail["amount"] = amount
				detail["currency"] = rate.Currency
				detail["billing_type"] = rate.BillingType

				// Check if payout exists (idempotency)
				l.Debug(fmt.Sprintf("[worker-%d] checking if payout exists for order: %s", workerID, order.PageID))
				exists, existingPayoutID, err := contractorPayoutsService.CheckPayoutExistsByContractorFee(ctx, order.PageID)
				if err != nil {
					l.Error(err, fmt.Sprintf("[worker-%d] failed to check payout existence for order %s", workerID, order.PageID))
					detail["status"] = "error"
					detail["reason"] = "failed to check payout existence"
					resultsChan <- orderResult{detail: detail, hasError: true}
					continue
				}

				if exists {
					l.Debug(fmt.Sprintf("[worker-%d] payout already exists for order %s: %s", workerID, order.PageID, existingPayoutID))
					detail["status"] = "skipped"
					detail["reason"] = "payout already exists"
					detail["payout_page_id"] = existingPayoutID
					resultsChan <- orderResult{detail: detail, skipped: true}
					continue
				}

				// Description is empty for Service Fee payout type
				description := ""
				l.Debug(fmt.Sprintf("[worker-%d] description set to empty for Service Fee payout", workerID))

				// Create payout
				payoutName := fmt.Sprintf("Development work on %s", formatMonthYear(month))
				l.Debug(fmt.Sprintf("[worker-%d] creating payout for order: %s name: %s amount: %.2f %s", workerID, order.PageID, payoutName, amount, rate.Currency))

				payoutInput := notionsvc.CreatePayoutInput{
					Name:             payoutName,
					ContractorPageID: order.ContractorPageID,
					TaskOrderID:      order.PageID,
					ServiceRateID:    rate.PageID,
					Amount:           amount,
					Currency:         rate.Currency,
					Date:             order.Date.Format("2006-01-02"),
					Description:      description,
				}

				payoutPageID, err := contractorPayoutsService.CreatePayout(ctx, payoutInput)
				if err != nil {
					l.Error(err, fmt.Sprintf("[worker-%d] failed to create payout for order %s", workerID, order.PageID))
					detail["status"] = "error"
					detail["reason"] = "failed to create payout"
					resultsChan <- orderResult{detail: detail, hasError: true}
					continue
				}

				l.Info(fmt.Sprintf("[worker-%d] created payout: %s for order: %s", workerID, payoutPageID, order.PageID))

				// Update Task Order Log and subitems status to "Completed"
				l.Debug(fmt.Sprintf("[worker-%d] updating order %s and subitems status to Completed", workerID, order.PageID))
				err = taskOrderLogService.UpdateOrderAndSubitemsStatus(ctx, order.PageID, "Completed")
				if err != nil {
					// Log error but don't fail - payout is already created
					l.Error(err, fmt.Sprintf("[worker-%d] failed to update order/subitems status: %s (payout created: %s)", workerID, order.PageID, payoutPageID))
				} else {
					l.Debug(fmt.Sprintf("[worker-%d] updated order %s and subitems status to Completed", workerID, order.PageID))
				}

				detail["status"] = "created"
				detail["payout_page_id"] = payoutPageID
				resultsChan <- orderResult{detail: detail, payoutCreated: true}
			}
		}(i)
	}

	// Send orders to workers
	for _, order := range approvedOrders {
		ordersChan <- order
	}
	close(ordersChan)

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	var (
		payoutsCreated = 0
		ordersSkipped  = 0
		errors         = 0
		details        = []map[string]any{}
	)

	for result := range resultsChan {
		details = append(details, result.detail)
		if result.payoutCreated {
			payoutsCreated++
		} else if result.skipped {
			ordersSkipped++
		} else if result.hasError {
			errors++
		}
	}

	// Return response
	l.Info(fmt.Sprintf("processing complete: payouts_created=%d skipped=%d errors=%d", payoutsCreated, ordersSkipped, errors))

	c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
		"payouts_created":  payoutsCreated,
		"orders_processed": len(approvedOrders),
		"orders_skipped":   ordersSkipped,
		"errors":           errors,
		"details":          details,
		"type":             payoutType,
	}, nil, nil, nil, "ok"))
}

// processRefundPayouts processes approved refund requests
// and creates payout entries of type "Refund"
func (h *handler) processRefundPayouts(c *gin.Context, l logger.Logger, payoutType string) {
	ctx := c.Request.Context()

	// Get services
	refundRequestsService := h.service.Notion.RefundRequests
	if refundRequestsService == nil {
		err := fmt.Errorf("refund requests service not configured")
		l.Error(err, "refund requests service is nil")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	contractorPayoutsService := h.service.Notion.ContractorPayouts
	if contractorPayoutsService == nil {
		err := fmt.Errorf("contractor payouts service not configured")
		l.Error(err, "contractor payouts service is nil")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Query approved refund requests
	l.Debug("querying refund requests with Status=Approved")
	approvedRefunds, err := refundRequestsService.QueryApprovedRefunds(ctx)
	if err != nil {
		l.Error(err, "failed to query approved refund requests")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Info(fmt.Sprintf("found %d refund requests with Status=Approved", len(approvedRefunds)))

	if len(approvedRefunds) == 0 {
		l.Info("no approved refund requests found, returning success with zero counts")
		c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
			"payouts_created":   0,
			"refunds_processed": 0,
			"refunds_skipped":   0,
			"errors":            0,
			"details":           []any{},
			"type":              payoutType,
		}, nil, nil, nil, "ok"))
		return
	}

	// Process each refund
	var (
		payoutsCreated = 0
		refundsSkipped = 0
		errors         = 0
		details        = []map[string]any{}
	)

	for _, refund := range approvedRefunds {
		l.Debug(fmt.Sprintf("processing refund: %s refundID: %s contractor: %s", refund.PageID, refund.RefundID, refund.ContractorPageID))

		detail := map[string]any{
			"refund_page_id":  refund.PageID,
			"refund_id":       refund.RefundID,
			"contractor_id":   refund.ContractorPageID,
			"contractor_name": refund.ContractorName,
			"amount":          refund.Amount,
			"currency":        refund.Currency,
			"reason":          refund.Reason,
			"payout_page_id":  nil,
			"status":          "",
			"error_reason":    nil,
		}

		// Validate contractor
		if refund.ContractorPageID == "" {
			l.Warn(fmt.Sprintf("skipping refund %s: no contractor found", refund.PageID))
			detail["status"] = "skipped"
			detail["error_reason"] = "contractor not found in relation"
			refundsSkipped++
			details = append(details, detail)
			continue
		}

		// Check if payout exists (idempotency)
		l.Debug(fmt.Sprintf("checking if payout exists for refund: %s", refund.PageID))
		exists, existingPayoutID, err := contractorPayoutsService.CheckPayoutExistsByRefundRequest(ctx, refund.PageID)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to check payout existence for refund %s", refund.PageID))
			detail["status"] = "error"
			detail["error_reason"] = "failed to check payout existence"
			errors++
			details = append(details, detail)
			continue
		}

		if exists {
			l.Debug(fmt.Sprintf("payout already exists for refund %s: %s", refund.PageID, existingPayoutID))
			detail["status"] = "skipped"
			detail["error_reason"] = "payout already exists"
			detail["payout_page_id"] = existingPayoutID
			refundsSkipped++
			details = append(details, detail)
			continue
		}

		// Create payout
		// Build payout name from reason
		payoutName := refund.Reason
		if payoutName == "" {
			payoutName = "Refund"
		}
		if refund.Description != "" {
			payoutName = fmt.Sprintf("%s - %s", payoutName, refund.Description)
		}
		l.Debug(fmt.Sprintf("creating payout for refund: %s name: %s", refund.PageID, payoutName))

		payoutInput := notionsvc.CreateRefundPayoutInput{
			Name:             payoutName,
			ContractorPageID: refund.ContractorPageID,
			RefundRequestID:  refund.PageID,
			Amount:           refund.Amount,
			Currency:         refund.Currency,
			Date:             refund.DateRequested,
		}

		payoutPageID, err := contractorPayoutsService.CreateRefundPayout(ctx, payoutInput)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to create payout for refund %s", refund.PageID))
			detail["status"] = "error"
			detail["error_reason"] = "failed to create payout"
			errors++
			details = append(details, detail)
			continue
		}

		l.Info(fmt.Sprintf("created payout: %s for refund: %s", payoutPageID, refund.PageID))

		// NOTE: Do NOT update refund status - leave as Approved
		// Status update is handled separately

		detail["status"] = "created"
		detail["payout_page_id"] = payoutPageID
		payoutsCreated++
		details = append(details, detail)
	}

	// Return response
	l.Info(fmt.Sprintf("processing complete: payouts_created=%d skipped=%d errors=%d", payoutsCreated, refundsSkipped, errors))

	c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
		"payouts_created":   payoutsCreated,
		"refunds_processed": len(approvedRefunds),
		"refunds_skipped":   refundsSkipped,
		"errors":            errors,
		"details":           details,
		"type":              payoutType,
	}, nil, nil, nil, "ok"))
}

// processCommissionPayouts processes pending commission invoice splits
// and creates payout entries of type "Commission"
func (h *handler) processCommissionPayouts(c *gin.Context, l logger.Logger, payoutType string) {
	ctx := c.Request.Context()

	// Get services
	invoiceSplitService := h.service.Notion.InvoiceSplit
	if invoiceSplitService == nil {
		err := fmt.Errorf("invoice split service not configured")
		l.Error(err, "invoice split service is nil")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	contractorPayoutsService := h.service.Notion.ContractorPayouts
	if contractorPayoutsService == nil {
		err := fmt.Errorf("contractor payouts service not configured")
		l.Error(err, "contractor payouts service is nil")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Query pending commission splits
	l.Debug("querying invoice splits with Status=Pending and Type=Commission")
	pendingSplits, err := invoiceSplitService.QueryPendingCommissionSplits(ctx)
	if err != nil {
		l.Error(err, "failed to query pending commission splits")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Info(fmt.Sprintf("found %d pending commission splits", len(pendingSplits)))

	if len(pendingSplits) == 0 {
		l.Info("no pending commission splits found, returning success with zero counts")
		c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
			"payouts_created":   0,
			"splits_processed":  0,
			"splits_skipped":    0,
			"errors":            0,
			"details":           []any{},
			"type":              payoutType,
		}, nil, nil, nil, "ok"))
		return
	}

	// Process each split
	var (
		payoutsCreated = 0
		splitsSkipped  = 0
		errors         = 0
		details        = []map[string]any{}
	)

	for _, split := range pendingSplits {
		l.Debug(fmt.Sprintf("processing split: %s name: %s person: %s", split.PageID, split.Name, split.PersonPageID))

		detail := map[string]any{
			"split_page_id":  split.PageID,
			"split_name":     split.Name,
			"person_id":      split.PersonPageID,
			"amount":         split.Amount,
			"currency":       split.Currency,
			"role":           split.Role,
			"payout_page_id": nil,
			"status":         "",
			"error_reason":   nil,
		}

		// Validate person
		if split.PersonPageID == "" {
			l.Warn(fmt.Sprintf("skipping split %s: no person found", split.PageID))
			detail["status"] = "skipped"
			detail["error_reason"] = "person not found in relation"
			splitsSkipped++
			details = append(details, detail)
			continue
		}

		// Check if payout exists (idempotency)
		l.Debug(fmt.Sprintf("checking if payout exists for split: %s", split.PageID))
		exists, existingPayoutID, err := contractorPayoutsService.CheckPayoutExistsByInvoiceSplit(ctx, split.PageID)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to check payout existence for split %s", split.PageID))
			detail["status"] = "error"
			detail["error_reason"] = "failed to check payout existence"
			errors++
			details = append(details, detail)
			continue
		}

		if exists {
			l.Debug(fmt.Sprintf("payout already exists for split %s: %s", split.PageID, existingPayoutID))
			detail["status"] = "skipped"
			detail["error_reason"] = "payout already exists"
			detail["payout_page_id"] = existingPayoutID
			splitsSkipped++
			details = append(details, detail)
			continue
		}

		// Create payout
		l.Debug(fmt.Sprintf("creating payout for split: %s name: %s", split.PageID, split.Name))

		payoutInput := notionsvc.CreateCommissionPayoutInput{
			Name:             split.Name,
			ContractorPageID: split.PersonPageID,
			InvoiceSplitID:   split.PageID,
			Amount:           split.Amount,
			Currency:         split.Currency,
			Date:             split.Month,
		}

		payoutPageID, err := contractorPayoutsService.CreateCommissionPayout(ctx, payoutInput)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to create payout for split %s", split.PageID))
			detail["status"] = "error"
			detail["error_reason"] = "failed to create payout"
			errors++
			details = append(details, detail)
			continue
		}

		l.Info(fmt.Sprintf("created payout: %s for split: %s", payoutPageID, split.PageID))

		detail["status"] = "created"
		detail["payout_page_id"] = payoutPageID
		payoutsCreated++
		details = append(details, detail)
	}

	// Return response
	l.Info(fmt.Sprintf("processing complete: payouts_created=%d skipped=%d errors=%d", payoutsCreated, splitsSkipped, errors))

	c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
		"payouts_created":  payoutsCreated,
		"splits_processed": len(pendingSplits),
		"splits_skipped":   splitsSkipped,
		"errors":           errors,
		"details":          details,
		"type":             payoutType,
	}, nil, nil, nil, "ok"))
}

// processBonusPayouts processes pending bonus invoice splits
// and creates payout entries of type "Bonus"
func (h *handler) processBonusPayouts(c *gin.Context, l logger.Logger, payoutType string) {
	ctx := c.Request.Context()

	// Get services
	invoiceSplitService := h.service.Notion.InvoiceSplit
	if invoiceSplitService == nil {
		err := fmt.Errorf("invoice split service not configured")
		l.Error(err, "invoice split service is nil")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	contractorPayoutsService := h.service.Notion.ContractorPayouts
	if contractorPayoutsService == nil {
		err := fmt.Errorf("contractor payouts service not configured")
		l.Error(err, "contractor payouts service is nil")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Query pending bonus splits
	l.Debug("querying invoice splits with Status=Pending and Type=Bonus")
	pendingSplits, err := invoiceSplitService.QueryPendingBonusSplits(ctx)
	if err != nil {
		l.Error(err, "failed to query pending bonus splits")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Info(fmt.Sprintf("found %d pending bonus splits", len(pendingSplits)))

	if len(pendingSplits) == 0 {
		l.Info("no pending bonus splits found, returning success with zero counts")
		c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
			"payouts_created":  0,
			"splits_processed": 0,
			"splits_skipped":   0,
			"errors":           0,
			"details":          []any{},
			"type":             payoutType,
		}, nil, nil, nil, "ok"))
		return
	}

	// Process each split
	var (
		payoutsCreated = 0
		splitsSkipped  = 0
		errors         = 0
		details        = []map[string]any{}
	)

	for _, split := range pendingSplits {
		l.Debug(fmt.Sprintf("processing bonus split: %s name: %s person: %s", split.PageID, split.Name, split.PersonPageID))

		detail := map[string]any{
			"split_page_id":  split.PageID,
			"split_name":     split.Name,
			"person_id":      split.PersonPageID,
			"amount":         split.Amount,
			"currency":       split.Currency,
			"role":           split.Role,
			"payout_page_id": nil,
			"status":         "",
			"error_reason":   nil,
		}

		// Validate person
		if split.PersonPageID == "" {
			l.Warn(fmt.Sprintf("skipping bonus split %s: no person found", split.PageID))
			detail["status"] = "skipped"
			detail["error_reason"] = "person not found in relation"
			splitsSkipped++
			details = append(details, detail)
			continue
		}

		// Check if payout exists (idempotency)
		l.Debug(fmt.Sprintf("checking if payout exists for bonus split: %s", split.PageID))
		exists, existingPayoutID, err := contractorPayoutsService.CheckPayoutExistsByInvoiceSplit(ctx, split.PageID)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to check payout existence for bonus split %s", split.PageID))
			detail["status"] = "error"
			detail["error_reason"] = "failed to check payout existence"
			errors++
			details = append(details, detail)
			continue
		}

		if exists {
			l.Debug(fmt.Sprintf("payout already exists for bonus split %s: %s", split.PageID, existingPayoutID))
			detail["status"] = "skipped"
			detail["error_reason"] = "payout already exists"
			detail["payout_page_id"] = existingPayoutID
			splitsSkipped++
			details = append(details, detail)
			continue
		}

		// Create payout
		l.Debug(fmt.Sprintf("creating payout for bonus split: %s name: %s", split.PageID, split.Name))

		payoutInput := notionsvc.CreateBonusPayoutInput{
			Name:             split.Name,
			ContractorPageID: split.PersonPageID,
			InvoiceSplitID:   split.PageID,
			Amount:           split.Amount,
			Currency:         split.Currency,
			Date:             split.Month,
		}

		payoutPageID, err := contractorPayoutsService.CreateBonusPayout(ctx, payoutInput)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to create payout for bonus split %s", split.PageID))
			detail["status"] = "error"
			detail["error_reason"] = "failed to create payout"
			errors++
			details = append(details, detail)
			continue
		}

		l.Info(fmt.Sprintf("created bonus payout: %s for split: %s", payoutPageID, split.PageID))

		detail["status"] = "created"
		detail["payout_page_id"] = payoutPageID
		payoutsCreated++
		details = append(details, detail)
	}

	// Return response
	l.Info(fmt.Sprintf("bonus processing complete: payouts_created=%d skipped=%d errors=%d", payoutsCreated, splitsSkipped, errors))

	c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
		"payouts_created":  payoutsCreated,
		"splits_processed": len(pendingSplits),
		"splits_skipped":   splitsSkipped,
		"errors":           errors,
		"details":          details,
		"type":             payoutType,
	}, nil, nil, nil, "ok"))
}
