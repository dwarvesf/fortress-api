package deliverymetric

import (
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/discord"
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/gin-gonic/gin"
)

func (h *handler) GetWeeklyReportDiscordMsg(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "delivery",
		"method":  "GetWeeklyReportDiscordMsg",
	})

	report, err := h.controller.DeliveryMetric.GetWeeklyReport()
	if err != nil {
		l.Errorf(err, "failed to get delivery metric weekly report", "body")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	leaderBoard, err := h.controller.DeliveryMetric.GetWeeklyLeaderBoard()
	if err != nil {
		l.Errorf(err, "failed to get delivery metric weekly report")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	reportView := view.ToDeliveryMetricWeeklyReport(report)
	leaderBoardView := view.ToDeliveryMetricLeaderBoard(leaderBoard)

	msg := discord.CreateDeliveryMetricWeeklyReportMessage(reportView, leaderBoardView)
	c.JSON(http.StatusOK, view.CreateResponse[any](msg, nil, nil, nil, ""))
}

func (h *handler) GetMonthlyReportDiscordMsg(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "delivery",
		"method":  "GetMonthlyReportDiscordMsg",
	})

	report, err := h.controller.DeliveryMetric.GetMonthlyReport()
	if err != nil {
		l.Errorf(err, "failed to get delivery metric weekly report")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	currentMonthReport := report.Reports[1]
	previousMonthReport := report.Reports[2]

	if c.Query("to-now") == "true" {
		currentMonthReport = report.Reports[0]
		previousMonthReport = report.Reports[1]
	}

	leaderBoard, err := h.controller.DeliveryMetric.GetMonthlyLeaderBoard(currentMonthReport.Month)
	if err != nil {
		l.Errorf(err, "failed to get delivery metric weekly report")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	reportView := view.ToDeliveryMetricMonthlyReport(currentMonthReport, previousMonthReport)
	leaderBoardView := view.ToDeliveryMetricLeaderBoard(leaderBoard)

	msg := discord.CreateDeliveryMetricMonthlyReportMessage(reportView, leaderBoardView)
	c.JSON(http.StatusOK, view.CreateResponse[any](msg, nil, nil, nil, ""))
}
