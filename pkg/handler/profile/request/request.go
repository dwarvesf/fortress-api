package request

import "github.com/dwarvesf/fortress-api/pkg/model"

// UpdateInfoInput input model for update profile
type UpdateInfoInput struct {
	TeamEmail     string `form:"teamEmail" json:"teamEmail" binding:"required,email"`
	PersonalEmail string `form:"personalEmail" json:"personalEmail" binding:"required,email"`
	PhoneNumber   string `form:"phoneNumber" json:"phoneNumber" binding:"required,max=12,min=10"`
	GithubID      string `form:"githubID" json:"githubID"`
	NotionID      string `form:"notionID" json:"notionID"`
	NotionName    string `form:"notionName" json:"notionName"`
	DiscordID     string `form:"discordID" json:"discordID"`
	DiscordName   string `form:"discordName" json:"discordName"`
}

func (i UpdateInfoInput) MapEmployeeInput(employee *model.Employee) {
	employee.TeamEmail = i.TeamEmail
	employee.PersonalEmail = i.PersonalEmail
	employee.PhoneNumber = i.PhoneNumber

	if i.DiscordID != "" {
		employee.DiscordID = i.DiscordID
	}
	if i.DiscordName != "" {
		employee.DiscordName = i.DiscordName
	}
	if i.GithubID != "" {
		employee.GithubID = i.GithubID
	}
	if i.NotionID != "" {
		employee.NotionID = i.NotionID
	}
	if i.NotionName != "" {
		employee.NotionName = i.NotionName
	}
}
