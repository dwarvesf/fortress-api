package model

import "time"

type Earn struct {
	ID       string
	Name     string
	Reward   int
	Progress int
	Priority string
	Tags     []string
	PICs     []Employee
	Status   string
	Function []string
	DueDate  *time.Time
}
