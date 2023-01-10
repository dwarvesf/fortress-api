package model

type Seniority struct {
	BaseModel

	Name  string `json:"name"`
	Code  string `json:"code"`
	Level int    `json:"level"`
}
