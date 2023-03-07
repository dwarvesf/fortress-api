package model

const (
	rateGem = 20
	rateUSD = 22500
)

type EmployeeSalary struct {
	ID UUID `json:"-"`

	EmployeeID UUID      `json:"employee_id"`
	Employee   *Employee `json:"-"`

	BaseZap          int64  `json:"base_zap"`
	BaseGem          int64  `json:"base_gem"`
	AdjustmentSalary int64  `json:"adjustment_salary"`
	AdjustmentReason string `json:"adjustment_reason"`
	BonusZap         int64  `json:"bonus_zap"`
	BonusGem         int64  `json:"bonus_gem"`
	LossZap          int64  `json:"loss_zap"`
	LossGem          int64  `json:"loss_gem"`
	TotalSalary      int64  `json:"total_salary"`
	Month            uint8  `json:"month"`
	Year             int32  `json:"year"`
	ActualPayDay     int8   `json:"actual_pay_day"`
	PlanPayDay       int8   `json:"plan_pay_day"`
	IsDone           bool   `json:"is_done"`
}
