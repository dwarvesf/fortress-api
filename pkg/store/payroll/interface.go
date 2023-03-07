package payroll

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// IStore implement operation method
type IStore interface {
	// TODO: rename to GetList
	GetSalary(db *gorm.DB, year int, month int) ([]model.Employee, error)
	Get(db *gorm.DB, userId string, year, month int) (*model.Payroll, error)
	Create(db *gorm.DB, p *model.Payroll) error
	InsertList(db *gorm.DB, payrolls []model.Payroll) error

	// GetList payroll row, the result included: User,
	// Role, Rank, Base Salary and Commission
	GetList(db *gorm.DB, q GetListPayrollInput) ([]model.Payroll, error)

	// UpdateSpecificFields for a payroll row
	// fields that will be update will be declared
	// as string match with column name in table payroll
	UpdateSpecificFields(db *gorm.DB, id string, fields map[string]interface{}) error
	GetLatestCommitTime(db *gorm.DB) (time.Time, error)
}
