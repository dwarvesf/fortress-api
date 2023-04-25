package basesalary

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, id string) (baseSalary *model.BaseSalary, err error)
	OneByEmployeeID(db *gorm.DB, employeeID string) (baseSalary *model.BaseSalary, err error)
	Save(db *gorm.DB, baseSalary *model.BaseSalary) (err error)
}
