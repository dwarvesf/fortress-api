package employee_handler

import "github.com/gin-gonic/gin"

type IHandler interface {
	List(c *gin.Context)
	One(c *gin.Context)
}
