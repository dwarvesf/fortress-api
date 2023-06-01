package engagement

import "github.com/gin-gonic/gin"

type IHandler interface {
	UpsertRollup(c *gin.Context)
	GetLastMessageID(c *gin.Context)
	IndexMessages(c *gin.Context)
}
