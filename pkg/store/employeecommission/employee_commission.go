package employeecommission

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Create make new one by id
func (s *store) Create(db *gorm.DB, employeeCommissions []model.EmployeeCommission) ([]model.EmployeeCommission, error) {
	return employeeCommissions, db.Create(&employeeCommissions).Error
}
