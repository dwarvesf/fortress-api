package employeerole

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, er *model.EmployeeRole) (employeeRole *model.EmployeeRole, err error)
	HardDeleteByEmployeeID(db *gorm.DB, employeeID string) (err error)
}
