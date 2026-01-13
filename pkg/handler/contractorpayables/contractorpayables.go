package contractorpayables

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	ctrlcontractorpayables "github.com/dwarvesf/fortress-api/pkg/controller/contractorpayables"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	controller ctrlcontractorpayables.IController
	logger     logger.Logger
	config     *config.Config
}

// New creates a new contractor payables handler
func New(controller ctrlcontractorpayables.IController, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		controller: controller,
		logger:     logger,
		config:     cfg,
	}
}

// PreviewCommit godoc
// @Summary Preview contractor payables commit
// @Description Preview which contractor payables would be committed for a given month and batch (PayDay)
// @Tags Contractor Payables
// @Accept json
// @Produce json
// @Param month query string true "Month in YYYY-MM format (e.g., 2025-01)"
// @Param batch query int true "PayDay batch (1 or 15)"
// @Success 200 {object} view.Response{data=PreviewCommitResponse}
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /api/v1/contractor-payables/preview-commit [get]
func (h *handler) PreviewCommit(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "contractorpayables",
		"method":  "PreviewCommit",
	})

	var req PreviewCommitRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		l.Error(err, "failed to bind query parameters")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Validate month format (YYYY-MM)
	if !isValidMonthFormat(req.Month) {
		l.Error(nil, "invalid month format")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil,
			nil, nil, "Invalid month format. Use YYYY-MM (e.g., 2025-01)"))
		return
	}

	l.Debug("calling controller.PreviewCommit")

	result, err := h.controller.PreviewCommit(c.Request.Context(), req.Month, req.Batch, req.Contractor)
	if err != nil {
		l.Error(err, "failed to preview commit")

		// Check if no payables found (not an error, return 200 with count=0)
		if strings.Contains(err.Error(), "no pending payables") {
			c.JSON(http.StatusOK, view.CreateResponse[any](
				ctrlcontractorpayables.PreviewCommitResponse{
					Month:       req.Month,
					Batch:       req.Batch,
					Count:       0,
					TotalAmount: 0,
					Contractors: []ctrlcontractorpayables.ContractorPreview{},
				},
				nil, nil, nil, ""))
			return
		}

		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](result, nil, nil, nil, ""))
}

// Commit godoc
// @Summary Commit contractor payables to Paid status
// @Description Execute the payout commit, updating all related records (Payables, Payouts, Invoice Splits, Refunds)
// @Tags Contractor Payables
// @Accept json
// @Produce json
// @Param request body CommitRequest true "Commit request"
// @Success 200 {object} view.Response{data=CommitResponse}
// @Success 207 {object} view.Response{data=CommitResponse} "Partial success (some updates failed)"
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /api/v1/contractor-payables/commit [post]
func (h *handler) Commit(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "contractorpayables",
		"method":  "Commit",
	})

	var req CommitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		l.Error(err, "failed to bind request body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Validate month format (YYYY-MM)
	if !isValidMonthFormat(req.Month) {
		l.Error(nil, "invalid month format")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil,
			nil, nil, "Invalid month format. Use YYYY-MM (e.g., 2025-01)"))
		return
	}

	l.Debug("calling controller.CommitPayables")

	result, err := h.controller.CommitPayables(c.Request.Context(), req.Month, req.Batch, req.Contractor)
	if err != nil {
		l.Error(err, "failed to commit payables")

		// Check if no payables found
		if strings.Contains(err.Error(), "no pending payables") {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}

		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Return 207 Multi-Status if there were any failures
	statusCode := http.StatusOK
	if result.Failed > 0 {
		statusCode = http.StatusMultiStatus
	}

	c.JSON(statusCode, view.CreateResponse[any](result, nil, nil, nil, ""))
}

// isValidMonthFormat validates month is in YYYY-MM format
func isValidMonthFormat(month string) bool {
	// Check format: YYYY-MM (7 characters, hyphen at position 4)
	if len(month) != 7 {
		return false
	}
	if month[4] != '-' {
		return false
	}

	// Parse to validate it's a valid date
	parts := strings.Split(month, "-")
	if len(parts) != 2 {
		return false
	}

	// Year should be 4 digits
	if len(parts[0]) != 4 {
		return false
	}

	// Month should be 2 digits and between 01-12
	if len(parts[1]) != 2 {
		return false
	}

	// Validate characters are numeric
	for _, c := range parts[0] {
		if c < '0' || c > '9' {
			return false
		}
	}
	for _, c := range parts[1] {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}
