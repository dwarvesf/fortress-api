package webhook

import "github.com/gin-gonic/gin"

type IHandler interface {
	N8n(c *gin.Context)
	CreateBasecampExpense(c *gin.Context)
	MarkInvoiceAsPaidViaBasecamp(c *gin.Context)
	MarkInvoiceAsPaidViaNoco(c *gin.Context)
	HandleNocoExpense(c *gin.Context)
	StoreAccountingTransaction(c *gin.Context)
	StoreNocoAccountingTransaction(c *gin.Context)
	UncheckBasecampExpense(c *gin.Context)
	ValidateBasecampExpense(c *gin.Context)
	ValidateOnLeaveRequest(c *gin.Context)
	ApproveOnLeaveRequest(c *gin.Context)
}
