package view

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type ProfileData struct {
	ID            model.UUID `json:"id"`
	FullName      string     `json:"fullName"`
	DisplayName   string     `json:"displayName"`
	Avatar        string     `json:"avatar"`
	Gender        string     `json:"gender"`
	DateOfBirth   *time.Time `json:"birthday"`
	TeamEmail     string     `json:"teamEmail"`
	PersonalEmail string     `json:"personalEmail"`
	PhoneNumber   string     `json:"phoneNumber"`
	GithubID      string     `json:"githubID"`
	NotionID      string     `json:"notionID"`
	NotionName    string     `json:"notionName"`
	DiscordID     string     `json:"discordID"`
	DiscordName   string     `json:"discordName"`
	Username      string     `json:"username"`
}

type UpdateProfileInfoData struct {
	model.BaseModel

	// basic info
	TeamEmail   string `json:"teamEmail"`
	PhoneNumber string `json:"phoneNumber"`
	GithubID    string `json:"githubID"`
	NotionID    string `json:"notionID"`
	NotionName  string `json:"notionName"`
	DiscordID   string `json:"discordID"`
	DiscordName string `json:"discordName"`
	Username    string `json:"username"`
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
		TeamEmail:   employee.TeamEmail,
		PhoneNumber: employee.PhoneNumber,
		GithubID:    employee.GithubID,
		NotionID:    employee.NotionID,
		NotionName:  employee.NotionName,
		DiscordID:   employee.DiscordID,
		DiscordName: employee.DiscordName,
		Username:    employee.Username,
	}

	return rs
}

func ToProfileData(employee *model.Employee) *ProfileData {
	return &ProfileData{
		ID:            employee.ID,
		FullName:      employee.FullName,
		DisplayName:   employee.DisplayName,
		Avatar:        employee.Avatar,
		Gender:        employee.Gender,
		DateOfBirth:   employee.DateOfBirth,
		TeamEmail:     employee.TeamEmail,
		PersonalEmail: employee.PersonalEmail,
		PhoneNumber:   employee.PhoneNumber,
		GithubID:      employee.GithubID,
		NotionID:      employee.NotionID,
		NotionName:    employee.NotionName,
		DiscordID:     employee.DiscordID,
		DiscordName:   employee.DiscordName,
		Username:      employee.Username,
	}
}
