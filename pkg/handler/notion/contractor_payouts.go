package notion

import (
	"fmt"
	"math"
	"net/http"
	"strings"
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

// countWorkingDays counts working days (Mon-Fri) between two dates inclusive
func countWorkingDays(start, end time.Time) int {
	if start.After(end) {
		return 0
	}

	count := 0
	current := start
	for !current.After(end) {
		weekday := current.Weekday()
		if weekday != time.Saturday && weekday != time.Sunday {
			count++
		}
		current = current.AddDate(0, 0, 1)
	}
	return count
}

// calculateMonthlyFixedAmount calculates the prorated monthly fixed amount based on working days
// startDate: contractor's start date from Contractor Rates
// orderDate: the order date (used to determine the month)
// monthlyFixed: the full monthly fixed amount
func calculateMonthlyFixedAmount(startDate *time.Time, orderDate time.Time, monthlyFixed float64, l logger.Logger) float64 {
	// Get first and last day of the order month
	firstOfMonth := time.Date(orderDate.Year(), orderDate.Month(), 1, 0, 0, 0, 0, orderDate.Location())
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	// Calculate total working days in the month
	totalWorkingDays := countWorkingDays(firstOfMonth, lastOfMonth)
	l.Debug(fmt.Sprintf("calculateMonthlyFixedAmount: month=%s totalWorkingDays=%d", firstOfMonth.Format("2006-01"), totalWorkingDays))

	if totalWorkingDays == 0 {
		l.Debug("calculateMonthlyFixedAmount: no working days in month, returning 0")
		return 0
	}

	// Determine the effective start date for this month
	effectiveStart := firstOfMonth
	if startDate != nil && startDate.After(firstOfMonth) && !startDate.After(lastOfMonth) {
		// Start date is within this month
		effectiveStart = *startDate
		l.Debug(fmt.Sprintf("calculateMonthlyFixedAmount: startDate=%s is within month, using as effectiveStart", startDate.Format("2006-01-02")))
	} else if startDate != nil && startDate.After(lastOfMonth) {
		// Start date is after this month - no working days
		l.Debug(fmt.Sprintf("calculateMonthlyFixedAmount: startDate=%s is after month, returning 0", startDate.Format("2006-01-02")))
		return 0
	}

	// Calculate actual working days from effective start to end of month
	actualWorkingDays := countWorkingDays(effectiveStart, lastOfMonth)
	l.Debug(fmt.Sprintf("calculateMonthlyFixedAmount: effectiveStart=%s actualWorkingDays=%d", effectiveStart.Format("2006-01-02"), actualWorkingDays))

	// Prorate the amount
	amount := monthlyFixed * float64(actualWorkingDays) / float64(totalWorkingDays)
	l.Debug(fmt.Sprintf("calculateMonthlyFixedAmount: monthlyFixed=%.2f * (%d/%d) = %.2f", monthlyFixed, actualWorkingDays, totalWorkingDays, amount))

	// Round up to nearest thousand
	roundedAmount := math.Ceil(amount/1000) * 1000
	l.Debug(fmt.Sprintf("calculateMonthlyFixedAmount: rounded up to nearest thousand: %.2f -> %.2f", amount, roundedAmount))

	return roundedAmount
}

var (
	PayoutType = map[string]string{
		"contractor_payroll": "Service Fee",
		"invoice_split":      "Invoice Split",
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

	// Get optional month filter (YYYY-MM format) - only used for contractor_payroll
	monthFilter := c.Query("month")
	if monthFilter != "" {
		if _, err := time.Parse("2006-01", monthFilter); err != nil {
			l.Debug(fmt.Sprintf("invalid month parameter: %s", monthFilter))
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, fmt.Errorf("invalid month format, expected YYYY-MM"), nil, ""))
			return
		}
	}

	l.Debug(fmt.Sprintf("payout type key: %s, value: %s, contractor: %s, pay_day: %d, month: %s", payoutTypeKey, payoutType, contractorFilter, payDayFilter, monthFilter))
	l.Info("starting CreateContractorPayouts cronjob")

	// Process based on payout type
	switch payoutTypeKey {
	case "contractor_payroll":
		h.processContractorPayrollPayouts(c, l, payoutType, contractorFilter, payDayFilter, monthFilter)
	case "invoice_split":
		h.processInvoiceSplitPayouts(c, l)
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
// monthFilter: optional filter by month (YYYY-MM), empty means no filter
func (h *handler) processContractorPayrollPayouts(c *gin.Context, l logger.Logger, payoutType string, contractorFilter string, payDayFilter int, monthFilter string) {
	ctx := c.Request.Context()

	l.Debug(fmt.Sprintf("processContractorPayrollPayouts: contractorFilter=%s payDayFilter=%d monthFilter=%s", contractorFilter, payDayFilter, monthFilter))

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

	// Query approved orders (Type=Order, Status=Approved, optional month filter)
	l.Debug(fmt.Sprintf("querying Task Order Log with Type=Order, Status=Approved, month=%s", monthFilter))
	approvedOrders, err := taskOrderLogService.QueryApprovedOrders(ctx, monthFilter)
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
		detail        map[string]any
		payoutCreated bool
		skipped       bool
		hasError      bool
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
					// Prorate monthly fixed based on actual working days from Start Date
					amount = calculateMonthlyFixedAmount(rate.StartDate, order.Date, rate.MonthlyFixed, l)
					l.Debug(fmt.Sprintf("[worker-%d] using monthly fixed rate: %.2f (prorated from %.2f)", workerID, amount, rate.MonthlyFixed))
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

// processInvoiceSplitPayouts processes pending invoice splits (Commission, Bonus, Fee)
// and creates payout entries of type "Commission"
func (h *handler) processInvoiceSplitPayouts(c *gin.Context, l logger.Logger) {
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

	// Query pending invoice splits (Commission, Bonus, Fee)
	l.Debug("querying invoice splits with Status=Pending and Type in (Commission, Bonus, Fee)")
	pendingSplits, err := invoiceSplitService.QueryPendingInvoiceSplits(ctx)
	if err != nil {
		l.Error(err, "failed to query pending invoice splits")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Info(fmt.Sprintf("found %d pending invoice splits", len(pendingSplits)))

	if len(pendingSplits) == 0 {
		l.Info("no pending invoice splits found, returning success with zero counts")
		c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
			"payouts_created":  0,
			"splits_processed": 0,
			"splits_skipped":   0,
			"errors":           0,
			"details":          []any{},
			"type":             "Commission",
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
		l.Debug(fmt.Sprintf("processing invoice split: %s name: %s type: %s person: %s", split.PageID, split.Name, split.Type, split.PersonPageID))

		detail := map[string]any{
			"split_page_id":  split.PageID,
			"split_name":     split.Name,
			"split_type":     split.Type,
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
			l.Warn(fmt.Sprintf("skipping invoice split %s: no person found", split.PageID))
			detail["status"] = "skipped"
			detail["error_reason"] = "person not found in relation"
			splitsSkipped++
			details = append(details, detail)
			continue
		}

		// Check if payout exists (idempotency)
		l.Debug(fmt.Sprintf("checking if payout exists for invoice split: %s", split.PageID))
		exists, existingPayoutID, err := contractorPayoutsService.CheckPayoutExistsByInvoiceSplit(ctx, split.PageID)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to check payout existence for invoice split %s", split.PageID))
			detail["status"] = "error"
			detail["error_reason"] = "failed to check payout existence"
			errors++
			details = append(details, detail)
			continue
		}

		if exists {
			l.Debug(fmt.Sprintf("payout already exists for invoice split %s: %s", split.PageID, existingPayoutID))
			detail["status"] = "skipped"
			detail["error_reason"] = "payout already exists"
			detail["payout_page_id"] = existingPayoutID
			splitsSkipped++
			details = append(details, detail)
			continue
		}

		// Create payout
		l.Debug(fmt.Sprintf("creating payout for invoice split: %s name: %s notes: %s", split.PageID, split.Name, split.Description))

		payoutInput := notionsvc.CreateCommissionPayoutInput{
			Name:             split.Name,
			ContractorPageID: split.PersonPageID,
			InvoiceSplitID:   split.PageID,
			Amount:           split.Amount,
			Currency:         split.Currency,
			Date:             split.Month,
			Description:      split.Description, // Notes field reads from Notion's Description formula
		}

		payoutPageID, err := contractorPayoutsService.CreateCommissionPayout(ctx, payoutInput)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to create payout for invoice split %s", split.PageID))
			detail["status"] = "error"
			detail["error_reason"] = "failed to create payout"
			errors++
			details = append(details, detail)
			continue
		}

		l.Info(fmt.Sprintf("created payout: %s for invoice split: %s", payoutPageID, split.PageID))

		detail["status"] = "created"
		detail["payout_page_id"] = payoutPageID
		payoutsCreated++
		details = append(details, detail)
	}

	// Return response
	l.Info(fmt.Sprintf("invoice split processing complete: payouts_created=%d skipped=%d errors=%d", payoutsCreated, splitsSkipped, errors))

	c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
		"payouts_created":  payoutsCreated,
		"splits_processed": len(pendingSplits),
		"splits_skipped":   splitsSkipped,
		"errors":           errors,
		"details":          details,
		"type":             "Commission",
	}, nil, nil, nil, "ok"))
}

