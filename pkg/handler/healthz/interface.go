package healthz_handler

import "github.com/gin-gonic/gin"

type IHandler interface {
	Healthz(c *gin.Context)
}
