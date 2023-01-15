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

type ActionItemReport struct {
	Quarter string `json:"quarter"`
	High    int64  `json:"high"`
	Medium  int64  `json:"medium"`
	Low     int64  `json:"low"`
}

type AverageEngineeringHealth struct {
	Quarter string  `json:"quarter"`
	Avg     float64 `json:"avg"`
}

type GroupEngineeringHealth struct {
	Quarter string  `json:"quarter"`
	Avg     float64 `json:"avg"`
	Area    string  `json:"area"`
}
