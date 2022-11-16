package employeeposition

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, employeePosition *model.EmployeePosition) (*model.EmployeePosition, error)
	HardDelete(db *gorm.DB, employeeID string) (err error)
}
