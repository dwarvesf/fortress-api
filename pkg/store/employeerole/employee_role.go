package employeerole

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
}

func New() IStore {
	return &store{}
}

// Create using fore create new Employee role
func (s *store) Create(db *gorm.DB, er *model.EmployeeRole) (*model.EmployeeRole, error) {
	return er, db.Create(&er).Error
}

// HardDeleteByEmployeeID hard delete employeeRole by employeeID
func (s *store) HardDeleteByEmployeeID(db *gorm.DB, employeeID string) error {
	return db.Unscoped().Where("employee_id = ?", employeeID).Delete(&model.EmployeeRole{}).Error
}
