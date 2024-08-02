package youtube

import "github.com/gin-gonic/gin"

type IHandler interface {
	LatestBroadcast(c *gin.Context)
	TranscribeBroadcast(c *gin.Context)
}
