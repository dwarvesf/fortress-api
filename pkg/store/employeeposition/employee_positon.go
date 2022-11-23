package employeeposition

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Create create new one by id
func (s *store) Create(db *gorm.DB, employeePosition *model.EmployeePosition) (*model.EmployeePosition, error) {
	return employeePosition, db.Create(&employeePosition).Error
}

// Delete delete many EmployeePositions by employeeID
func (s *store) DeleteByEmployeeID(db *gorm.DB, employeeID string) error {
	return db.Unscoped().Where("employee_id = ?", employeeID).Delete(&model.EmployeePosition{}).Error
}
