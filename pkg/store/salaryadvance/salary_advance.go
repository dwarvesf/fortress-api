package salaryadvance

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) One(db *gorm.DB, id string) (*model.SalaryAdvance, error) {
	var salaryAdvance *model.SalaryAdvance
	return salaryAdvance, db.Where("id = ?", id).
		First(&salaryAdvance).Error
}

func (s *store) ListNotPayBackByEmployeeID(db *gorm.DB, employeeID string) ([]model.SalaryAdvance, error) {
	var advanceSalaries []model.SalaryAdvance
	return advanceSalaries, db.Where("employee_id = ?", employeeID).Where("is_paid_back = ?", false).Find(&advanceSalaries).Error
}

func (s *store) Save(db *gorm.DB, salaryAdvance *model.SalaryAdvance) (err error) {
	return db.Save(&salaryAdvance).Error
}
