package metadata_handler

import "github.com/gin-gonic/gin"

type IHandler interface {
	WorkingStatus(c *gin.Context)
}
