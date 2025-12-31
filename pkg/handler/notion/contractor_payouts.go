package notion

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	notionsvc "github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

// CreateContractorPayouts godoc
// @Summary Create contractor payouts from new contractor fees
// @Description Processes contractor fees with Payment Status=New and creates payout entries
// @Tags Cronjobs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} view.Response
// @Failure 500 {object} view.Response
// @Router /cronjobs/contractor-payouts [post]
func (h *handler) CreateContractorPayouts(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "Notion",
		"method":  "CreateContractorPayouts",
	})
	ctx := c.Request.Context()

	l.Info("starting CreateContractorPayouts cronjob")

	// Step 1: Get services
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

	// Step 2: Query new fees (Payment Status=New)
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
		}, nil, nil, nil, "ok"))
		return
	}

	// Step 3: Process each fee
	var (
		payoutsCreated = 0
		feesSkipped    = 0
		errors         = 0
		details        = []map[string]any{}
	)

	for _, fee := range newFees {
		l.Debug(fmt.Sprintf("processing fee: %s contractor: %s", fee.PageID, fee.ContractorName))

		detail := map[string]any{
			"fee_page_id":       fee.PageID,
			"contractor_name":   fee.ContractorName,
			"contractor_id":     fee.ContractorPageID,
			"amount":            fee.TotalAmount,
			"month":             fee.Month,
			"payout_page_id":    nil,
			"status":            "",
			"reason":            nil,
		}

		// Step 3a: Validate contractor
		if fee.ContractorPageID == "" {
			l.Warn(fmt.Sprintf("skipping fee %s: no contractor found", fee.PageID))
			detail["status"] = "skipped"
			detail["reason"] = "contractor not found in relation"
			feesSkipped++
			details = append(details, detail)
			continue
		}

		// Step 3b: Check if payout exists (idempotency)
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

		// Step 3c: Create payout
		payoutName := fmt.Sprintf("%s - %s", fee.ContractorName, fee.Month)
		l.Debug(fmt.Sprintf("creating payout for fee: %s name: %s", fee.PageID, payoutName))

		payoutInput := notionsvc.CreatePayoutInput{
			Name:             payoutName,
			ContractorPageID: fee.ContractorPageID,
			ContractorFeeID:  fee.PageID,
			Amount:           fee.TotalAmount,
			Month:            fee.Month,
			Date:             fee.Date,
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

		// Step 3d: Update fee Payment Status to "Pending"
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

	// Step 4: Return response
	l.Info(fmt.Sprintf("processing complete: payouts_created=%d skipped=%d errors=%d", payoutsCreated, feesSkipped, errors))

	c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
		"payouts_created": payoutsCreated,
		"fees_processed":  len(newFees),
		"fees_skipped":    feesSkipped,
		"errors":          errors,
		"details":         details,
	}, nil, nil, nil, "ok"))
}
