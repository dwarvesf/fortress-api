package view

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// EmployeeData view for listing data
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
	DiscordID     string     `json:"discordID,omitempty"`
	GithubID      string     `json:"githubID,omitempty"`
	NotionID      string     `json:"notionID,omitempty"`

	// working info
	WorkingStatus model.WorkingStatus `json:"status"`
	JoinedDate    *time.Time          `json:"joinedDate"`
	LeftDate      *time.Time          `json:"leftDate"`

	AccountStatus model.AccountStatus   `json:"accountStatus"`
	Seniority     *model.Seniority      `json:"seniority"`
	Chapter       *model.Chapter        `json:"chapter"`
	LineManager   *BasisEmployeeInfo    `json:"lineManager"`
	Positions     []model.Position      `json:"positions"`
	Roles         []model.Role          `json:"roles"`
	Projects      []EmployeeProjectData `json:"projects"`
}

type BasisEmployeeInfo struct {
	ID       string `json:"id"`
	FullName string `json:"fullName"`
	Avatar   string `json:"avatar"`
}
type UpdateEmployeeStatusResponse struct {
	Data EmployeeData `json:"data"`
}

type EmployeeListDataResponse struct {
	Data []EmployeeData `json:"data"`
}

type UpdataEmployeeStatusResponse struct {
	Data EmployeeData `json:"data"`
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
	DiscordID     string     `json:"discordID"`
	GithubID      string     `json:"githubID"`
}

type ProfileDataResponse struct {
	Data ProfileData `json:"data"`
}

type EditEmployeeResponse struct {
	Data EmployeeData `json:"data"`
}

func ToEmployeeData(employee *model.Employee) *EmployeeData {
	projects := make([]EmployeeProjectData, 0, len(employee.ProjectMembers))
	for _, v := range employee.ProjectMembers {
		projects = append(projects, ToEmployeeProjectData(&v.Project))
	}

	rs := &EmployeeData{
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
		GithubID:      employee.GithubID,
		DiscordID:     employee.DiscordID,
		NotionID:      employee.NotionID,
		WorkingStatus: employee.WorkingStatus,
		Chapter:       employee.Chapter,
		Seniority:     employee.Seniority,
		JoinedDate:    employee.JoinedDate,
		LeftDate:      employee.LeftDate,
		AccountStatus: employee.AccountStatus,
		Projects:      projects,
		Positions:     employee.Positions,
		Roles:         employee.Roles,
	}

	if employee.Seniority != nil {
		rs.Seniority = employee.Seniority
	}

	if employee.Chapter != nil {
		rs.Chapter = employee.Chapter
	}

	if employee.LineManager != nil {
		rs.LineManager = &BasisEmployeeInfo{
			ID:       employee.LineManager.ID.String(),
			FullName: employee.LineManager.FullName,
			Avatar:   employee.LineManager.Avatar,
		}
	}

	return rs
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
