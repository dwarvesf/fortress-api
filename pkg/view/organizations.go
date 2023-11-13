package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type Organization struct {
	ID     string `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
} // @name Organization

func ToOrganization(org *model.Organization) *Organization {
	if org == nil {
		return nil
	}

	return &Organization{
		ID:     org.ID.String(),
		Code:   org.Code,
		Name:   org.Name,
		Avatar: org.Avatar,
	}
}

func ToOrganizations(orgs []model.EmployeeOrganization) []Organization {
	rs := make([]Organization, 0, len(orgs))
	for _, v := range orgs {
		r := Organization{
			ID:     v.Organization.ID.String(),
			Code:   v.Organization.Code,
			Name:   v.Organization.Name,
			Avatar: v.Organization.Avatar,
		}
		rs = append(rs, r)
	}

	return rs
}
