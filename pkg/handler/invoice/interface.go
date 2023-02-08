package invoice

import "github.com/gin-gonic/gin"

type IHandler interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	GetLatestInvoice(c *gin.Context)
}
