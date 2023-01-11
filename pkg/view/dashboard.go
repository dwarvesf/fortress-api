package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type ProjectSizeResponse struct {
	Data []*model.ProjectSize `json:"data"`
}
