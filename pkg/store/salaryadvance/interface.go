package salaryadvance

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, id string) (salaryAdvance *model.SalaryAdvance, err error)
	ListNotPayBackByEmployeeID(db *gorm.DB, employeeID string) (salaryAdvance []model.SalaryAdvance, err error)
	Save(db *gorm.DB, salaryAdvance *model.SalaryAdvance) (err error)
	ListAggregatedSalaryAdvance(db *gorm.DB, idPaid *bool, paging model.Pagination, order model.SortOrder) (report []model.AggregatedSalaryAdvance, err error)
	TotalAggregatedSalaryAdvance(db *gorm.DB, idPaid *bool) (count, totalIcy int64, totalUSD float64, err error)
}
