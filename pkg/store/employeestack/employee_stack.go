package employeestack

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Create create new one by id
func (s *store) Create(db *gorm.DB, employeeStack *model.EmployeeStack) (*model.EmployeeStack, error) {
	return employeeStack, db.Create(&employeeStack).Error
}

// DeleteByEmployeeID delete many EmployeeStaks by employeeID
func (s *store) DeleteByEmployeeID(db *gorm.DB, employeeID string) error {
	return db.Unscoped().Where("employee_id = ?", employeeID).Delete(&model.EmployeeStack{}).Error
}
