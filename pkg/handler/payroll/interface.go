package payroll

import "github.com/gin-gonic/gin"

type IHandler interface {
	ListPayrollsMonthly(c *gin.Context)
	CommitPayroll(c *gin.Context)
}
