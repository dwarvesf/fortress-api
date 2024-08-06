package model

import "github.com/shopspring/decimal"

type CommissionModel struct {
	Beneficiary    BasicEmployeeInfo
	CommissionType string
	CommissionRate decimal.Decimal
	Description    string
	Sub            *CommissionModel
}

type BasicEmployeeInfo struct {
	ID          string
	FullName    string
	DisplayName string
	Avatar      string
	Username    string
	ReferredBy  string
}
