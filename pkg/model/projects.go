package model

import (
	"time"

	"github.com/shopspring/decimal"
)

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
	Country   string
	Type      ProjectType
	StartDate *time.Time
	EndDate   *time.Time
	Status    ProjectStatus
	Slots     []ProjectSlot
	Members   []ProjectMember
	Heads     []ProjectHead
}

type DeploymentType string

const (
	MemberDeploymentTypeOfficial DeploymentType = "official"
	MemberDeploymentTypeShadow   DeploymentType = "shadow"
)

func (e DeploymentType) IsValid() bool {
	switch e {
	case
		MemberDeploymentTypeOfficial,
		MemberDeploymentTypeShadow:
		return true
	}
	return false
}

func (e DeploymentType) String() string {
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

type ProjectSlot struct {
	BaseModel

	ProjectID      UUID
	Position       string
	DeploymentType DeploymentType
	Rate           decimal.Decimal
	Discount       decimal.Decimal
	UpsellPersonID UUID
	SeniorityID    UUID
	Status         ProjectMemberStatus

	IsLead bool `gorm:"-"`

	Project              Project
	ProjectMember        ProjectMember
	ProjectSlotPositions []ProjectSlotPosition
}

type ProjectMember struct {
	BaseModel

	ProjectID      UUID
	EmployeeID     UUID
	ProjectSlotID  UUID
	JoinedDate     *time.Time
	LeftDate       *time.Time
	Status         ProjectMemberStatus
	Rate           decimal.Decimal
	Discount       decimal.Decimal
	DeploymentType DeploymentType
	UpsellPersonID UUID
	SeniorityID    UUID

	Employee               Employee
	Project                Project
	Seniority              Seniority
	ProjectMemberPositions []ProjectMemberPosition
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
	JoinedDate     time.Time
	LeftDate       *time.Time
	CommissionRate decimal.Decimal
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
