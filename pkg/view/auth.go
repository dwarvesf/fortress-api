package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type AuthData struct {
	Employee    EmployeeListData `json:"employee"`
	AccessToken string           `json:"accessToken"`
}

func ToAuthData(accessToken string, employee *model.Employee) *AuthData {
	return &AuthData{
		Employee:    *ToEmployeeListData(employee),
		AccessToken: accessToken,
	}
}
