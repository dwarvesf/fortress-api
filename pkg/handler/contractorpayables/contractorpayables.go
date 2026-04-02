package contractorpayables

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	ctrlcontractorpayables "github.com/dwarvesf/fortress-api/pkg/controller/contractorpayables"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	controller ctrlcontractorpayables.IController
	service    *service.Service
	logger     logger.Logger
	config     *config.Config
}

// New creates a new contractor payables handler
func New(controller ctrlcontractorpayables.IController, service *service.Service, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		controller: controller,
		service:    service,
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

	if req.FileName != "" {
		year := req.Year
		if year == 0 {
			year = time.Now().Year()
		}
		if year < 2000 || year > 3000 {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, nil, nil, fmt.Sprintf("Invalid year: %d", year)))
			return
		}

		l.Fields(logger.Fields{"file_name": req.FileName, "year": year}).Debug("calling controller.PreviewCommitByFile")
		result, err := h.controller.PreviewCommitByFile(c.Request.Context(), req.FileName, year)
		if err != nil {
			l.Error(err, "failed to preview file commit")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}

		c.JSON(http.StatusOK, view.CreateResponse[any](result, nil, nil, nil, ""))
		return
	}

	// Validate month format (YYYY-MM)
	if !isValidMonthFormat(req.Month) {
		l.Error(nil, "invalid month format")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil,
			nil, nil, "Invalid month format. Use YYYY-MM (e.g., 2025-01)"))
		return
	}
	if req.Batch != 1 && req.Batch != 15 {
		l.Error(nil, "invalid batch")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil,
			nil, nil, "Invalid batch. Use 1 or 15"))
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

	result, err := h.commit(c, req, l)
	if err != nil {
		l.Error(err, "failed to commit payables")
		if strings.Contains(err.Error(), "invalid month format") || strings.Contains(err.Error(), "invalid batch") || strings.Contains(err.Error(), "invalid year") {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, nil, nil, err.Error()))
			return
		}

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

	// Send Discord notification to audit log
	msg := buildAuditMessage(req, result)

	discordMsg, err := h.service.Discord.SendMessage(model.DiscordMessage{
		Content: msg,
	}, h.config.Discord.Webhooks.AuditLog)
	if err != nil {
		l.Error(err, "failed to post Discord message")
		c.JSON(statusCode, view.CreateResponse[any](result, nil, err, discordMsg, ""))
		return
	}

	c.JSON(statusCode, view.CreateResponse[any](result, nil, nil, nil, ""))
}

func (h *handler) commit(c *gin.Context, req CommitRequest, l logger.Logger) (*ctrlcontractorpayables.CommitResponse, error) {
	if req.FileName != "" {
		year := req.Year
		if year == 0 {
			year = time.Now().Year()
		}
		if year < 2000 || year > 3000 {
			return nil, fmt.Errorf("invalid year: %d", year)
		}

		l.Fields(logger.Fields{"file_name": req.FileName, "year": year}).Debug("calling controller.CommitPayablesByFile")
		return h.controller.CommitPayablesByFile(c.Request.Context(), req.FileName, year)
	}

	if !isValidMonthFormat(req.Month) {
		return nil, fmt.Errorf("invalid month format. use YYYY-MM (e.g., 2025-01)")
	}
	if req.Batch != 1 && req.Batch != 15 {
		return nil, fmt.Errorf("invalid batch: %d", req.Batch)
	}

	l.Debug("calling controller.CommitPayables")
	return h.controller.CommitPayables(c.Request.Context(), req.Month, req.Batch, req.Contractor)
}

func buildAuditMessage(req CommitRequest, result *ctrlcontractorpayables.CommitResponse) string {
	if req.FileName != "" {
		year := req.Year
		if year == 0 {
			year = result.Year
		}
		return fmt.Sprintf("**Contractor Payables Commit**\nMode: file | Year: %d | File: %s\nUpdated: %d | Failed: %d",
			year, req.FileName, result.Updated, result.Failed)
	}

	contractorInfo := ""
	if req.Contractor != "" {
		contractorInfo = fmt.Sprintf(" for contractor=%s", req.Contractor)
	}
	return fmt.Sprintf("**Contractor Payables Commit**\nMonth: %s | Batch: %d%s\nUpdated: %d | Failed: %d",
		req.Month, req.Batch, contractorInfo, result.Updated, result.Failed)
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