// SyncPayouts godoc
// @Summary Sync payout fields from their linked source records
// @Description Syncs payout fields (description, amount) from linked Invoice Split records
// @Tags Notion
// @Accept json
// @Produce json
// @Param source query string true "Sync source: split (from Invoice Split)"
// @Param fields query string false "Comma-separated fields to sync (default: description)"
// @Param id query string false "Filter by payout short ID (e.g., JGSSC from Auto Name suffix)"
// @Security BearerAuth
// @Success 200 {object} view.Response
// @Failure 400 {object} view.Response
// @Failure 500 {object} view.Response
// @Router /notion/contractor-payouts/sync [post]
func (h *handler) SyncPayouts(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "Notion",
		"method":  "SyncPayouts",
	})

	// Get required source parameter
	source := c.Query("source")
	if source == "" {
		err := fmt.Errorf("missing required parameter: source")
		l.Error(err, "source parameter is required")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Get optional fields parameter (default: description)
	fieldsParam := c.Query("fields")
	if fieldsParam == "" {
		fieldsParam = "description"
	}

	// Get optional id filter (short ID from Auto Name suffix, e.g., "JGSSC")
	idFilter := c.Query("id")

	l.Debug(fmt.Sprintf("sync source=%s fields=%s id=%s", source, fieldsParam, idFilter))

	// Process based on source
	switch source {
	case "split":
		h.syncPayoutsFromSplit(c, l, fieldsParam, idFilter)
	default:
		err := fmt.Errorf("invalid source: %s (supported: split)", source)
		l.Error(err, "invalid source parameter")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
	}
}

