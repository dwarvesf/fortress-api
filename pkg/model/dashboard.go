package model

type ProjectSize struct {
	ID   UUID   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
	Size int64  `json:"size"`
}
