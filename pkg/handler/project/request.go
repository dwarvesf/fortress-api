package project

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type GetListProjectInput struct {
	model.Pagination

	Name   string `form:"name" json:"name"`
	Status string `form:"status" json:"status"`
	Type   string `form:"type" json:"type"`
}

type updateAccountStatusBody struct {
	ProjectStatus model.ProjectStatus `json:"status"`
}

func (i *GetListProjectInput) Validate() error {
	if i.Type != "" && !model.ProjectType(i.Type).IsValid() {
		return ErrInvalidProjectType
	}

	if i.Status != "" && !model.ProjectStatus(i.Status).IsValid() {
		return ErrInvalidProjectStatus
	}
	return nil
}
