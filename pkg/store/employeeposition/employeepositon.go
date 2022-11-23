package employeeposition

// TODO: format FILENAME thanh employee_position.go

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
}

func New() IStore {
	return &store{}
}

// Create create new one by id
func (s *store) Create(db *gorm.DB, employeePosition *model.EmployeePosition) (*model.EmployeePosition, error) {
	return employeePosition, db.Create(&employeePosition).Error
}

// HardDelete hard delete one by id
func (s *store) HardDelete(db *gorm.DB, employeeID string) error {
	return db.Table("employee_positions").Unscoped().Where("employee_id = ?", employeeID).Delete(&model.EmployeePosition{}).Error
}
