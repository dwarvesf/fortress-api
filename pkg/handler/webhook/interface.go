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
	HandleNocodbLeave(c *gin.Context)
	HandleNotionLeave(c *gin.Context)
	HandleNotionRefund(c *gin.Context)
	HandleNotionOnLeave(c *gin.Context)
	HandleNotionTimesheet(c *gin.Context)
	HandleNotionInvoiceGenerate(c *gin.Context)
	HandleNotionInvoiceSend(c *gin.Context)
	HandleNotionTaskOrderSendEmail(c *gin.Context)
	HandleDiscordInteraction(c *gin.Context)
	HandleGenInvoice(c *gin.Context)
}
