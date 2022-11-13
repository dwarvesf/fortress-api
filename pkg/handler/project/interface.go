package project

import "github.com/gin-gonic/gin"

type IHandler interface {
	List(c *gin.Context)
	UpdateProjectStatus(c *gin.Context)
}
