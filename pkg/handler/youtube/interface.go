package youtube

import "github.com/gin-gonic/gin"

type IHandler interface {
	TranscribeBroadcast(c *gin.Context)
}
