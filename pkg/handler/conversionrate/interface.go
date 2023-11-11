package conversionrate

import "github.com/gin-gonic/gin"

type IHandler interface {
	List(c *gin.Context)
	Sync(c *gin.Context)
}
