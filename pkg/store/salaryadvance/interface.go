package salaryadvance

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, id string) (salaryAdvance *model.SalaryAdvance, err error)
	ListNotPayBackByEmployeeID(db *gorm.DB, employeeID string) (salaryAdvance []model.SalaryAdvance, err error)
	Save(db *gorm.DB, salaryAdvance *model.SalaryAdvance) (err error)
}
