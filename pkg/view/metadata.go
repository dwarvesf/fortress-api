package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type WorkingStatusData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PositionResponse struct {
	Data []model.Position `json:"data"`
}
