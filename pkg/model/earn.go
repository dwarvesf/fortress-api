package model

import "time"

type Earn struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Reward   int        `json:"reward"`
	Progress int        `json:"progress"`
	Priority string     `json:"priority"`
	Tags     []string   `json:"tags"`
	PICs     []Employee `json:"pics"`
	Status   string     `json:"status"`
	Function []string   `json:"function"`
	DueDate  *time.Time `json:"due_date"`
}
