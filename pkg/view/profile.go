package view

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type ProfileData struct {
	ID                 string     `json:"id"`
	FullName           string     `json:"fullName"`
	DisplayName        string     `json:"displayName"`
	Avatar             string     `json:"avatar"`
	Gender             string     `json:"gender"`
	DateOfBirth        *time.Time `json:"birthday"`
	TeamEmail          string     `json:"teamEmail"`
	PersonalEmail      string     `json:"personalEmail"`
	PhoneNumber        string     `json:"phoneNumber"`
	GithubID           string     `json:"githubID"`
	NotionID           string     `json:"notionID"`
	NotionName         string     `json:"notionName"`
	NotionEmail        string     `json:"notionEmail"`
	DiscordID          string     `json:"discordID"`
	DiscordName        string     `json:"discordName"`
	Username           string     `json:"username"`
	PlaceOfResidence   string     `json:"placeOfResidence"`
	Address            string     `json:"address"`
	Country            string     `json:"country"`
	City               string     `json:"city"`
	LinkedInName       string     `json:"linkedInName"`
	Roles              []Role     `json:"roles"`
	WiseRecipientID    string     `json:"wiseRecipientID"`
	WiseAccountNumber  string     `json:"wiseAccountNumber"`
	WiseRecipientEmail string     `json:"wiseRecipientEmail"`
	WiseRecipientName  string     `json:"wiseRecipientName"`
	WiseCurrency       string     `json:"wiseCurrency"`
} // @name ProfileData

type UpdateProfileInfoData struct {
	ID        string     `json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`

	// basic info
	TeamEmail          string `json:"teamEmail"`
	PhoneNumber        string `json:"phoneNumber"`
	GithubID           string `json:"githubID"`
	NotionID           string `json:"notionID"`
	NotionName         string `json:"notionName"`
	NotionEmail        string `json:"notionEmail"`
	DiscordID          string `json:"discordID"`
	DiscordName        string `json:"discordName"`
	Username           string `json:"username"`
	PlaceOfResidence   string `json:"placeOfResidence"`
	Address            string `json:"address"`
	Country            string `json:"country"`
	City               string `json:"city"`
	LinkedInName       string `json:"linkedInName"`
	WiseRecipientID    string `json:"wiseRecipientID"`
	WiseAccountNumber  string `json:"wiseAccountNumber"`
	WiseRecipientEmail string `json:"wiseRecipientEmail"`
	WiseRecipientName  string `json:"wiseRecipientName"`
	WiseCurrency       string `json:"wiseCurrency"`
} // @name UpdateProfileInfoData

type ProfileDataResponse struct {
	Data ProfileData `json:"data"`
} // @name ProfileDataResponse

type UpdateProfileInfoResponse struct {
	Data UpdateProfileInfoData `json:"data"`
} // @name UpdateProfileInfoResponse

func ToUpdateProfileInfoData(employee *model.Employee) *UpdateProfileInfoData {
	rs := &UpdateProfileInfoData{
		ID:                 employee.ID.String(),
		CreatedAt:          employee.CreatedAt,
		UpdatedAt:          employee.UpdatedAt,
		TeamEmail:          employee.TeamEmail,
		PhoneNumber:        employee.PhoneNumber,
		Username:           employee.Username,
		PlaceOfResidence:   employee.PlaceOfResidence,
		Address:            employee.Address,
		Country:            employee.Country,
		City:               employee.City,
		WiseRecipientID:    employee.WiseRecipientID,
		WiseAccountNumber:  employee.WiseAccountNumber,
		WiseRecipientEmail: employee.WiseRecipientEmail,
		WiseRecipientName:  employee.WiseRecipientName,
		WiseCurrency:       employee.WiseCurrency,
	}

	if len(employee.SocialAccounts) > 0 {
		for _, sa := range employee.SocialAccounts {
			switch sa.Type {
			case model.SocialAccountTypeGitHub:
				rs.GithubID = sa.AccountID
			case model.SocialAccountTypeNotion:
				rs.NotionID = sa.AccountID
				rs.NotionName = sa.Name
			case model.SocialAccountTypeLinkedIn:
				rs.LinkedInName = sa.Name
			}
		}
	}

	if employee.DiscordAccount != nil {
		rs.DiscordID = employee.DiscordAccount.DiscordID
		rs.DiscordName = employee.DiscordAccount.Username
	}

	return rs
}

func ToProfileData(employee *model.Employee) *ProfileData {
	empSocialData := SocialAccount{}
	for _, sa := range employee.SocialAccounts {
		switch sa.Type {
		case model.SocialAccountTypeGitHub:
			empSocialData.GithubID = sa.AccountID
		case model.SocialAccountTypeNotion:
			empSocialData.NotionID = sa.AccountID
			empSocialData.NotionName = sa.Name
		case model.SocialAccountTypeLinkedIn:
			empSocialData.LinkedInName = sa.AccountID
		}
	}

	rs := &ProfileData{
		ID:                 employee.ID.String(),
		FullName:           employee.FullName,
		DisplayName:        employee.DisplayName,
		Avatar:             employee.Avatar,
		Gender:             employee.Gender,
		DateOfBirth:        employee.DateOfBirth,
		TeamEmail:          employee.TeamEmail,
		PersonalEmail:      employee.PersonalEmail,
		PhoneNumber:        employee.PhoneNumber,
		Username:           employee.Username,
		PlaceOfResidence:   employee.PlaceOfResidence,
		Address:            employee.Address,
		Country:            employee.Country,
		City:               employee.City,
		WiseRecipientID:    employee.WiseRecipientID,
		WiseAccountNumber:  employee.WiseAccountNumber,
		WiseRecipientEmail: employee.WiseRecipientEmail,
		WiseRecipientName:  employee.WiseRecipientName,
		WiseCurrency:       employee.WiseCurrency,
		GithubID:           empSocialData.GithubID,
		NotionID:           empSocialData.NotionID,
		NotionName:         empSocialData.NotionName,
		NotionEmail:        empSocialData.NotionEmail,
		LinkedInName:       empSocialData.LinkedInName,
		Roles:              ToEmployeeRoles(employee.EmployeeRoles),
	}

	if employee.DiscordAccount != nil {
		rs.DiscordID = employee.DiscordAccount.DiscordID
		rs.DiscordName = employee.DiscordAccount.Username
	}

	return rs
}
