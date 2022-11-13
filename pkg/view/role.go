package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type RoleData struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

func ToRoleListData(roles []model.Role) []RoleData {
	result := make([]RoleData, 0, len(roles))
	for _, r := range roles {
		role := RoleData{
			Name: r.Name,
			Code: r.Code,
		}
		result = append(result, role)
	}

	return result
}
