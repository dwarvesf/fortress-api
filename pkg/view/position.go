package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type Position struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func ToPositions(pos []model.EmployeePosition) []Position {
	rs := make([]Position, 0, len(pos))
	for _, v := range pos {
		r := Position{
			Code: v.Position.Code,
			Name: v.Position.Name,
		}
		rs = append(rs, r)
	}

	return rs
}

func ToProjectSlotPositions(pos []model.ProjectSlotPosition) []Position {
	rs := make([]Position, 0, len(pos))
	for _, v := range pos {
		r := Position{
			Code: v.Position.Code,
			Name: v.Position.Name,
		}
		rs = append(rs, r)
	}

	return rs
}

func ToProjectMemberPositions(pos []model.ProjectMemberPosition) []Position {
	rs := make([]Position, 0, len(pos))
	for _, v := range pos {
		r := Position{
			Code: v.Position.Code,
			Name: v.Position.Name,
		}
		rs = append(rs, r)
	}

	return rs
}
