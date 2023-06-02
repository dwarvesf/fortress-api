package webhook

import "github.com/gin-gonic/gin"

type IHandler interface {
	N8n(c *gin.Context)
	CreateBasecampExpense(c *gin.Context)
	MarkInvoiceAsPaidViaBasecamp(c *gin.Context)
	StoreAccountingTransaction(c *gin.Context)
	UncheckBasecampExpense(c *gin.Context)
	ValidateBasecampExpense(c *gin.Context)
}