// syncPayoutsFromSplit syncs payout fields from linked Invoice Split records
// idFilter: optional filter by payout short ID (suffix of Auto Name, e.g., "JGSSC")
func (h *handler) syncPayoutsFromSplit(c *gin.Context, l logger.Logger, fieldsParam string, idFilter string) {
	ctx := c.Request.Context()

	// Parse fields to sync
	fieldsToSync := make(map[string]bool)
	for _, field := range splitFields(fieldsParam) {
		fieldsToSync[field] = true
	}

	l.Debug(fmt.Sprintf("fields to sync: %v idFilter: %s", fieldsToSync, idFilter))

	// Validate fields
	validFields := map[string]bool{"description": true, "amount": true}
	for field := range fieldsToSync {
		if !validFields[field] {
			err := fmt.Errorf("invalid field: %s (supported: description, amount)", field)
			l.Error(err, "invalid field parameter")
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
	}

	// Get services
	contractorPayoutsService := h.service.Notion.ContractorPayouts
	if contractorPayoutsService == nil {
		err := fmt.Errorf("contractor payouts service not configured")
		l.Error(err, "contractor payouts service is nil")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	invoiceSplitService := h.service.Notion.InvoiceSplit
	if invoiceSplitService == nil {
		err := fmt.Errorf("invoice split service not configured")
		l.Error(err, "invoice split service is nil")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Query payouts with Invoice Split relation
	l.Debug("querying payouts with Invoice Split relation")
	payouts, err := contractorPayoutsService.QueryPayoutsWithInvoiceSplit(ctx)
	if err != nil {
		l.Error(err, "failed to query payouts with Invoice Split")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Info(fmt.Sprintf("found %d payouts with Invoice Split relation", len(payouts)))

	// Filter by short ID (suffix of Auto Name) if provided
	if idFilter != "" {
		l.Debug(fmt.Sprintf("filtering payouts by short ID: %s", idFilter))
		var filteredPayouts []notionsvc.PayoutEntry
		for _, p := range payouts {
			// Extract short ID from Auto Name (e.g., "JGSSC" from "PYT :: 202511 :: [...] :: JGSSC")
			shortID := extractShortID(p.Name)
			if shortID == idFilter {
				filteredPayouts = append(filteredPayouts, p)
			}
		}
		l.Debug(fmt.Sprintf("filtered from %d to %d payouts", len(payouts), len(filteredPayouts)))
		payouts = filteredPayouts
	}

	if len(payouts) == 0 {
		c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
			"total_processed": 0,
			"updated":         0,
			"skipped":         0,
			"errors":          0,
			"fields_synced":   getFieldsList(fieldsToSync),
			"details":         []any{},
		}, nil, nil, nil, "ok"))
		return
	}

	// Process each payout
	var (
		updated = 0
		skipped = 0
		errors  = 0
		details = []map[string]any{}
	)

	for _, payout := range payouts {
		l.Debug(fmt.Sprintf("processing payout pageID=%s splitID=%s", payout.PageID, payout.InvoiceSplitID))

		detail := map[string]any{
			"payout_id": payout.PageID,
			"split_id":  payout.InvoiceSplitID,
			"status":    "",
			"changes":   map[string]any{},
		}

		// Fetch Invoice Split data
		splitData, err := invoiceSplitService.GetInvoiceSplitSyncData(ctx, payout.InvoiceSplitID)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to fetch Invoice Split data for payout=%s split=%s", payout.PageID, payout.InvoiceSplitID))
			detail["status"] = "error"
			detail["error"] = fmt.Sprintf("failed to fetch split data: %v", err)
			errors++
			details = append(details, detail)
			continue
		}

		// Compare and prepare updates
		updates := notionsvc.PayoutFieldUpdates{}
		changes := map[string]any{}
		hasChanges := false

		// Sync Description if requested
		if fieldsToSync["description"] {
			if payout.Description != splitData.Description {
				updates.Description = &splitData.Description
				changes["description"] = map[string]any{
					"old": payout.Description,
					"new": splitData.Description,
				}
				hasChanges = true
				l.Debug(fmt.Sprintf("payout %s: description change from '%s' to '%s'", payout.PageID, payout.Description, splitData.Description))
			}
		}

		// Sync Amount if requested (for future use)
		if fieldsToSync["amount"] {
			if payout.Amount != splitData.Amount {
				updates.Amount = &splitData.Amount
				changes["amount"] = map[string]any{
					"old": payout.Amount,
					"new": splitData.Amount,
				}
				hasChanges = true
				l.Debug(fmt.Sprintf("payout %s: amount change from %.2f to %.2f", payout.PageID, payout.Amount, splitData.Amount))
			}
		}

		detail["changes"] = changes

		// Skip if no changes
		if !hasChanges {
			l.Debug(fmt.Sprintf("payout %s: no changes needed", payout.PageID))
			detail["status"] = "skipped"
			skipped++
			details = append(details, detail)
			continue
		}

		// Update payout
		err = contractorPayoutsService.UpdatePayoutFields(ctx, payout.PageID, updates)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to update payout %s", payout.PageID))
			detail["status"] = "error"
			detail["error"] = fmt.Sprintf("failed to update: %v", err)
			errors++
			details = append(details, detail)
			continue
		}

		l.Info(fmt.Sprintf("updated payout %s", payout.PageID))
		detail["status"] = "updated"
		updated++
		details = append(details, detail)
	}

	l.Info(fmt.Sprintf("sync complete: updated=%d skipped=%d errors=%d", updated, skipped, errors))

	c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
		"total_processed": len(payouts),
		"updated":         updated,
		"skipped":         skipped,
		"errors":          errors,
		"fields_synced":   getFieldsList(fieldsToSync),
		"details":         details,
	}, nil, nil, nil, "ok"))
}

// extractShortID extracts the short ID suffix from Auto Name formula
// e.g., "PYT :: 202512 :: [SPL :: RENAISS :: ooohminh] :: 79LUH" -> "79LUH"
func extractShortID(name string) string {
	parts := strings.Split(name, " :: ")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// splitFields splits a comma-separated string into a slice of trimmed strings
func splitFields(s string) []string {
	var result []string
	for _, part := range strings.Split(s, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// getFieldsList returns a slice of field names from the map
func getFieldsList(fieldsMap map[string]bool) []string {
	var result []string
	for field := range fieldsMap {
		result = append(result, field)
	}
	return result
}
