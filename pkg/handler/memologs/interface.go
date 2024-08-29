package memologs

import "github.com/gin-gonic/gin"

type IHandler interface {
	Create(c *gin.Context)
	List(c *gin.Context)
	Sync(c *gin.Context)
	ListOpenPullRequest(c *gin.Context)
	ListByDiscordID(c *gin.Context)
}
