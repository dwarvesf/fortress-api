package employee

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
}

func New() IStore {
	return &store{}
}

// One get 1 employee by id
func (s *store) One(db *gorm.DB, id string) (*model.Employee, error) {
	var employee *model.Employee
	return employee, db.Where("id = ?", id).
		Preload("Roles").
		Preload("Chapter").
		Preload("Seniority").
		First(&employee).
		Error
}

// OneByTeamEmail get 1 employee by team email
func (s *store) OneByTeamEmail(db *gorm.DB, teamEmail string) (*model.Employee, error) {
	var employee *model.Employee
	return employee, db.Where("team_email = ?", teamEmail).First(&employee).Error
}

// Search get employees by filter and pagination
func (s *store) Search(db *gorm.DB, filter SearchFilter, pagination model.Pagination) ([]*model.Employee, int64, error) {
	var total int64
	var employees []*model.Employee

	query := db.Table("employees")

	if filter.WorkingStatus != "" {
		query = query.Where("working_status = ?", filter.WorkingStatus)
	}
	query = query.Count(&total)

	query = query.Order(pagination.Sort)
	limit, offset := pagination.ToLimitOffset()
	if pagination.Page > 0 {
		query = query.Limit(limit)
	}

	query = query.Preload("ProjectMembers", "deleted_at IS NULL").
		Preload("ProjectMembers.Project").
		Preload("ProjectMembers.Project.Heads").
		Preload("EmployeePositions", "deleted_at IS NULL").
		Preload("EmployeePositions.Position").
		Preload("Roles", "deleted_at IS NULL").
		Offset(offset)

	return employees, total, query.Find(&employees).Error
}

func (s *store) UpdateEmployeeStatus(db *gorm.DB, employeeID string, accountStatus model.AccountStatus) (*model.Employee, error) {
	employee := &model.Employee{}
	return employee, db.Model(&employee).Where("id = ?", employeeID).Update("account_status", string(accountStatus)).Find(&employee).Error
}

func (s *store) UpdateGeneralInfo(db *gorm.DB, body UpdateGeneralInfoInput, id string) (*model.Employee, error) {
	employee := &model.Employee{}

	employee.FullName = body.FullName
	employee.TeamEmail = body.Email
	employee.PhoneNumber = body.Phone
	employee.DiscordID = body.DiscordID
	employee.GithubID = body.GithubID
	employee.LineManagerID = body.LineManagerID

	return employee, db.Table("employees").Where("id = ?", id).Updates(&employee).
		Preload("LineManager").
		First(&employee).Error
}

func (s *store) Create(db *gorm.DB, e *model.Employee) (employee *model.Employee, err error) {
	return e, db.Create(e).Error
}

func (s *store) UpdatePersonalInfo(db *gorm.DB, body UpdatePersonalInfoInput, id string) (*model.Employee, error) {
	employee := &model.Employee{}

	employee.DateOfBirth = body.DoB
	employee.Gender = body.Gender
	employee.Address = body.Address
	employee.PersonalEmail = body.PersonalEmail

	return employee, db.Table("employees").Where("id = ?", id).Updates(&employee).First(&employee).Error
}

func (s *store) Update(db *gorm.DB, id string, employee *model.Employee) (*model.Employee, error) {
	return employee, db.Table("employees").Where("id = ?", id).Updates(&employee).
		Preload("Chapter").
		Preload("Seniority").
		Preload("LineManager").
		Preload("EmployeePositions", "deleted_at IS NULL").
		Preload("EmployeePositions.Position").
		Preload("EmployeeStacks", "deleted_at IS NULL").
		Preload("EmployeeStacks.Stack").
		First(&employee).Error
}

// Exists check the existence of employee
func (s *store) Exists(db *gorm.DB, id string) (bool, error) {
	type res struct {
		Result bool
	}

	result := res{}
	query := db.Raw("SELECT EXISTS (SELECT * FROM employees WHERE id = ?) as result", id)

	return result.Result, query.Scan(&result).Error
}
