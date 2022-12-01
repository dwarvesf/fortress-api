package employeebonus

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) GetByEmployeeID(db *gorm.DB, employeeID model.UUID) ([]model.EmployeeBonus, error) {
	var res []model.EmployeeBonus
	return res, db.Where("is_active = true AND employee_id = ?", employeeID).Find(&res).Error
}
