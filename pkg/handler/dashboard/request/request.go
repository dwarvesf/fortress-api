package request

type GetEngagementDashboardDetailRequest struct {
	Filter    string `form:"filter" json:"filter"`
	StartDate string `form:"startDate" json:"startDate"`
}
