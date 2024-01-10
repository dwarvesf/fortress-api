package metadata

import "github.com/gin-gonic/gin"

type IHandler interface {
	Banks(c *gin.Context)
	Chapters(c *gin.Context)
	CreatePosition(c *gin.Context)
	CreateStack(c *gin.Context)
	DeletePosition(c *gin.Context)
	DeleteStack(c *gin.Context)
	GetCities(c *gin.Context)
	GetCountries(c *gin.Context)
	GetCurrencies(c *gin.Context)
	GetQuestions(c *gin.Context)
	GetRoles(c *gin.Context)
	Organizations(c *gin.Context)
	Positions(c *gin.Context)
	ProjectStatuses(c *gin.Context)
	Seniorities(c *gin.Context)
	Stacks(c *gin.Context)
	UpdatePosition(c *gin.Context)
	UpdateStack(c *gin.Context)
	WorkingStatuses(c *gin.Context)
}
