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
	WorkUnit WorkUnit
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
