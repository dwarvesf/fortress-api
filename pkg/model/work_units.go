package model

import (
	"gorm.io/datatypes"
)

type WorkUnit struct {
	BaseModel

	Name           string
	Status         string
	Type           string
	SourceURL      string
	SourceMetadata datatypes.JSON
	EmployeeID     UUID
}

type WorkUnitType string

const (
	WorkUnitTypeDevelopment WorkUnitType = "development"
	WorkUnitTypeManagement  WorkUnitType = "management"
	WorkUnitTypeTraining    WorkUnitType = "training"
	WorkUnitTypeLearning    WorkUnitType = "learning"
)

func (e WorkUnitType) Valid() bool {
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
