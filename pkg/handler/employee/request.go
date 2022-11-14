package employee

import "github.com/dwarvesf/fortress-api/pkg/model"

type GetListEmployeeQuery struct {
	model.Pagination

	WorkingStatus string `json:"workingStatus" form:"workingStatus"`
}

type EditGeneralInfo struct {
	Fullname      string `form:"fullName" json:"fullName" binding:"required,max=99"`
	Email         string `form:"email" json:"email" binding:"required,email"`
	Phone         string `form:"phone" json:"phone" binding:"required,max=12,min=10"`
	LineManagerID string `form:"lineManagerID" json:"lineManagerID"`
	DiscordID     string `form:"discordID" json:"discordID"`
	GithubID      string `form:"githubID" json:"githubID"`
	NotionID      string `form:"notionID" json:"notionID"`
}
