package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type Role struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

func ToRoles(roles []model.EmployeeRole) []Role {
	rs := make([]Role, 0, len(roles))
	for _, v := range roles {
		r := Role{
			ID:   v.Role.ID.String(),
			Code: v.Role.Code,
			Name: v.Role.Name,
		}
		rs = append(rs, r)
	}

	return rs
}
