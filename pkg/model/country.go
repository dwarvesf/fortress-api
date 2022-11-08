package model

type Country struct {
	BaseModel

	Name   string          `json:"name"`
	Code   string          `json:"code"`
	Cities JSONArrayString `json:"cities"`
}
