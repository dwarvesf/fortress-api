package invoice

import "github.com/gin-gonic/gin"

type IHandler interface {
	GetTemplate(c *gin.Context)
	List(c *gin.Context)
	Send(c *gin.Context)
	UpdateStatus(c *gin.Context)
}
