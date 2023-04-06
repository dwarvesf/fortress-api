package model

import "github.com/shopspring/decimal"

type ProjectCommissionConfig struct {
	BaseModel

	ProjectID      UUID
	Position       HeadPosition
	CommissionRate decimal.Decimal
}

type ProjectCommissionConfigs []ProjectCommissionConfig

func (m *ProjectCommissionConfigs) ToMap() map[string]decimal.Decimal {
	rs := make(map[string]decimal.Decimal)
	for _, itm := range *m {
		rs[itm.Position.String()] = itm.CommissionRate
	}
	return rs
}
