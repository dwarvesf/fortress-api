package metadata

import "github.com/gin-gonic/gin"

type IHandler interface {
	WorkingStatus(c *gin.Context)
	Chapters(c *gin.Context)
	AccountRoles(c *gin.Context)
	AccountStatuses(c *gin.Context)
	Positions(c *gin.Context)
	GetCountries(c *gin.Context)
	GetCities(c *gin.Context)
}
