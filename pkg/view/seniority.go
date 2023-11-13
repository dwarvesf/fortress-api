package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type Seniority struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
} // @name Seniority

func ToSeniority(seniority model.Seniority) Seniority {
	return Seniority{
		ID:   seniority.ID.String(),
		Code: seniority.Code,
		Name: seniority.Name,
	}
}
