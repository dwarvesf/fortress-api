package view

import (
	"encoding/json"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type MetaData struct {
	ID   string `json:"id"`
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
	Data []model.Stack `json:"data"`
}

type RolesResponse struct {
	Data []model.Role `json:"data"`
}

type PositionResponse struct {
	Data []model.Position `json:"data"`
}

type CountriesResponse struct {
	Data []Country `json:"data"`
}

type CitiesResponse struct {
	Data []string `json:"data"`
}

type OrganizationsResponse struct {
	Data []model.Organization `json:"data"`
}

// Question model question for get list question api
type Question struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	Subcategory string `json:"subcategory"`
	Content     string `json:"content"`
	Order       int64  `json:"order"`
}

// GetQuestionResponse response for get question api
type GetQuestionResponse struct {
	Data []Question `json:"data"`
}

func ToListQuestion(questions []*model.Question) []*Question {
	var rs []*Question
	for _, q := range questions {
		rs = append(rs, &Question{
			ID:          q.ID.String(),
			Category:    q.Category.String(),
			Subcategory: q.Subcategory.String(),
			Content:     q.Content,
			Type:        q.Type.String(),
			Order:       q.Order,
		})
	}

	return rs
}

func ToRoles(roles []*model.Role) []*Role {
	var rs []*Role
	for _, r := range roles {
		rs = append(rs, &Role{
			ID:   r.ID.String(),
			Code: r.Code,
			Name: toRoleName(r),
		})
	}

	return rs
}

type Country struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	Cities []City `json:"cities"`
}

type City struct {
	Name string `json:"name"`
	Lat  string `json:"lat"`
	Long string `json:"long"`
}

func ToCountryView(country []*model.Country) ([]Country, error) {
	var rs []Country
	for _, c := range country {
		var cities []City
		err := json.Unmarshal(c.Cities, &cities)
		if err != nil {
			return nil, err
		}
		for _, c := range cities {
			cities = append(cities, City{
				Name: c.Name,
				Lat:  c.Lat,
				Long: c.Long,
			})
		}

		rs = append(rs, Country{
			ID:     c.ID.String(),
			Name:   c.Name,
			Code:   c.Code,
			Cities: cities,
		})
	}
	return rs, nil
}
