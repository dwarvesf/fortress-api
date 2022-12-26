package view

import "github.com/dwarvesf/fortress-api/pkg/model"

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
