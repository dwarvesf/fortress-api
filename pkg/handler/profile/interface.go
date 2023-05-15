package profile

import "github.com/gin-gonic/gin"

type IHandler interface {
	GetProfile(c *gin.Context)
	UpdateInfo(c *gin.Context)
	SubmitOnboardingForm(c *gin.Context)
	Upload(c *gin.Context)
	UploadAvatar(c *gin.Context)
}
