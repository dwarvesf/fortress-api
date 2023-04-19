package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	// "git.d.foundation/fortress/ragnarok/src/service/vault"
)

type Payroll struct {
	ID                  UUID           `sql:",type:uuid" json:"id"`
	EmployeeID          UUID           `json:"employee_id"`
	BaseSalaryAmount    int64          `json:"base_salary_amount"`
	ContractAmount      int64          `json:"contract_amount"`
	Total               VietnamDong    `json:"total"`
	ConversionAmount    VietnamDong    `json:"conversion_amount"`
	Month               int64          `json:"month"`
	Year                int64          `json:"year"`
	CommissionAmount    VietnamDong    `json:"commission_amount"`
	CommissionExplain   datatypes.JSON `json:"commission_explain"`
	UserRankSnapshot    datatypes.JSON `json:"user_rank_snapshot"`
	TotalExplain        datatypes.JSON `json:"total_explain"`
	ProjectBonusAmount  VietnamDong    `json:"project_bonus"`
	ProjectBonusExplain datatypes.JSON `json:"project_bonus_explain"`
	DueDate             *time.Time     `json:"due_date"`
	IsPaid              bool           `json:"is_paid"`

	Employee Employee `json:"employee"`

	TotalAllowance       float64               `json:"total_allowance" gorm:"-"`        // TotalAllowance is amount of allowance in email template
	CommissionExplains   []CommissionExplain   `json:"commission_explains" gorm:"-"`    // CommissionExplains is commission explains in email template
	ProjectBonusExplains []ProjectBonusExplain `json:"project_bonus_explains" gorm:"-"` // ProjectBonusExplains is project bonus explains in email template
	TWAmount             float64               `json:"twAmount" gorm:"-"`               // TotalAllowance is amount of allowance in email template
	TWRate               float64               `json:"twRate" gorm:"-"`                 // TWRate is rate of allowance in email template
	TWFee                float64               `json:"twFee" gorm:"-"`
}

func (p *Payroll) BeforeCreate(tx *gorm.DB) error {
	p.ID = NewUUID()
	return nil
}

func (Payroll) TableName() string { return "payrolls" }

// ProjectBonusExplain  explain where and when
// the project bonus come from
// in each row of table payroll
type ProjectBonusExplain struct {
	Name             string      `json:"name"`
	Month            int         `json:"month"`
	Year             int         `json:"year"`
	Amount           VietnamDong `json:"amount"`
	FormattedAmount  string      `json:"formatted_amount"`
	Description      string      `json:"description"`
	BasecampTodoID   int         `json:"todo_id"`
	BasecampBucketID int         `json:"bucket_id"`
}

// CommissionExplain  explain where and when
// the commission come from
// in each row of table payroll
type CommissionExplain struct {
	ID               UUID        `json:"id"`
	Name             string      `json:"name"`
	Month            int         `json:"month"`
	Year             int         `json:"year"`
	Amount           VietnamDong `json:"amount"`
	FormattedAmount  string      `json:"formatted_amount"`
	BasecampTodoID   int         `json:"todo_id"`
	BasecampBucketID int         `json:"bucket_id"`
}

// Batch enumeration
type Batch int

const (
	// FirstBatch represent payroll batch that due date in date: 1st of a month
	FirstBatch Batch = 1

	// SecondBatch represent payroll batch that due date in date: 15th of a month
	SecondBatch = 15
)
