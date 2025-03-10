package dynamicevents

import (
	"github.com/gin-gonic/gin"
)

type IHandler interface {
	Events(c *gin.Context)
}
