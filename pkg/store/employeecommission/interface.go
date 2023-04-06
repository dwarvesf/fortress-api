package employeecommission

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, employeeCommissions []model.EmployeeCommission) ([]model.EmployeeCommission, error)
}
