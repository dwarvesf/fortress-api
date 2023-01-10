package view

import (
	"sort"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type EngagementDashboardQuestionStat struct {
	Title     string     `json:"title"`
	StartDate *time.Time `json:"startDate"`
	Point     float64    `json:"point"`
}
type EngagementDashboard struct {
	Content    string                            `json:"content"`
	QuestionID string                            `json:"questionID"`
	Stats      []EngagementDashboardQuestionStat `json:"stats"`
}

type EngagementDashboardQuestionDetailStat struct {
	Field     string     `json:"field"`
	StartDate *time.Time `json:"startDate"`
	Point     float64    `json:"point"`
}

type EngagementDashboardDetail struct {
	QuestionID string                                  `json:"questionID"`
	Stats      []EngagementDashboardQuestionDetailStat `json:"stats"`
}

type GetDashboardResourceUtilizationResponse struct {
	Data []model.ResourceUtilization `json:"data"`
}

func ToEngagementDashboard(statistic []*model.StatisticEngagementDashboard) []EngagementDashboard {
	questionMapper := make(map[string][]EngagementDashboardQuestionStat)
	questionIDMapper := make(map[string]string)
	for _, s := range statistic {
		questionMapper[s.Content] = append(questionMapper[s.Content], EngagementDashboardQuestionStat{
			Title:     s.Title,
			StartDate: &s.StartDate,
			Point:     s.Point,
		})
		questionIDMapper[s.Content] = s.QuestionID.String()
	}

	dashboard := make([]EngagementDashboard, 0)

	for k, v := range questionMapper {
		sort.Slice(v, func(i, j int) bool {
			return v[i].StartDate.After(*v[j].StartDate)
		})
		dashboard = append(dashboard, EngagementDashboard{
			Content:    k,
			Stats:      v,
			QuestionID: questionIDMapper[k],
		})
	}
	return dashboard
}

func ToEngagementDashboardDetails(statistic []*model.StatisticEngagementDashboard) []EngagementDashboardDetail {
	questionMapper := make(map[string][]EngagementDashboardQuestionDetailStat)
	for _, s := range statistic {
		questionMapper[s.QuestionID.String()] = append(questionMapper[s.Content], EngagementDashboardQuestionDetailStat{
			Field:     s.Name,
			StartDate: &s.StartDate,
			Point:     s.Point,
		})
	}

	dashboard := make([]EngagementDashboardDetail, 0)

	for k, v := range questionMapper {
		dashboard = append(dashboard, EngagementDashboardDetail{
			QuestionID: k,
			Stats:      v,
		})
	}
	return dashboard
}

type GetEngagementDashboardResponse struct {
	Data *EngagementDashboard `json:"data"`
}

type GetEngagementDashboardDetailResponse struct {
	Data *EngagementDashboardDetail `json:"data"`
}
