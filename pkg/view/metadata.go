package view

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type MetaData struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
} // @name MetaData

type MetadataResponse struct {
	Data []MetaData `json:"data"`
} // @name MetaDataResponse

type SeniorityResponse struct {
	Data []Seniority `json:"data"`
} // @name SeniorityResponse

type ChapterResponse struct {
	Data []Chapter `json:"data"`
} // @name ChapterResponse

type StackResponse struct {
	PaginationResponse
	Data []Stack `json:"data"`
} // @name StackResponse

type RolesResponse struct {
	Data []Role `json:"data"`
} // @name RolesResponse

type PositionResponse struct {
	Data []Position `json:"data"`
} // @name PositionResponse

type CountriesResponse struct {
	Data []Country `json:"data"`
} // @name CountriesResponse

type CitiesResponse struct {
	Data []string `json:"data"`
} // @name CitiesResponse

type OrganizationsResponse struct {
	Data []Organization `json:"data"`
} // @name OrganizationsResponse

// Question model question for get list question api
type Question struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	Subcategory string `json:"subcategory"`
	Content     string `json:"content"`
	Order       int64  `json:"order"`
} // @name Question

// GetQuestionResponse response for get question api
type GetQuestionResponse struct {
	Data []Question `json:"data"`
} // @name GetQuestionResponse

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
} // @name Country

type City struct {
	Name string `json:"name"`
	Lat  string `json:"lat"`
	Long string `json:"long"`
} // @name City

func ToCountry(c *model.Country) *Country {
	if c == nil {
		return nil
	}

	cities := make([]City, len(c.Cities))
	for i, city := range c.Cities {
		cities[i] = City{
			Name: city.Name,
			Lat:  city.Lat,
			Long: city.Long,
		}
	}

	return &Country{
		ID:     c.ID.String(),
		Name:   c.Name,
		Code:   c.Code,
		Cities: cities,
	}
}

func ToCountryView(countries []*model.Country) ([]Country, error) {
	var rs []Country
	for _, c := range countries {
		cities := make([]City, len(c.Cities))
		for i, city := range c.Cities {
			cities[i] = City{
				Name: city.Name,
				Lat:  city.Lat,
				Long: city.Long,
			}
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
