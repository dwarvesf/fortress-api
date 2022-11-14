package employee

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type EditGeneralInfo struct {
	Fullname    string     `form:"fullName" json:"fullName" binding:"required,max=99"`
	Email       string     `form:"email" json:"email" binding:"required,email"`
	Phone       string     `form:"phone" json:"phone" binding:"required,max=12,min=10"`
	LineManager model.UUID `form:"lineManager" json:"lineManager"`
	DiscordID   string     `form:"discordID" json:"discordID"`
	GithubID    string     `form:"githubID" json:"githubID"`
}

type EditSkills struct {
	Role      model.UUID   `form:"role" json:"role" binding:"required"`
	Chapter   model.UUID   `form:"chapter" json:"chapter"`
	Seniority model.UUID   `form:"seniority" json:"seniority" binding:"required"`
	Stack     []model.UUID `form:"stack" json:"stack" binding:"required"`
}

type EditPersonalInfo struct {
	DoB           *time.Time `form:"dob" json:"dob" binding:"required"`
	Gender        string     `form:"gender" json:"gender" binding:"required"`
	Address       string     `form:"address" json:"address" binding:"required,max=200"`
	PersonalEmail string     `form:"personalEmail" json:"personalEmail" binding:"required,email"`
}

type GetListEmployeeQuery struct {
	model.Pagination

	WorkingStatus string `json:"workingStatus" form:"workingStatus"`
}
