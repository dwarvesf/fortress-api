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
	DiscordID     string     `json:"discordID"`
	GithubID      string     `json:"githubID"`
	NotionID      string     `json:"notionID"`
}

type UpdateProfileInfoData struct {
	model.BaseModel

	// basic info
	TeamEmail   string `json:"teamEmail"`
	PhoneNumber string `json:"phoneNumber"`
	DiscordID   string `json:"discordID"`
	GithubID    string `json:"githubID"`
	NotionID    string `json:"notionID"`
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
		DiscordID:   employee.DiscordID,
		GithubID:    employee.GithubID,
		NotionID:    employee.NotionID,
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
		DiscordID:     employee.DiscordID,
		GithubID:      employee.GithubID,
		NotionID:      employee.NotionID,
	}
}
