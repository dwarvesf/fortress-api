package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type MetaData struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type SeniorityResponse struct {
	Data []model.Seniority `json:"data"`
}

type ChapterResponse struct {
	Data []model.Chapter `json:"data"`
}

type StackResponse struct {
	Data []model.Chapter `json:"data"`
}

type AccountRoleResponse struct {
	Data []model.Role `json:"data"`
}

type PositionResponse struct {
	Data []model.Position `json:"data"`
}

type CountriesResponse struct {
	Data []model.Country `json:"data"`
}

type CitiesResponse struct {
	Data []string `json:"data"`
}
