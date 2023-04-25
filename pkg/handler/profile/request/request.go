package request

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"strings"
)

// UpdateInfoInput input model for update profile
type UpdateInfoInput struct {
	PersonalEmail      string `form:"personalEmail" json:"personalEmail" binding:"required,email"`
	PhoneNumber        string `form:"phoneNumber" json:"phoneNumber" binding:"required,max=18,min=8"`
	PlaceOfResidence   string `form:"placeOfResidence" json:"placeOfResidence" binding:"required"`
	Address            string `form:"address" json:"address"`
	Country            string `form:"country" json:"country"`
	City               string `form:"city" json:"city"`
	GithubID           string `form:"githubID" json:"githubID"`
	NotionID           string `form:"notionID" json:"notionID"`
	NotionName         string `form:"notionName" json:"notionName"`
	NotionEmail        string `form:"notionEmail" json:"notionEmail"`
	DiscordName        string `form:"discordName" json:"discordName"`
	LinkedInName       string `form:"linkedInName" json:"linkedInName"`
	WiseRecipientID    string `form:"wiseRecipientID" json:"wiseRecipientID"`
	WiseRecipientEmail string `form:"wiseRecipientEmail" json:"wiseRecipientEmail" binding:"email"`
	WiseRecipientName  string `form:"wiseRecipientName" json:"wiseRecipientName"`
	WiseAccountNumber  string `form:"wiseAccountNumber" json:"wiseAccountNumber"`
	WiseCurrency       string `form:"wiseCurrency" json:"wiseCurrency"`
}

func (i UpdateInfoInput) MapEmployeeInput(employee *model.Employee) {
	employee.PersonalEmail = i.PersonalEmail
	employee.PhoneNumber = i.PhoneNumber
	employee.PlaceOfResidence = i.PlaceOfResidence
	employee.Address = i.Address
	employee.City = i.City
	employee.Country = i.Country

	if i.GithubID != "" {
		employee.GithubID = i.GithubID
	}

	if i.NotionID != "" {
		employee.NotionID = i.NotionID
	}

	if i.NotionName != "" {
		employee.NotionName = i.NotionName
	}

	if i.NotionEmail != "" {
		employee.NotionEmail = i.NotionEmail
	}

	if i.DiscordName != "" {
		employee.DiscordName = i.DiscordName
	}

	if i.LinkedInName != "" {
		employee.LinkedInName = i.LinkedInName
	}

	if strings.TrimSpace(i.WiseRecipientID) != "" {
		employee.WiseRecipientID = i.WiseRecipientID
	}
	if strings.TrimSpace(i.WiseRecipientEmail) != "" {
		employee.WiseRecipientEmail = i.WiseRecipientEmail
	}
	if strings.TrimSpace(i.WiseRecipientName) != "" {
		employee.WiseRecipientName = i.WiseRecipientName
	}
	if strings.TrimSpace(i.WiseAccountNumber) != "" {
		employee.WiseAccountNumber = i.WiseAccountNumber
	}
	if strings.TrimSpace(i.WiseCurrency) != "" {
		employee.WiseCurrency = i.WiseCurrency
	}
}
