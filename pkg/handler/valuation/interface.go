package valuation

import "github.com/gin-gonic/gin"

type IHandler interface {
	One(c *gin.Context)
}
