package request

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
