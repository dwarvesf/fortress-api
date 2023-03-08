package model

import (
	"text/template"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	// "git.d.foundation/fortress/ragnarok/src/service/vault"
	"github.com/dwarvesf/fortress-api/pkg/utils"
)

type Payroll struct {
	ID                  UUID           `sql:",type:uuid" json:"id"`
	UserID              UUID           `json:"user_id"`
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

func (Payroll) TableName() string { return "payroll" }

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

// ToPaidSuccessfulEmailContent to parse the payroll object
// into template when sending email after payroll is paid
func (p Payroll) GetPaidSuccessfulEmailFuncMap() map[string]interface{} {
	// the salary will be the contract(companyAccountAmount in DB)
	// plus the base salary(personalAccountAmount in DB)

	var addresses string
	addresses = "huy@d.foundation"
	// if vault.Get("run_mode") == "prod" {
	// 	addresses = "quang@d.foundation, accounting@d.foundation"
	// }

	return template.FuncMap{
		"ccList": func() string {
			return addresses
		},
		"userFirstName": func() string {
			return p.Employee.GetFirstNameFromFullName()
		},
		"currency": func() string {
			return p.Employee.BaseSalary.Currency.Symbol
		},
		"currencyName": func() string {
			return p.Employee.BaseSalary.Currency.Name
		},
		"formattedCurrentMonth": func() string {
			fm := time.Month(int(p.Month))
			return fm.String()
		},
		"formattedBaseSalaryAmount": func() string {
			return utils.FormatNumber(p.BaseSalaryAmount)
		},
		"formattedTotalAllowance": func() string {
			return utils.FormatNumber(int64(p.TotalAllowance))
		},
		"haveBonusOrCommission": func() bool {
			return len(p.CommissionExplains) > 0 || len(p.ProjectBonusExplains) > 0
		},
		"haveCommission": func() bool {
			return len(p.CommissionExplains) > 0
		},
		"haveBonus": func() bool {
			return len(p.ProjectBonusExplains) > 0
		},
		"commissionExplain": func() []CommissionExplain {
			return p.CommissionExplains
		},
		"projectBonusExplains": func() []ProjectBonusExplain {
			return p.ProjectBonusExplains
		},
	}
}

// Batch enumeration
type Batch int

const (
	// FirstBatch represent payroll batch that due date in date: 1st of a month
	FirstBatch Batch = 1

	// SecondBatch represent payroll batch that due date in date: 15th of a month
	SecondBatch = 15
)
