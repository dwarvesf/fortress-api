package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type Role struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
} // @name Role

func ToEmployeeRoles(roles []model.EmployeeRole) []Role {
	rs := make([]Role, 0, len(roles))
	for _, v := range roles {
		r := Role{
			ID:   v.Role.ID.String(),
			Code: v.Role.Code,
			Name: toRoleName(&v.Role),
		}
		rs = append(rs, r)
	}

	return rs
}

func toRoleName(role *model.Role) string {
	roleName := ""
	switch role.Color {
	case "red":
		roleName = "🔴 " + role.Name
	case "yellow":
		roleName = "🟡 " + role.Name
	case "green":
		roleName = "🟢 " + role.Name
	default:
		roleName = "⚪️ " + role.Name
	}

	return roleName
}
