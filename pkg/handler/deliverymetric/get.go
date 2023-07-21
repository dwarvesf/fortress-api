package deliverymetric

import (
	"net/http"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/gin-gonic/gin"

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
	c.JSON(http.StatusOK, view.CreateResponse[any](report, nil, nil, nil, ""))
}

func (h *handler) GetMonthlyReport(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "delivery",
		"method":  "GetMonthlyReport",
	})

	// Get data of current month
	report, err := h.controller.DeliveryMetric.GetMonthlyReport()
	if err != nil {
		l.Error(err, "failed to get monthly report")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, "failed to get monthly report"))
		return
	}

	// Return data
	c.JSON(http.StatusOK, view.CreateResponse[any](report, nil, nil, nil, ""))
}

func (h *handler) GetWeeklyLeaderBoard(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "delivery",
		"method":  "GetWeeklyLeaderBoard",
	})

	// Get data of current week
	report, err := h.controller.DeliveryMetric.GetWeeklyLeaderBoard()
	if err != nil {
		l.Error(err, "failed to get weekly leaderboard")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, "failed to get weekly leaderboard"))
		return
	}

	// Return data
	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToDeliveryMetricLeaderBoard(report), nil, nil, nil, ""))
}

func (h *handler) GetMonthlyLeaderBoard(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "delivery",
		"method":  "GetMonthlyLeaderBoard",
	})

	var month *time.Time
	if c.Query("month") != "" {
		m, err := time.Parse(time.RFC3339, c.Query("month"))
		if err != nil {
			l.Error(err, "failed to parse month")
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, "failed to parse month"))
			return
		}
		month = &m
	}

	// Get data of current month
	report, err := h.controller.DeliveryMetric.GetMonthlyLeaderBoard(month)
	if err != nil {
		l.Error(err, "failed to get monthly leaderboard")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, "failed to get monthly leaderboard"))
		return
	}

	// Return data
	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToDeliveryMetricLeaderBoard(report), nil, nil, nil, ""))
}
