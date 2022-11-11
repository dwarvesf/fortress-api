package view

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// EmployeeListData view for listing data
type EmployeeData struct {
	model.BaseModel

	// basic info
	FullName      string     `json:"fullName"`
	DisplayName   string     `json:"displayName"`
	TeamEmail     string     `json:"teamEmail"`
	PersonalEmail string     `json:"personalEmail"`
	Avatar        string     `json:"avatar"`
	PhoneNumber   string     `json:"phoneNumber"`
	Address       string     `json:"address"`
	MBTI          string     `json:"mbti"`
	Gender        string     `json:"gender"`
	Horoscope     string     `json:"horoscope"`
	DateOfBirth   *time.Time `json:"birthday"`

	// working info
	WorkingStatus model.WorkingStatus `json:"status"`
	JoinedDate    *time.Time          `json:"joinedDate"`
	LeftDate      *time.Time          `json:"leftDate"`
}
type UpdateEmployeeStatusResponse struct {
	Data EmployeeData `json:"data"`
}

type EmployeeListDataResponse struct {
	Data []EmployeeData `json:"data"`
}

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
	DiscordID     string     `json:"discordId"`
	GithubID      string     `json:"githubId"`
}

type ProfileDataResponse struct {
	Data ProfileData `json:"data"`
}

func ToEmployeeData(employee *model.Employee) *EmployeeData {
	return &EmployeeData{
		BaseModel: model.BaseModel{
			ID:        employee.ID,
			CreatedAt: employee.CreatedAt,
			UpdatedAt: employee.UpdatedAt,
		},
		FullName:      employee.FullName,
		DisplayName:   employee.DisplayName,
		TeamEmail:     employee.TeamEmail,
		PersonalEmail: employee.PersonalEmail,
		Avatar:        employee.Avatar,
		PhoneNumber:   employee.PhoneNumber,
		Address:       employee.Address,
		MBTI:          employee.MBTI,
		Gender:        employee.Gender,
		Horoscope:     employee.Horoscope,
		DateOfBirth:   employee.DateOfBirth,
		WorkingStatus: employee.WorkingStatus,
		JoinedDate:    employee.JoinedDate,
		LeftDate:      employee.LeftDate,
	}
}

func ToEmployeeListData(employees []*model.Employee) []EmployeeData {
	rs := make([]EmployeeData, 0, len(employees))
	for _, emp := range employees {
		empRes := ToEmployeeData(emp)
		rs = append(rs, *empRes)
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
	}
}
