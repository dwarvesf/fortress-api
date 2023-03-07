package bonus

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

// New initilize new store for bonus
func New() IStore {
	return &store{}
}

func (s *store) GetByUserID(db *gorm.DB, id model.UUID) ([]model.EmployeeBonus, error) {
	var res []model.EmployeeBonus
	return res, db.Where("is_active = true AND employee_id = ?", id).Find(&res).Error
}
