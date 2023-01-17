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

type GroupAudit struct {
	Quarter    string  `json:"quarter"`
	Frontend   float64 `json:"frontend"`
	Backend    float64 `json:"backend"`
	System     float64 `json:"system"`
	Process    float64 `json:"process"`
	Blockchain float64 `json:"blockchain"`
	Mobile     float64 `json:"mobile"`
}

type AverageAudit struct {
	Quarter string  `json:"quarter"`
	Avg     float64 `json:"avg"`
}

type ActionItemSquashReport struct {
	SnapDate time.Time `json:"snapDate"`
	All      int64     `json:"all"`
	High     int64     `json:"high"`
	Medium   int64     `json:"medium"`
	Low      int64     `json:"low"`
}
