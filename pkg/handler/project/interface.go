package project

import "github.com/gin-gonic/gin"

type IHandler interface {
	ArchiveWorkUnit(c *gin.Context)
	AssignMember(c *gin.Context)
	Create(c *gin.Context)
	CreateWorkUnit(c *gin.Context)
	DeleteMember(c *gin.Context)
	DeleteSlot(c *gin.Context)
	Details(c *gin.Context)
	GetMembers(c *gin.Context)
	GetWorkUnits(c *gin.Context)
	List(c *gin.Context)
	SyncProjectMemberStatus(c *gin.Context)
	UnarchiveWorkUnit(c *gin.Context)
	UnassignMember(c *gin.Context)
	UpdateContactInfo(c *gin.Context)
	UpdateGeneralInfo(c *gin.Context)
	UpdateMember(c *gin.Context)
	UpdateProjectStatus(c *gin.Context)
	UpdateSendingSurveyState(c *gin.Context)
	UpdateWorkUnit(c *gin.Context)
	UploadAvatar(c *gin.Context)
	IcyWeeklyDistribution(c *gin.Context)
}
