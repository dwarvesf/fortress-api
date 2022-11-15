package permission

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
}

func New() IStore {
	return &store{}
}

// GetByEmployeeID get list of permissions by employee id
func (s *store) GetByEmployeeID(db *gorm.DB, employeeID string) ([]*model.Permission, error) {
	var permissions []*model.Permission
	return permissions, db.
		Joins("JOIN role_permissions rp ON permissions.id = rp.permission_id").
		Joins("JOIN employee_roles er ON er.role_id = rp.role_id").
		Where("er.employee_id = ?", employeeID).Find(&permissions).Error
}
