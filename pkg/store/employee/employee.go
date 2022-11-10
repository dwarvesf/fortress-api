package employee_store

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
	db *gorm.DB
}

func New(db *gorm.DB) IStore {
	return &store{
		db: db,
	}
}

// One get 1 employee by id
func (s *store) One(id string) (*model.Employee, error) {
	var employee *model.Employee
	return employee, s.db.Where("id = ?", id).First(&employee).Error
}

// OneByTeamEmail get 1 employee by team email
func (s *store) OneByTeamEmail(teamEmail string) (*model.Employee, error) {
	var employee *model.Employee
	return employee, s.db.Where("team_email = ?", teamEmail).First(&employee).Error
}

// Search get employees by filter and pagination
func (s *store) Search(filter SearchFilter, pagination model.Pagination) ([]*model.Employee, int64, error) {
	db := s.db.Table("employees")
	var total int64

	if filter.WorkingStatus != "" {
		db = db.Where("working_status = ?", filter.WorkingStatus)
	}
	db = db.Count(&total)

	if pagination.Page > 1 {
		db = db.Offset(int((pagination.Page - 1) * pagination.Size))
	}
	db = db.Limit(int(pagination.Size))
	db = db.Order(pagination.Sort)

	var employees []*model.Employee
	return employees, total, db.Find(&employees).Error
}

func (s *store) UpdateEmployeeStatus(employeeID string, accountStatus model.AccountStatus) (*model.Employee, error) {
	employee := &model.Employee{}
	return employee, s.db.Model(&employee).Where("id = ?", employeeID).Update("account_status", string(accountStatus)).Find(&employee).Error
}
