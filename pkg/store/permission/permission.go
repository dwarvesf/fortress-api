package permission

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// GetByEmployeeID get list of permissions by employee id
func (s *store) GetByEmployeeID(db *gorm.DB, employeeID string) ([]*model.Permission, error) {
	var permissions []*model.Permission
	return permissions, db.
		Select("DISTINCT permissions.*").
		Joins("JOIN role_permissions rp ON permissions.id = rp.permission_id").
		Joins("JOIN employee_roles er ON er.role_id = rp.role_id").
		Where("er.employee_id = ?", employeeID).
		Order("permissions.code").
		Find(&permissions).Error
}

func (s *store) HasPermission(db *gorm.DB, employeeID string, permCode string) (bool, error) {
	var res struct {
		Result bool
	}

	query := db.Raw(`
	SELECT EXISTS (
		SELECT * 
		FROM permissions p
			JOIN role_permissions rp ON p.id = rp.permission_id
			JOIN employee_roles er ON rp.role_id = er.role_id
			JOIN employees e ON er.employee_id = e.id AND e.id = ? 
		WHERE p.code = ?
	) as result`, employeeID, permCode)

	return res.Result, query.Scan(&res).Error
}
