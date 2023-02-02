package audit

import "github.com/gin-gonic/gin"

type IHandler interface {
	Sync(c *gin.Context)
}
