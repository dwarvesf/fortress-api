package notion

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	notionsvc "github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

// CreateContractorFees godoc
// @Summary Create contractor fees from approved task orders
// @Description Processes approved task order logs and creates contractor fee entries with matching contractor rates
// @Tags Cronjobs
// @Accept json
// @Produce json
// @Param discord query string false "Filter by contractor Discord username (for testing specific contractor)"
// @Security BearerAuth
// @Success 200 {object} view.Response
// @Failure 500 {object} view.Response
// @Router /cronjobs/contractor-fees [post]
func (h *handler) CreateContractorFees(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "Notion",
		"method":  "CreateContractorFees",
	})
	ctx := c.Request.Context()

	// Get optional discord filter from query parameter
	discordFilter := c.Query("discord")
	if discordFilter != "" {
		l.Debug(fmt.Sprintf("discord filter applied: %s", discordFilter))
	}

	l.Info("starting CreateContractorFees cronjob")

	// Step 1: Get services
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

	contractorFeesService := h.service.Notion.ContractorFees
	if contractorFeesService == nil {
		err := fmt.Errorf("contractor fees service not configured")
		l.Error(err, "contractor fees service is nil")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Step 2: Query approved orders
	l.Debug("querying approved orders")
	approvedOrders, err := taskOrderLogService.QueryApprovedOrders(ctx)
	if err != nil {
		l.Error(err, "failed to query approved orders")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Info(fmt.Sprintf("found %d approved orders", len(approvedOrders)))

	// Apply discord filter if provided
	if discordFilter != "" {
		var filteredOrders []*notionsvc.ApprovedOrderData
		for _, order := range approvedOrders {
			if order.ContractorDiscord == discordFilter {
				filteredOrders = append(filteredOrders, order)
			}
		}
		l.Debug(fmt.Sprintf("filtered orders by discord=%s: %d -> %d", discordFilter, len(approvedOrders), len(filteredOrders)))
		approvedOrders = filteredOrders
	}

	if len(approvedOrders) == 0 {
		l.Info("no approved orders found, returning success with zero counts")
		c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
			"contractor_fees_created": 0,
			"orders_processed":        0,
			"orders_skipped":          0,
			"errors":                  0,
			"details":                 []any{},
			"discord_filter":          discordFilter,
		}, nil, nil, nil, "ok"))
		return
	}

	// Step 3: Process each order
	var (
		feesCreated   = 0
		ordersSkipped = 0
		errors        = 0
		details       = []map[string]any{}
	)

	for _, order := range approvedOrders {
		l.Debug(fmt.Sprintf("processing order: %s contractor: %s", order.PageID, order.ContractorName))

		detail := map[string]any{
			"order_page_id":          order.PageID,
			"contractor_name":        order.ContractorName,
			"contractor_discord":     order.ContractorDiscord,
			"contractor_page_id":     order.ContractorPageID,
			"contractor_fee_page_id": nil,
			"status":                 "",
			"reason":                 nil,
		}

		// Step 3a: Validate contractor
		if order.ContractorPageID == "" {
			l.Warn(fmt.Sprintf("skipping order %s: no contractor found", order.PageID))
			detail["status"] = "skipped"
			detail["reason"] = "contractor not found in rollup"
			ordersSkipped++
			details = append(details, detail)
			continue
		}

		// Step 3b: Find matching rate
		l.Debug(fmt.Sprintf("finding active rate for contractor: %s date: %s", order.ContractorPageID, order.Date.Format("2006-01-02")))
		rate, err := contractorRatesService.FindActiveRateByContractor(ctx, order.ContractorPageID, order.Date)
		if err != nil {
			l.Error(err, fmt.Sprintf("no active rate for contractor %s at date %s", order.ContractorPageID, order.Date.Format("2006-01-02")))
			detail["status"] = "error"
			detail["reason"] = fmt.Sprintf("no active contractor rate found for date %s", order.Date.Format("2006-01-02"))
			errors++
			details = append(details, detail)
			continue
		}

		l.Debug(fmt.Sprintf("found rate: %s for contractor: %s", rate.PageID, order.ContractorName))

		// Step 3c: Check if fee exists (idempotency)
		l.Debug(fmt.Sprintf("checking if fee exists for order: %s", order.PageID))
		exists, existingFeeID, err := contractorFeesService.CheckFeeExistsByTaskOrder(ctx, order.PageID)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to check fee existence for order %s", order.PageID))
			detail["status"] = "error"
			detail["reason"] = "failed to check fee existence"
			errors++
			details = append(details, detail)
			continue
		}

		if exists {
			l.Debug(fmt.Sprintf("fee already exists for order %s: %s", order.PageID, existingFeeID))
			detail["status"] = "skipped"
			detail["reason"] = "contractor fee already exists"
			detail["contractor_fee_page_id"] = existingFeeID
			ordersSkipped++
			details = append(details, detail)
			continue
		}

		// Step 3d: Create fee
		l.Debug(fmt.Sprintf("creating fee for order: %s with rate: %s", order.PageID, rate.PageID))
		feePageID, err := contractorFeesService.CreateContractorFee(ctx, order.PageID, rate.PageID)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to create fee for order %s", order.PageID))
			detail["status"] = "error"
			detail["reason"] = "failed to create contractor fee"
			errors++
			details = append(details, detail)
			continue
		}

		l.Info(fmt.Sprintf("created contractor fee: %s for order: %s", feePageID, order.PageID))

		// Step 3e: Update order status to Completed
		l.Debug(fmt.Sprintf("updating order %s status to Completed", order.PageID))
		err = taskOrderLogService.UpdateOrderStatus(ctx, order.PageID, "Completed")
		if err != nil {
			// Log error but don't fail - fee is already created
			l.Error(err, fmt.Sprintf("failed to update order status: %s (fee created: %s)", order.PageID, feePageID))
		} else {
			l.Debug(fmt.Sprintf("updated order %s status to Completed", order.PageID))
		}

		detail["status"] = "created"
		detail["contractor_fee_page_id"] = feePageID
		feesCreated++
		details = append(details, detail)
	}

	// Step 4: Return response
	l.Info(fmt.Sprintf("processing complete: fees_created=%d skipped=%d errors=%d", feesCreated, ordersSkipped, errors))

	c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
		"contractor_fees_created": feesCreated,
		"orders_processed":        len(approvedOrders),
		"orders_skipped":          ordersSkipped,
		"errors":                  errors,
		"details":                 details,
		"discord_filter":          discordFilter,
	}, nil, nil, nil, "ok"))
}
