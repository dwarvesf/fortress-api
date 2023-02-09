package webhook

import "github.com/gin-gonic/gin"

type IHandler interface {
	N8n(c *gin.Context)
}
