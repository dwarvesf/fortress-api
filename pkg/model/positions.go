package model

type Position struct {
	BaseModel

	Name string `json:"name"`
	Code string `json:"code"`
}
