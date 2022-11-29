package model

import (
	"time"
)

type ProjectUnitMember struct {
	BaseModel

	Name       string
	Status     string
	LeftDate   *time.Time
	JoinedDate *time.Time
	EmployeeID UUID
	WorkUnitID UUID
	ProjectID  UUID
}
