package employeebonus

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	GetByEmployeeID(db *gorm.DB, employee_id model.UUID) ([]model.EmployeeBonus, error)
}
