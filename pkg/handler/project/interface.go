package project

import "github.com/gin-gonic/gin"

type IHandler interface {
	Create(c *gin.Context)
	List(c *gin.Context)
	UpdateProjectStatus(c *gin.Context)
	GetMembers(c *gin.Context)
	UpdateMember(c *gin.Context)
	AssignMember(c *gin.Context)
	UnassignMember(c *gin.Context)
	DeleteMember(c *gin.Context)
	DeleteSlot(c *gin.Context)
	Details(c *gin.Context)
	UpdateGeneralInfo(c *gin.Context)
	UpdateContactInfo(c *gin.Context)
	GetWorkUnits(c *gin.Context)
	CreateWorkUnit(c *gin.Context)
	UpdateWorkUnit(c *gin.Context)
	ArchiveWorkUnit(c *gin.Context)
	UnarchiveWorkUnit(c *gin.Context)
	UpdateSendingSurveyState(c *gin.Context)
	UploadAvatar(c *gin.Context)
}
