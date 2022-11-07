package metadata

import "github.com/gin-gonic/gin"

type IHandler interface {
	WorkingStatus(c *gin.Context)
}
