package employeestack

// TODO: format FILENAME thanh employee_stack.go

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
func (s *store) Create(db *gorm.DB, employeeStack *model.EmployeeStack) (*model.EmployeeStack, error) {
	return employeeStack, db.Create(&employeeStack).Error
}

// HardDelete hard delete one by id
func (s *store) HardDelete(db *gorm.DB, employeeID string) error {
	return db.Table("employee_stacks").Unscoped().Where("employee_id = ?", employeeID).Delete(&model.EmployeeStack{}).Error
}
