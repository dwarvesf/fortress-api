package model

import (
	"time"
)

type WorkUnitMember struct {
	BaseModel

	Status     string
	StartDate  time.Time
	EndDate    *time.Time
	EmployeeID UUID
	WorkUnitID UUID
	ProjectID  UUID

	Employee Employee
}

type WorkUnitPeer struct {
	EmployeeID UUID
	ReviewerID UUID
}

type WorkUnitMemberStatus string

const (
	WorkUnitMemberStatusActive   WorkUnitMemberStatus = "active"
	WorkUnitMemberStatusInactive WorkUnitMemberStatus = "inactive"
)

func (e WorkUnitMemberStatus) IsValid() bool {
	switch e {
	case
		WorkUnitMemberStatusActive,
		WorkUnitMemberStatusInactive:
		return true
	}
	return false
}

func (e WorkUnitMemberStatus) String() string {
	return string(e)
}
