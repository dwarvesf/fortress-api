package model

import "time"

type OperationalService struct {
	BaseModel

	Name         string
	Amount       int
	Note         string
	RegisterDate time.Time
	StartAt      time.Time
	EndAt        time.Time
	IsActive     bool
	CurrencyID   UUID
	Currency     *Currency `gorm:"foreignKey:CurrencyID;references:ID"`
}
