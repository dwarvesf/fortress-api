package employeeinvitation

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Create generate new one
func (s *store) Create(db *gorm.DB, employeeInvitation *model.EmployeeInvitation) (*model.EmployeeInvitation, error) {
	return employeeInvitation, db.Create(&employeeInvitation).Error
}

// OneByEmployeeID get one by employeeID
func (s *store) OneByEmployeeID(db *gorm.DB, employeeID string) (*model.EmployeeInvitation, error) {
	var rs *model.EmployeeInvitation
	return rs, db.Where("employee_id = ?", employeeID).First(&rs).Error
}

func (s *store) Save(db *gorm.DB, employeeInvitation *model.EmployeeInvitation) error {
	return db.Save(&employeeInvitation).Error
}
