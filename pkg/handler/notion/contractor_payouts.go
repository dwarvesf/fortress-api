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
		err := fmt.Errorf("payout type 'commission' not implemented yet")
		l.Error(err, "commission payout type not implemented")
		c.JSON(http.StatusNotImplemented, view.CreateResponse[any](nil, nil, err, nil, ""))
	case "refund":
		err := fmt.Errorf("payout type 'refund' not implemented yet")
		l.Error(err, "refund payout type not implemented")
		c.JSON(http.StatusNotImplemented, view.CreateResponse[any](nil, nil, err, nil, ""))
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
