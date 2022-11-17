package project

import "github.com/gin-gonic/gin"

type IHandler interface {
	Create(c *gin.Context)
	List(c *gin.Context)
	UpdateProjectStatus(c *gin.Context)
	GetMembers(c *gin.Context)
	UpdateMember(c *gin.Context)
	AssignMember(c *gin.Context)
	DeleteMember(c *gin.Context)
}
