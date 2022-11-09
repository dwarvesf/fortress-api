package model

type Chapter struct {
	BaseModel

	Name string `json:"name"`
	Code string `json:"code"`
}
