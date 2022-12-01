package model

import (
	"time"
)

type ProjectCommission struct {
	BaseModel

	ProjectID    UUID
	CommissionID UUID
	ApplyFrom    time.Time
	ApplyTo      time.Time
	Percentage   JSON
}
