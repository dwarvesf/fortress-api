package model

import "time"

type ProjectType string

const (
	ProjectTypeDwarves      ProjectType = "dwarves"
	ProjectTypeFixedCost    ProjectType = "fixed-cost"
	ProjectTypeTimeMaterial ProjectType = "time-material"
)

func (e ProjectType) IsValid() bool {
	switch e {
	case
		ProjectTypeDwarves,
		ProjectTypeFixedCost,
		ProjectTypeTimeMaterial:
		return true
	}
	return false
}

func (e ProjectType) String() string {
	return string(e)
}

type ProjectStatus string

const (
	ProjectStatusOnBoarding ProjectStatus = "on-boarding"
	ProjectStatusActive     ProjectStatus = "active"
	ProjectStatusPaused     ProjectStatus = "paused"
	ProjectStatusClosed     ProjectStatus = "closed"
)

func (e ProjectStatus) IsValid() bool {
	switch e {
	case
		ProjectStatusOnBoarding,
		ProjectStatusActive,
		ProjectStatusPaused,
		ProjectStatusClosed:
		return true
	}
	return false
}

func (e ProjectStatus) String() string {
	return string(e)
}

type Project struct {
	BaseModel

	Name      string
	Type      ProjectType
	StartDate *time.Time
	EndDate   *time.Time
	Status    ProjectStatus
	Members   []ProjectMember
	Heads     []ProjectHead
}

type MemberDeploymentType string

const (
	MemberDeploymentTypeOfficial MemberDeploymentType = "official"
	MemberDeploymentTypeShadow   MemberDeploymentType = "shadow"
)

func (e MemberDeploymentType) IsValid() bool {
	switch e {
	case
		MemberDeploymentTypeOfficial,
		MemberDeploymentTypeShadow:
		return true
	}
	return false
}

func (e MemberDeploymentType) String() string {
	return string(e)
}

type ProjectMemberStatus string

const (
	ProjectMemberStatusPending    ProjectMemberStatus = "pending"
	ProjectMemberStatusOnBoarding ProjectMemberStatus = "on-boarding"
	ProjectMemberStatusActive     ProjectMemberStatus = "active"
	ProjectMemberStatusInactive   ProjectMemberStatus = "inactive"
)

func (e ProjectMemberStatus) IsValid() bool {
	switch e {
	case
		ProjectMemberStatusOnBoarding,
		ProjectMemberStatusActive,
		ProjectMemberStatusInactive,
		ProjectMemberStatusPending:
		return true
	}
	return false
}

func (e ProjectMemberStatus) String() string {
	return string(e)
}

type ProjectMember struct {
	BaseModel

	ProjectID      UUID
	EmployeeID     UUID
	JoinedDate     time.Time
	LeftDate       time.Time
	Position       string
	Status         ProjectMemberStatus
	DeploymentType MemberDeploymentType
	UpsellPersonID UUID
	Employee       Employee
}

type HeadPosition string

const (
	HeadPositionTechnicalLead   HeadPosition = "technical-lead"
	HeadPositionDeliveryManager HeadPosition = "delivery-manager"
	HeadPositionAccountManager  HeadPosition = "account-manager"
	HeadPositionSalePerson      HeadPosition = "sale-person"
)

func (e HeadPosition) IsValid() bool {
	switch e {
	case
		HeadPositionTechnicalLead,
		HeadPositionDeliveryManager,
		HeadPositionAccountManager,
		HeadPositionSalePerson:
		return true
	}
	return false
}

func (e HeadPosition) String() string {
	return string(e)
}

type ProjectHead struct {
	BaseModel

	ProjectID      UUID
	EmployeeID     UUID
	CommissionRate float64
	Position       HeadPosition
	Employee       Employee
}

func (p ProjectHead) IsLead() bool {
	return p.Position == HeadPositionTechnicalLead
}

func (p ProjectHead) IsAccountManager() bool {
	return p.Position == HeadPositionAccountManager
}

func (p ProjectHead) IsSalePerson() bool {
	return p.Position == HeadPositionSalePerson
}

func (p ProjectHead) IsDeliveryManager() bool {
	return p.Position == HeadPositionDeliveryManager
}

type ProjectStack struct {
	BaseModel

	ProjectID UUID
	StackID   UUID
}
