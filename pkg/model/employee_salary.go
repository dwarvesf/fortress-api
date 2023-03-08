package model

import "github.com/jinzhu/gorm/dialects/postgres"

const (
	rateGem = 20
	rateUSD = 22500
)

type EmployeeSalary struct {
	ID UUID `json:"-"`

	EmployeeID UUID      `json:"employee_id"`
	Employee   *Employee `json:"-"`

	CommissionAmount    int64          `json:"commission_amount"`
	CommissionDetail    postgres.Jsonb `json:"commission_detail"`
	ReimbursementAmount int64          `json:"reimbursement_amount"`
	ReimbursementDetail postgres.Jsonb `json:"reimbursement_detail"`
	BonusAmount         int64          `json:"bonus_amount"`
	BonusDetail         postgres.Jsonb `json:"bonus_detail"`
	TotalAmount         int64          `json:"total_amount"`

	Month        uint8 `json:"month"`
	Year         int32 `json:"year"`
	ActualPayDay int8  `json:"actual_pay_day"`
	PlanPayDay   int8  `json:"plan_pay_day"`
	IsDone       bool  `json:"is_done"`
}
