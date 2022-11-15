package employee

import "github.com/gin-gonic/gin"

type IHandler interface {
	List(c *gin.Context)
	One(c *gin.Context)
	UpdateEmployeeStatus(c *gin.Context)
	GetProfile(c *gin.Context)
	UpdateGeneralInfo(c *gin.Context)
	Create(c *gin.Context)
	UpdateSkills(c *gin.Context)
	UpdatePersonalInfo(c *gin.Context)
}
