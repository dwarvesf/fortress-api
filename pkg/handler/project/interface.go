package project

import "github.com/gin-gonic/gin"

type IHandler interface {
	Create(c *gin.Context)
	List(c *gin.Context)
	UpdateProjectStatus(c *gin.Context)
}
