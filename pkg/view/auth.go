package view

import "github.com/dwarvesf/fortress-api/pkg/model"

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
