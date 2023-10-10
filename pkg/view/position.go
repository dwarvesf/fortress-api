package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type Position struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
} // @name Position

func ToPosition(pos *model.Position) *Position {
	if pos == nil {
		return nil
	}

	return &Position{
		ID:   pos.ID.String(),
		Code: pos.Code,
		Name: pos.Name,
	}
}

func ToPositions(pos []model.Position) []Position {
	rs := make([]Position, 0, len(pos))
	for _, p := range pos {
		r := Position{
			ID:   p.ID.String(),
			Code: p.Code,
			Name: p.Name,
		}
		rs = append(rs, r)
	}

	return rs
}

func ToEmployeePositions(pos []model.EmployeePosition) []Position {
	rs := make([]Position, 0, len(pos))
	for _, v := range pos {
		r := Position{
			ID:   v.Position.ID.String(),
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
			ID:   v.Position.ID.String(),
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
			ID:   v.Position.ID.String(),
			Code: v.Position.Code,
			Name: v.Position.Name,
		}
		rs = append(rs, r)
	}

	return rs
}
