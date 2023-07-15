package deliverymetric

import (
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/gin-gonic/gin"
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/logger"
)

func (h *handler) GetWeeklyReport(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "delivery",
		"method":  "GetWeeklyReport",
	})

	// Get data of current week
	report, err := h.controller.DeliveryMetric.GetWeeklyReport()
	if err != nil {
		l.Error(err, "failed to get weekly report")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, "failed to get weekly report"))
		return
	}

	// Return data
	c.JSON(http.StatusOK, view.CreateResponse[any](report, nil, nil, nil, "ok"))
}
