package request

import "github.com/dwarvesf/fortress-api/pkg/model"

type WorkSurveysInput struct {
	ProjectID string `json:"projectID" form:"projectID"`
}
type ActionItemInput struct {
	ProjectID string `json:"projectID" form:"projectID"`
}
type GetEngagementDashboardDetailRequest struct {
	Filter    string `form:"filter" json:"filter"`
	StartDate string `form:"startDate" json:"startDate"`
}

type WorkUnitDistributionInput struct {
	Sort model.SortOrder    `json:"sort" form:"sort"`
	Type model.WorkUnitType `json:"type" form:"type"`
	Name string             `json:"name" form:"name"`
}
