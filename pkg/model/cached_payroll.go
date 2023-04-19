package model

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CachedPayroll struct {
	ID       UUID           `sql:",type:uuid" json:"id"`
	Month    int            `json:"month"`
	Year     int            `json:"year"`
	Batch    int            `json:"batch"`
	Payrolls datatypes.JSON `json:"payrolls"`
}

func (p *CachedPayroll) BeforeCreate(scope *gorm.DB) error {
	p.ID = NewUUID()
	return nil
}

func (CachedPayroll) TableName() string { return "cached_payrolls" }
