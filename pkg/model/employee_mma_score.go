package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// EmployeeMMAScore define the model for table employee_mma_scores
type EmployeeMMAScore struct {
	BaseModel

	EmployeeID    UUID
	MasteryScore  decimal.Decimal
	AutonomyScore decimal.Decimal
	MeaningScore  decimal.Decimal
	RatedAt       *time.Time
}

type EmployeeMMAScoreData struct {
	EmployeeID    UUID
	FullName      string
	MMAID         UUID
	MasteryScore  decimal.Decimal
	AutonomyScore decimal.Decimal
	MeaningScore  decimal.Decimal
	RatedAt       *time.Time
}
