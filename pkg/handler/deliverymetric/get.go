package deliverymetric

import (
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
)

func (h *handler) GetWeeklyReport(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "delivery",
		"method":  "GetWeeklyReport",
	})

	sync := c.Query("sync")
	if sync == "true" {
		err := h.controller.DeliveryMetric.Sync()
		if err != nil {
			l.Error(err, "failed to sync new data")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, "failed to sync new data"))
			return
		}
	}

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

func (h *handler) GetMonthlyReport(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "delivery",
		"method":  "GetMonthlyReport",
	})

	sync := c.Query("sync")
	if sync == "true" {
		err := h.controller.DeliveryMetric.Sync()
		if err != nil {
			l.Error(err, "failed to sync new data")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, "failed to sync new data"))
			return
		}
	}

	// Get data of current month
	report, err := h.controller.DeliveryMetric.GetMonthlyReport()
	if err != nil {
		l.Error(err, "failed to get monthly report")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, "failed to get monthly report"))
		return
	}

	// Return data
	c.JSON(http.StatusOK, view.CreateResponse[any](report, nil, nil, nil, "ok"))
}
