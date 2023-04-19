package model

import "time"

type ProjectSize struct {
	ID     UUID   `json:"id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	Avatar string `json:"avatar"`
	Size   int64  `json:"size"`
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

type ResolvedActionItem struct {
	ID      UUID   `json:"id"`
	Quarter string `json:"quarter"`
	Count   int64  `json:"count"`
}

type AuditSummary struct {
	Quarter string  `json:"quarter"`
	ID      UUID    `json:"id"`
	Name    string  `json:"name"`
	Avatar  string  `json:"avatar"`
	Size    int64   `json:"size"`
	Code    string  `json:"code"`
	High    int64   `json:"high"`
	Medium  int64   `json:"medium"`
	Low     int64   `json:"low"`
	Done    int64   `json:"done"`
	Health  float64 `json:"health"`
	Audit   float64 `json:"audit"`
}

type ResourceUtilization struct {
	Date      time.Time `json:"date"`
	Staffed   int       `json:"staffed"`
	Internal  int       `json:"internal"`
	Available int       `json:"available"`
}

type TotalWorkUnitDistribution struct {
	TotalLineManagerCount int64
	TotalProjectHead      int64
	TotalLearning         int64
	TotalDevelopment      int64
	TotalManagement       int64
	TotalTraining         int64
}

type ManagementInfo struct {
	ID       UUID          `json:"id"`
	Name     string        `json:"name"`
	Code     string        `json:"code"`
	Avatar   string        `json:"avatar"`
	Type     ProjectType   `json:"type"`
	Status   ProjectStatus `json:"status"`
	Position HeadPosition  `json:"position"`
}
