package news

import (
	"github.com/gin-gonic/gin"
)

type IHandler interface {
	Fetch(c *gin.Context)
}
