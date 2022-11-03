package view

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// EmployeeListData view for listing data
type EmployeeListData struct {
	model.BaseModel

	// basic info
	FullName      string     `json:"full_name"`
	DisplayName   string     `json:"display_name"`
	TeamEmail     string     `json:"team_email"`
	PersonalEmail string     `json:"personal_email"`
	Avatar        string     `json:"avatar"`
	PhoneNumber   string     `json:"phone_number"`
	Address       string     `json:"address"`
	MBTI          string     `json:"mbti"`
	Gender        string     `json:"gender"`
	Horoscope     string     `json:"horoscope"`
	DateOfBirth   *time.Time `json:"birthday"`

	// working info
	EmploymentStatus model.EmploymentStatus `json:"status"`
	JoinedDate       *time.Time             `json:"joined_date"`
	LeftDate         *time.Time             `json:"left_date"`
}

func ToEmployeeListData(employee *model.Employee) *EmployeeListData {
	return &EmployeeListData{
		BaseModel: model.BaseModel{
			ID:        employee.ID,
			CreatedAt: employee.CreatedAt,
			UpdatedAt: employee.UpdatedAt,
		},
		FullName:         employee.FullName,
		DisplayName:      employee.DisplayName,
		TeamEmail:        employee.TeamEmail,
		PersonalEmail:    employee.PersonalEmail,
		Avatar:           employee.Avatar,
		PhoneNumber:      employee.PhoneNumber,
		Address:          employee.Address,
		MBTI:             employee.MBTI,
		EmploymentStatus: employee.EmploymentStatus,
		JoinedDate:       employee.JoinedDate,
		LeftDate:         employee.LeftDate,
	}
}
