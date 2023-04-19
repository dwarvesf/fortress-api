package invoice

import "github.com/gin-gonic/gin"

type IHandler interface {
	UpdateStatus(c *gin.Context)
	GetLatestInvoice(c *gin.Context)
	GetTemplate(c *gin.Context)
	Send(c *gin.Context)
}
