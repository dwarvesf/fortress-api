package employee

import "github.com/gin-gonic/gin"

type IHandler interface {
	Create(c *gin.Context)
	Details(c *gin.Context)
	ListByDiscordRequest(c *gin.Context)
	GetLineManagers(c *gin.Context)
	List(c *gin.Context)
	ListWithMMAScore(c *gin.Context)
	UpdateEmployeeStatus(c *gin.Context)
	UpdateGeneralInfo(c *gin.Context)
	UpdateSkills(c *gin.Context)
	UpdatePersonalInfo(c *gin.Context)
	UploadAvatar(c *gin.Context)
	UpdateRole(c *gin.Context)
	UpdateBaseSalary(c *gin.Context)
	SalaryAdvance(c *gin.Context)
	CheckSalaryAdvance(c *gin.Context)
	SalaryAdvanceReport(c *gin.Context)
	GetEmployeeEarnTransactions(c *gin.Context)
	GetEmployeeTotalEarn(c *gin.Context)
	GetTotalEarn(c *gin.Context)

	PublicList(c *gin.Context)
}
