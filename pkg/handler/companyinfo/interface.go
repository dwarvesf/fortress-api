package companyinfo

import "github.com/gin-gonic/gin"

type IHandler interface {
	List(c *gin.Context)
}
