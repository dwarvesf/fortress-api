package deliverymetric

import "github.com/gin-gonic/gin"

type IHandler interface {
	GetWeeklyReport(c *gin.Context)
	GetMonthlyReport(c *gin.Context)
	GetWeeklyLeaderBoard(c *gin.Context)
	GetMonthlyLeaderBoard(c *gin.Context)

	GetWeeklyReportDiscordMsg(c *gin.Context)
	GetMonthlyReportDiscordMsg(c *gin.Context)

	Sync(c *gin.Context)
}
