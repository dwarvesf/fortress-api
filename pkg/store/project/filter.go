package project

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type GetListProjectInput struct {
	Status string `json:"status"`
	Name   string `json:"name"`
	Type   string `json:"type"`
}

type UpdateGeneralInfoInput struct {
	Name      string
	StartDate *time.Time
	CountryID model.UUID
	Stacks    []*model.UUID
}
