package payroll

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	GetLatestCommitTime(tx *gorm.DB) (*time.Time, error)
	List(tx *gorm.DB, q PayrollInput) (payrolls []model.Payroll, err error)
	ListDashboardPayrolls(tx *gorm.DB, q PayrollDashboardInput) (payrolls []model.Payroll, err error)
	Save(tx *gorm.DB, pr []model.Payroll) error
}
