package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type Stack struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

func ToEmployeeStacks(stacks []model.EmployeeStack) []Stack {
	rs := make([]Stack, 0, len(stacks))
	for _, v := range stacks {
		r := Stack{
			ID:   v.Stack.ID.String(),
			Code: v.Stack.Code,
			Name: v.Stack.Name,
		}
		rs = append(rs, r)
	}

	return rs
}

func ToProjectStacks(stacks []model.ProjectStack) []Stack {
	rs := make([]Stack, 0, len(stacks))
	for _, v := range stacks {
		r := Stack{
			ID:   v.Stack.ID.String(),
			Code: v.Stack.Code,
			Name: v.Stack.Name,
		}
		rs = append(rs, r)
	}

	return rs
}
