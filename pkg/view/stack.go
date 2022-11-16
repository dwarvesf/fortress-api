package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type Stack struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

func ToStacks(stacks []model.EmployeeStack) []Stack {
	rs := make([]Stack, 0, len(stacks))
	for _, v := range stacks {
		r := Stack{
			Code: v.Stack.Code,
			Name: v.Stack.Name,
		}
		rs = append(rs, r)
	}

	return rs
}
