package view

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type ProfileData struct {
	ID               model.UUID `json:"id"`
	FullName         string     `json:"fullName"`
	DisplayName      string     `json:"displayName"`
	Avatar           string     `json:"avatar"`
	Gender           string     `json:"gender"`
	DateOfBirth      *time.Time `json:"birthday"`
	TeamEmail        string     `json:"teamEmail"`
	PersonalEmail    string     `json:"personalEmail"`
	PhoneNumber      string     `json:"phoneNumber"`
	GithubID         string     `json:"githubID"`
	NotionID         string     `json:"notionID"`
	NotionName       string     `json:"notionName"`
	NotionEmail      string     `json:"notionEmail"`
	DiscordID        string     `json:"discordID"`
	DiscordName      string     `json:"discordName"`
	Username         string     `json:"username"`
	PlaceOfResidence string     `json:"placeOfResidence"`
	Address          string     `json:"address"`
	Country          string     `json:"country"`
	City             string     `json:"city"`
	LinkedInName     string     `json:"linkedInName"`
	Roles            []Role     `json:"roles"`
}

type UpdateProfileInfoData struct {
	model.BaseModel

	// basic info
	TeamEmail        string `json:"teamEmail"`
	PhoneNumber      string `json:"phoneNumber"`
	GithubID         string `json:"githubID"`
	NotionID         string `json:"notionID"`
	NotionName       string `json:"notionName"`
	NotionEmail      string `json:"notionEmail"`
	DiscordID        string `json:"discordID"`
	DiscordName      string `json:"discordName"`
	Username         string `json:"username"`
	PlaceOfResidence string `json:"placeOfResidence"`
	Address          string `json:"address"`
	Country          string `json:"country"`
	City             string `json:"city"`
	LinkedInName     string `json:"linkedInName"`
}

type ProfileDataResponse struct {
	Data ProfileData `json:"data"`
}

type UpdateProfileInfoResponse struct {
	Data UpdateProfileInfoData `json:"data"`
}

func ToUpdateProfileInfoData(employee *model.Employee) *UpdateProfileInfoData {
	rs := &UpdateProfileInfoData{
		BaseModel: model.BaseModel{
			ID:        employee.ID,
			CreatedAt: employee.CreatedAt,
			UpdatedAt: employee.UpdatedAt,
		},
		TeamEmail:        employee.TeamEmail,
		PhoneNumber:      employee.PhoneNumber,
		GithubID:         employee.GithubID,
		NotionID:         employee.NotionID,
		NotionName:       employee.NotionName,
		NotionEmail:      employee.NotionEmail,
		DiscordID:        employee.DiscordID,
		DiscordName:      employee.DiscordName,
		Username:         employee.Username,
		PlaceOfResidence: employee.PlaceOfResidence,
		Address:          employee.Address,
		Country:          employee.Country,
		City:             employee.City,
		LinkedInName:     employee.LinkedInName,
	}

	return rs
}

func ToProfileData(employee *model.Employee) *ProfileData {
	empSocialData := SocialAccount{}
	for _, sa := range employee.SocialAccounts {
		switch sa.Type {
		case model.SocialAccountTypeDiscord:
			empSocialData.DiscordID = sa.AccountID
			empSocialData.DiscordName = sa.Name
		case model.SocialAccountTypeGitHub:
			empSocialData.GithubID = sa.AccountID
		case model.SocialAccountTypeNotion:
			empSocialData.NotionID = sa.AccountID
			empSocialData.NotionName = sa.Name
		case model.SocialAccountTypeLinkedIn:
			empSocialData.LinkedInName = sa.AccountID
		}
	}

	return &ProfileData{
		ID:               employee.ID,
		FullName:         employee.FullName,
		DisplayName:      employee.DisplayName,
		Avatar:           employee.Avatar,
		Gender:           employee.Gender,
		DateOfBirth:      employee.DateOfBirth,
		TeamEmail:        employee.TeamEmail,
		PersonalEmail:    employee.PersonalEmail,
		PhoneNumber:      employee.PhoneNumber,
		Username:         employee.Username,
		PlaceOfResidence: employee.PlaceOfResidence,
		Address:          employee.Address,
		Country:          employee.Country,
		City:             employee.City,
		GithubID:         empSocialData.GithubID,
		NotionID:         empSocialData.NotionID,
		NotionName:       empSocialData.NotionName,
		NotionEmail:      empSocialData.NotionEmail,
		DiscordID:        empSocialData.DiscordID,
		DiscordName:      empSocialData.DiscordName,
		LinkedInName:     empSocialData.LinkedInName,
		Roles:            ToRoles(employee.EmployeeRoles),
	}
}
