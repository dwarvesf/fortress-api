package employeechapter

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Create make new one by id
func (s *store) Create(db *gorm.DB, employeeChapter *model.EmployeeChapter) (*model.EmployeeChapter, error) {
	return employeeChapter, db.Create(&employeeChapter).Error
}

// DeleteByEmployeeID delete many EmployeeChapters by employeeID
func (s *store) DeleteByEmployeeID(db *gorm.DB, employeeID string) error {
	return db.Unscoped().Where("employee_id = ?", employeeID).Delete(&model.EmployeeChapter{}).Error
}
