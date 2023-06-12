package discord

import "github.com/gin-gonic/gin"

type IHandler interface {
	SyncDiscordInfo(c *gin.Context)
	BirthdayDailyMessage(c *gin.Context)
	OnLeaveMessage(c *gin.Context)
}
