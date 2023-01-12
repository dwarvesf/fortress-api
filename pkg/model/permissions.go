package model

type Permission struct {
	BaseModel

	Code string
	Name string
}

func ToPermissionMap(perms []*Permission) map[string]string {
	m := make(map[string]string)
	for _, v := range perms {
		m[v.Code] = v.Name
	}

	return m
}
