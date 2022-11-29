package model

import (
	"time"
)

type WorkUnitMember struct {
	BaseModel

	Status     string
	JoinedDate time.Time
	LeftDate   *time.Time
	EmployeeID UUID
	WorkUnitID UUID
	ProjectID  UUID

	Employee Employee
}
