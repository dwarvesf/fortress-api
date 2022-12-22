package view

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type AuthData struct {
	Employee    EmployeeData `json:"employee"`
	AccessToken string       `json:"accessToken"`
}

func ToAuthData(accessToken string, employee *model.Employee) *AuthData {
	return &AuthData{
		Employee:    *ToEmployeeData(employee),
		AccessToken: accessToken,
	}
}

type LoggedInUserData struct {
	ID          model.UUID `json:"id"`
	FullName    string     `json:"fullName"`
	DisplayName string     `json:"displayName"`
	Avatar      string     `json:"avatar"`
	TeamEmail   string     `json:"teamEmail"`
	Role        string     `json:"role"`
	Permissions []string   `json:"permissions"`
}

func ToAuthorizedUserData(employee *model.Employee, perms []*model.Permission) *LoggedInUserData {
	permissions := make([]string, len(perms))
	for i, p := range perms {
		permissions[i] = p.Code
	}

	return &LoggedInUserData{
		ID:          employee.ID,
		FullName:    employee.FullName,
		DisplayName: employee.DisplayName,
		Avatar:      employee.Avatar,
		TeamEmail:   employee.TeamEmail,
		Role:        employee.EmployeeRoles[0].Role.Name,
		Permissions: permissions,
	}
}

type AuthUserResponse struct {
	Data LoggedInUserData `json:"data"`
}
