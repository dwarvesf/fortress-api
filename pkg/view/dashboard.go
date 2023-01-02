package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type GetDashboardResourceUtilizationResponse struct {
	Data []model.ResourceUtilization `json:"data"`
}
