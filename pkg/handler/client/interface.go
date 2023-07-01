package client

import "github.com/gin-gonic/gin"

type IHandler interface {
	Create(c *gin.Context)
	List(c *gin.Context)
	Detail(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)

	PublicList(c *gin.Context)
}
