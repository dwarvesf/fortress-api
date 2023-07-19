package delivery

import (
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/gin-gonic/gin"
)

func (h handler) GetWeeklyReport(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "delivery",
		"method":  "GetWeeklyReport",
	})

	// Get data of current week
	report, err := h.controller.DeliveryMetric.GetWeeklyReport()
	if err != nil {
		l.Error(err, "failed to get weekly report")
		c.JSON(500, gin.H{
			"message": "failed to get weekly report",
		})
		return
	}

	// Return data
	c.JSON(200, report)
}
