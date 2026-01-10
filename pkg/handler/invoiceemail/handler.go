package invoiceemail

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/invoiceemail"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	processor invoiceemail.IProcessor
	logger    logger.Logger
	config    *config.Config
}

// New creates a new invoice email handler
func New(processor invoiceemail.IProcessor, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		processor: processor,
		logger:    logger,
		config:    cfg,
	}
}

// ProcessInvoiceEmailsResponse represents the response from processing invoice emails
type ProcessInvoiceEmailsResponse struct {
	TotalEmails int                          `json:"totalEmails"`
	Processed   int                          `json:"processed"`
	Skipped     int                          `json:"skipped"`
	Errors      int                          `json:"errors"`
	Results     []invoiceemail.ProcessResult `json:"results,omitempty"`
}

// ProcessInvoiceEmails godoc
// @Summary Process incoming invoice emails
// @Description Processes unread invoice emails from the monitored inbox, extracts Invoice IDs, and updates Notion payables status
// @Tags Invoice Email
// @Accept json
// @Produce json
// @Success 200 {object} view.Response{data=ProcessInvoiceEmailsResponse}
// @Failure 500 {object} view.ErrorResponse
// @Router /api/v1/cron/process-invoice-emails [post]
func (h *handler) ProcessInvoiceEmails(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "invoiceemail",
		"method":  "ProcessInvoiceEmails",
	})

	if !h.config.InvoiceListener.Enabled {
		l.Debug("invoice listener is disabled")
		c.JSON(http.StatusOK, view.CreateResponse[any](
			ProcessInvoiceEmailsResponse{
				TotalEmails: 0,
				Processed:   0,
				Skipped:     0,
				Errors:      0,
			},
			nil, nil, nil, "invoice listener is disabled"))
		return
	}

	if h.processor == nil {
		l.Debug("invoice email processor is not initialized")
		c.JSON(http.StatusOK, view.CreateResponse[any](
			ProcessInvoiceEmailsResponse{
				TotalEmails: 0,
				Processed:   0,
				Skipped:     0,
				Errors:      0,
			},
			nil, nil, nil, "invoice email processor is not initialized"))
		return
	}

	l.Debug("starting invoice email processing")

	stats, err := h.processor.ProcessIncomingInvoices(c.Request.Context())
	if err != nil {
		l.Error(err, "failed to process invoice emails")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	response := ProcessInvoiceEmailsResponse{
		TotalEmails: stats.TotalEmails,
		Processed:   stats.Processed,
		Skipped:     stats.Skipped,
		Errors:      stats.Errors,
		Results:     stats.Results,
	}

	l.Debugf("processing complete: total=%d processed=%d skipped=%d errors=%d",
		stats.TotalEmails, stats.Processed, stats.Skipped, stats.Errors)

	c.JSON(http.StatusOK, view.CreateResponse[any](response, nil, nil, nil, ""))
}
