package project

type GetListProjectInput struct {
	Statuses            []string `json:"statuses"`
	Name                string   `json:"name"`
	Type                string   `json:"type"`
	AllowsSendingSurvey bool     `json:"allowsSendingSurvey"`
}
