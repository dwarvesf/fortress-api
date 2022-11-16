package model

type Stack struct {
	BaseModel

	Name string `json:"name"`
	Code string `json:"code"`
}

// ToStackMap create map from stacks
func ToStackMap(stacks []*Stack) map[UUID]string {
	rs := map[UUID]string{}
	for _, s := range stacks {
		rs[s.ID] = s.Name
	}

	return rs
}
