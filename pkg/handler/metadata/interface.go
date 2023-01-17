package metadata

import "github.com/gin-gonic/gin"

type IHandler interface {
	WorkingStatuses(c *gin.Context)
	Seniorities(c *gin.Context)
	Chapters(c *gin.Context)
	AccountRoles(c *gin.Context)
	Positions(c *gin.Context)
	GetCountries(c *gin.Context)
	GetCities(c *gin.Context)
	ProjectStatuses(c *gin.Context)
	Stacks(c *gin.Context)
	GetQuestions(c *gin.Context)
	UpdateStack(c *gin.Context)
	CreateStack(c *gin.Context)
	DeleteStack(c *gin.Context)
	UpdatePosition(c *gin.Context)
	CreatePosition(c *gin.Context)
	DeletePosition(c *gin.Context)
	Organizations(c *gin.Context)
}
