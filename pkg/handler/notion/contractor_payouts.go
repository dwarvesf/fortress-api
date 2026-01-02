package notion

import (
	"fmt"
	"net/http"
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
		"contractor_payroll": "Contractor Payroll",
		"bonus":              "Bonus",
		"commission":         "Commission",
		"refund":             "Refund",
	}
)

// CreateContractorPayouts godoc
// @Summary Create contractor payouts from new contractor fees
// @Description Processes contractor fees with Payment Status=New and creates payout entries
// @Tags Cronjobs
// @Accept json
// @Produce json
// @Param type query string false "Payout type (default: contractor_payroll)"
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

	l.Debug(fmt.Sprintf("payout type key: %s, value: %s", payoutTypeKey, payoutType))
	l.Info("starting CreateContractorPayouts cronjob")

	// Process based on payout type
	switch payoutTypeKey {
	case "contractor_payroll":
		h.processContractorPayrollPayouts(c, l, payoutType)
	case "bonus":
		err := fmt.Errorf("payout type 'bonus' not implemented yet")
		l.Error(err, "bonus payout type not implemented")
		c.JSON(http.StatusNotImplemented, view.CreateResponse[any](nil, nil, err, nil, ""))
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

// processContractorPayrollPayouts processes contractor fees with Payment Status=New
// and creates payout entries of type "Contractor Payroll"
func (h *handler) processContractorPayrollPayouts(c *gin.Context, l logger.Logger, payoutType string) {
	ctx := c.Request.Context()

	// Get services
	contractorFeesService := h.service.Notion.ContractorFees
	if contractorFeesService == nil {
		err := fmt.Errorf("contractor fees service not configured")
		l.Error(err, "contractor fees service is nil")
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

	// Query new fees (Payment Status=New)
	l.Debug("querying contractor fees with Payment Status=New")
	newFees, err := contractorFeesService.QueryNewFees(ctx)
	if err != nil {
		l.Error(err, "failed to query new contractor fees")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Info(fmt.Sprintf("found %d contractor fees with Payment Status=New", len(newFees)))

	if len(newFees) == 0 {
		l.Info("no new contractor fees found, returning success with zero counts")
		c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
			"payouts_created": 0,
			"fees_processed":  0,
			"fees_skipped":    0,
			"errors":          0,
			"details":         []any{},
			"type":            payoutType,
		}, nil, nil, nil, "ok"))
		return
	}

	// Process each fee
	var (
		payoutsCreated = 0
		feesSkipped    = 0
		errors         = 0
		details        = []map[string]any{}
	)

	for _, fee := range newFees {
		l.Debug(fmt.Sprintf("processing fee: %s contractor: %s", fee.PageID, fee.ContractorName))

		detail := map[string]any{
			"fee_page_id":     fee.PageID,
			"contractor_name": fee.ContractorName,
			"contractor_id":   fee.ContractorPageID,
			"amount":          fee.TotalAmount,
			"month":           fee.Month,
			"payout_page_id":  nil,
			"status":          "",
			"reason":          nil,
		}

		// Validate contractor
		if fee.ContractorPageID == "" {
			l.Warn(fmt.Sprintf("skipping fee %s: no contractor found", fee.PageID))
			detail["status"] = "skipped"
			detail["reason"] = "contractor not found in relation"
			feesSkipped++
			details = append(details, detail)
			continue
		}

		// Check if payout exists (idempotency)
		l.Debug(fmt.Sprintf("checking if payout exists for fee: %s", fee.PageID))
		exists, existingPayoutID, err := contractorPayoutsService.CheckPayoutExistsByContractorFee(ctx, fee.PageID)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to check payout existence for fee %s", fee.PageID))
			detail["status"] = "error"
			detail["reason"] = "failed to check payout existence"
			errors++
			details = append(details, detail)
			continue
		}

		if exists {
			l.Debug(fmt.Sprintf("payout already exists for fee %s: %s", fee.PageID, existingPayoutID))
			detail["status"] = "skipped"
			detail["reason"] = "payout already exists"
			detail["payout_page_id"] = existingPayoutID
			feesSkipped++
			details = append(details, detail)
			continue
		}

		// Create payout
		// Format month from YYYY-MM to "Month, Year" (e.g., "2025-01" -> "January, 2025")
		payoutName := fmt.Sprintf("Development work on %s", formatMonthYear(fee.Month))
		l.Debug(fmt.Sprintf("creating payout for fee: %s name: %s", fee.PageID, payoutName))

		payoutInput := notionsvc.CreatePayoutInput{
			Name:             payoutName,
			ContractorPageID: fee.ContractorPageID,
			ContractorFeeID:  fee.PageID,
			Amount:           fee.TotalAmount,
			Currency:         fee.Currency,
			Month:            fee.Month,
			Date:             fee.Date,
			Type:             payoutType,
		}

		payoutPageID, err := contractorPayoutsService.CreatePayout(ctx, payoutInput)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to create payout for fee %s", fee.PageID))
			detail["status"] = "error"
			detail["reason"] = "failed to create payout"
			errors++
			details = append(details, detail)
			continue
		}

		l.Info(fmt.Sprintf("created payout: %s for fee: %s", payoutPageID, fee.PageID))

		// Update fee Payment Status to "Pending"
		l.Debug(fmt.Sprintf("updating fee %s payment status to Pending", fee.PageID))
		err = contractorFeesService.UpdatePaymentStatus(ctx, fee.PageID, "Pending")
		if err != nil {
			// Log error but don't fail - payout is already created
			l.Error(err, fmt.Sprintf("failed to update fee payment status: %s (payout created: %s)", fee.PageID, payoutPageID))
		} else {
			l.Debug(fmt.Sprintf("updated fee %s payment status to Pending", fee.PageID))
		}

		detail["status"] = "created"
		detail["payout_page_id"] = payoutPageID
		payoutsCreated++
		details = append(details, detail)
	}

	// Return response
	l.Info(fmt.Sprintf("processing complete: payouts_created=%d skipped=%d errors=%d", payoutsCreated, feesSkipped, errors))

	c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
		"payouts_created": payoutsCreated,
		"fees_processed":  len(newFees),
		"fees_skipped":    feesSkipped,
		"errors":          errors,
		"details":         details,
		"type":            payoutType,
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

		// Derive month from DateRequested (YYYY-MM-DD -> YYYY-MM)
		month := ""
		if len(refund.DateRequested) >= 7 {
			month = refund.DateRequested[:7]
		}

		payoutInput := notionsvc.CreateRefundPayoutInput{
			Name:             payoutName,
			ContractorPageID: refund.ContractorPageID,
			RefundRequestID:  refund.PageID,
			Amount:           refund.Amount,
			Currency:         refund.Currency,
			Month:            month,
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
