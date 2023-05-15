package vault

import "github.com/gin-gonic/gin"

type IHandler interface {
	StoreVaultTransaction(c *gin.Context)
}
