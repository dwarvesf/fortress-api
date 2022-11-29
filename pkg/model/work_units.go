package model

import (
	"gorm.io/datatypes"
)

type WorkUnit struct {
	BaseModel

	Name           string
	Status         WorkUnitStatus
	Type           WorkUnitType
	SourceURL      string
	SourceMetadata datatypes.JSON
	ProjectID      UUID

	WorkUnitMembers []*WorkUnitMember
	WorkUnitStacks  []*WorkUnitStack
}

type WorkUnitStatus string

const (
	WorkUnitStatusActive   WorkUnitStatus = "active"
	WorkUnitStatusArchived WorkUnitStatus = "archived"
)

func (e WorkUnitStatus) IsValid() bool {
	switch e {
	case
		WorkUnitStatusActive,
		WorkUnitStatusArchived:
		return true
	}
	return false
}

func (e WorkUnitStatus) String() string {
	return string(e)
}

type WorkUnitType string

const (
	WorkUnitTypeDevelopment WorkUnitType = "development"
	WorkUnitTypeManagement  WorkUnitType = "management"
	WorkUnitTypeTraining    WorkUnitType = "training"
	WorkUnitTypeLearning    WorkUnitType = "learning"
)

func (e WorkUnitType) IsValid() bool {
	switch e {
	case
		WorkUnitTypeDevelopment,
		WorkUnitTypeManagement,
		WorkUnitTypeTraining,
		WorkUnitTypeLearning:
		return true
	}
	return false
}

func (e WorkUnitType) String() string {
	return string(e)
}
