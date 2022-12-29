package request

import "github.com/dwarvesf/fortress-api/pkg/model"

// UpdateInfoInput input model for update profile
type UpdateInfoInput struct {
	TeamEmail        string `form:"teamEmail" json:"teamEmail" binding:"required,email"`
	PersonalEmail    string `form:"personalEmail" json:"personalEmail" binding:"required,email"`
	PhoneNumber      string `form:"phoneNumber" json:"phoneNumber" binding:"required,max=12,min=10"`
	GithubID         string `form:"githubID" json:"githubID"`
	NotionID         string `form:"notionID" json:"notionID"`
	NotionName       string `form:"notionName" json:"notionName"`
	NotionEmail      string `form:"notionEmail" json:"notionEmail"`
	DiscordID        string `form:"discordID" json:"discordID"`
	DiscordName      string `form:"discordName" json:"discordName"`
	LinkedInName     string `form:"linkedInName" json:"linkedInName"`
	PlaceOfResidence string `form:"placeOfResidence" json:"placeOfResidence" binding:"required"`
	Address          string `form:"address" json:"address"`
	Country          string `form:"country" json:"country"`
	City             string `form:"city" json:"city"`
}

func (i UpdateInfoInput) MapEmployeeInput(employee *model.Employee) {
	employee.TeamEmail = i.TeamEmail
	employee.PersonalEmail = i.PersonalEmail
	employee.PhoneNumber = i.PhoneNumber
	employee.PlaceOfResidence = i.PlaceOfResidence
	employee.Address = i.Address
	employee.City = i.City
	employee.Country = i.Country

	employee.DiscordID = i.DiscordID
	employee.DiscordName = i.DiscordName

	employee.NotionID = i.NotionID
	employee.NotionName = i.NotionName
	employee.NotionEmail = i.NotionEmail

	employee.GithubID = i.GithubID
	employee.LinkedInName = i.LinkedInName
}
