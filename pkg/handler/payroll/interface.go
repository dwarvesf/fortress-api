package payroll

import "github.com/gin-gonic/gin"

type IHandler interface {
	GetPayrollsByMonth(c *gin.Context)
	GetPayrollsBHXH(c *gin.Context)
	CommitPayroll(c *gin.Context)
	MarkPayrollAsPaid(c *gin.Context)
}
