package model

import "time"

type ProjectSize struct {
	ID   UUID   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
	Size int64  `json:"size"`
}

type WorkSurvey struct {
	EndDate  time.Time `json:"endDate"`
	Workload float64   `json:"workload"`
	Deadline float64   `json:"deadline"`
	Learning float64   `json:"learning"`
}
