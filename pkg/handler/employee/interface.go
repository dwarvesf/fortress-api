package employee

import "github.com/gin-gonic/gin"

type IHandler interface {
	List(c *gin.Context)
	One(c *gin.Context)
	UpdateEmployeeStatus(c *gin.Context)
	GetProfile(c *gin.Context)
}
