package employeestack

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, employeeStack *model.EmployeeStack) (*model.EmployeeStack, error)
	DeleteByEmployeeID(db *gorm.DB, employeeID string) error
}
