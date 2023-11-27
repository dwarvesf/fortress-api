package discord

import "github.com/gin-gonic/gin"

type IHandler interface {
	SyncDiscordInfo(c *gin.Context)
	BirthdayDailyMessage(c *gin.Context)
	OnLeaveMessage(c *gin.Context)
	ReportBraineryMetrics(c *gin.Context)
	DeliveryMetricsReport(c *gin.Context)
	PublishIcyActivityLog(c *gin.Context)
}
