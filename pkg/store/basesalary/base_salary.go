package basesalary

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) One(db *gorm.DB, id string) (*model.BaseSalary, error) {
	var baseSalary *model.BaseSalary
	return baseSalary, db.Where("id = ?", id).
		Preload("Currency").
		First(&baseSalary).Error
}

func (s *store) OneByEmployeeID(db *gorm.DB, employeeID string) (*model.BaseSalary, error) {
	var baseSalary *model.BaseSalary
	return baseSalary, db.Where("employee_id = ?", employeeID).
		Preload("Currency").
		First(&baseSalary).Error
}

func (s *store) Save(db *gorm.DB, baseSalary *model.BaseSalary) (err error) {
	return db.Save(&baseSalary).Preload("Currency").Error
}
