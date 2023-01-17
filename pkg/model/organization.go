package model

type Organization struct {
	BaseModel

	Name   string `json:"name"`
	Code   string `json:"code"`
	Avatar string `json:"avatar"`
}

// ToOrganizationMap create map from organizations
func ToOrganizationMap(organizations []*Organization) map[UUID]string {
	rs := map[UUID]string{}
	for _, s := range organizations {
		rs[s.ID] = s.Name
	}

	return rs
}
