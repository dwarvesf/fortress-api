package discord

import "github.com/gin-gonic/gin"

type IHandler interface {
	SyncDiscordInfo(c *gin.Context)
	BirthdayDailyMessage(c *gin.Context)
	OnLeaveMessage(c *gin.Context)
	ReportBraineryMetrics(c *gin.Context)
	DeliveryMetricsReport(c *gin.Context)
	SyncMemo(c *gin.Context)
	NotifyWeeklyMemos(c *gin.Context)
	CreateScheduledEvent(c *gin.Context)
	ListScheduledEvent(c *gin.Context)
	SetScheduledEventSpeakers(c *gin.Context)
	ListDiscordResearchTopics(c *gin.Context)
}
