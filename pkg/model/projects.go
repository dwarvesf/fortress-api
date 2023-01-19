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

	Name                string          `gorm:"default:null"`
	CountryID           UUID            `gorm:"default:null"`
	Type                ProjectType     `gorm:"default:null"`
	StartDate           *time.Time      `gorm:"default:null"`
	EndDate             *time.Time      `gorm:"default:null"`
	Status              ProjectStatus   `gorm:"default:null"`
	ProjectEmail        string          `gorm:"default:null"`
	ClientEmail         string          `gorm:"default:null"`
	Avatar              string          `gorm:"default:null"`
	AllowsSendingSurvey bool            `gorm:"default:null"`
	Code                string          `gorm:"default:null"`
	Function            ProjectFunction `gorm:"default:null"`

	Country        *Country
	Slots          []ProjectSlot
	Heads          []*ProjectHead
	ProjectMembers []ProjectMember
	ProjectStacks  []ProjectStack
}

type DeploymentType string

const (
	MemberDeploymentTypeOfficial DeploymentType = "official"
	MemberDeploymentTypeShadow   DeploymentType = "shadow"
	MemberDeploymentTypePartTime DeploymentType = "part-time"
)

func (e DeploymentType) IsValid() bool {
	switch e {
	case
		MemberDeploymentTypeOfficial,
		MemberDeploymentTypeShadow,
		MemberDeploymentTypePartTime:
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
	SeniorityID    UUID
	UpsellPersonID UUID
	DeploymentType DeploymentType
	Status         ProjectMemberStatus
	Rate           decimal.Decimal
	Discount       decimal.Decimal

	Seniority            Seniority
	Project              Project
	ProjectMember        ProjectMember
	ProjectSlotPositions []ProjectSlotPosition
}

type ProjectMember struct {
	BaseModel

	ProjectID      UUID
	EmployeeID     UUID
	ProjectSlotID  UUID
	StartDate      *time.Time
	EndDate        *time.Time
	Status         ProjectMemberStatus
	Rate           decimal.Decimal
	Discount       decimal.Decimal
	DeploymentType DeploymentType
	UpsellPersonID UUID
	SeniorityID    UUID

	IsLead bool `gorm:"-"`

	Employee               Employee
	Project                Project
	Seniority              *Seniority
	ProjectMemberPositions []ProjectMemberPosition
	Positions              []Position `gorm:"-"`
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
	StartDate      time.Time
	EndDate        *time.Time
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

	Stack Stack
}

type ProjectFunction string

const (
	ProjectFunctionDevelopment ProjectFunction = "development"
	ProjectFunctionLearning    ProjectFunction = "learning"
	ProjectFunctionTraining    ProjectFunction = "training"
	ProjectFunctionManagement  ProjectFunction = "management"
)

func (e ProjectFunction) IsValid() bool {
	switch e {
	case
		ProjectFunctionDevelopment,
		ProjectFunctionLearning,
		ProjectFunctionTraining,
		ProjectFunctionManagement:
		return true
	}
	return false
}

func (e ProjectFunction) String() string {
	return string(e)
}

func IsUserActiveInProject(userID string, pm []ProjectMember) bool {
	for _, p := range pm {
		if p.EmployeeID.String() == userID && p.Status == ProjectMemberStatusActive {
			return true
		}
	}

	return false
}
