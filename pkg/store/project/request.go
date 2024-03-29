package project

type GetListProjectInput struct {
	Statuses            []string `json:"statuses"`
	Name                string   `json:"name"`
	Types               []string `json:"type"`
	AllowsSendingSurvey bool     `json:"allowsSendingSurvey"`
}
