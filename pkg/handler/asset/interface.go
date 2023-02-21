package asset

import "github.com/gin-gonic/gin"

type IHandler interface {
	Upload(c *gin.Context)
}
