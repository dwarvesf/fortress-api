package model

type AccountStatus struct {
	BaseModel

	Name string `json:"name"`
	Code string `json:"code"`
}
