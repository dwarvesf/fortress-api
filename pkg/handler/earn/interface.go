package earn

import "github.com/gin-gonic/gin"

type IHandler interface {
	ListEarn(c *gin.Context)
}
