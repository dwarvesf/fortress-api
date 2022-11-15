package employee

import (
	"time"

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

	// 1.2 update info
	employee.FullName = body.FullName
	employee.TeamEmail = body.Email
	employee.PhoneNumber = body.Phone
	employee.DiscordID = body.DiscordID
	employee.GithubID = body.GithubID

	if body.LineManagerID != "" {
		employee.LineManagerID = model.MustGetUUIDFromString(body.LineManagerID)
	}

	// 1.3 save to DB
	return employee, db.Table("employees").Where("id = ?", id).Updates(&employee).
		Preload("LineManager").
		First(&employee).Error
}

func (s *store) Create(db *gorm.DB, e *model.Employee) (employee *model.Employee, err error) {
	return e, db.Create(e).Error
}

func (s *store) UpdateSkills(db *gorm.DB, body UpdateSkillsInput, id string) (*model.Employee, error) {

	var employee *model.Employee

	// get employee by employee id
	err := db.Where("id = ?", id).First(&employee).Error
	if err != nil {
		return nil, err
	}

	// 1.1 delete all roles of the employee
	now := time.Now()
	employeePosition := model.EmployeePosition{}
	db.Table("employee_positions").Where("employee_id = ?", id).Update("deleted_at", now)

	// 1.2 create role for employee
	for _, positionID := range body.Positions {
		employeePosition.ID = model.NewUUID()
		employeePosition.EmployeeID = model.MustGetUUIDFromString(id)
		employeePosition.PositionID = positionID

		err = db.Table("employee_positions").Create(&employeePosition).Error
		if err != nil {
			return nil, err
		}
	}

	// 2 update tech stacks for employee

	// 2.1 delete all employee_stacks for employee id
	employeeStack := model.EmployeeStack{}
	err = db.Table("employee_stacks").Where("employee_id = ?", id).Update("deleted_at", now).Error
	if err != nil {
		return nil, err
	}

	// 2.2 create stacks for employee
	for _, stackID := range body.Stacks {
		employeeStack.ID = model.NewUUID()
		employeeStack.EmployeeID = model.MustGetUUIDFromString(id)
		employeeStack.StackID = stackID

		err = db.Table("employee_stacks").Create(&employeeStack).Error
		if err != nil {
			return nil, err
		}
	}

	// 3 update employee table

	// 3.2 update infor
	employee.ChapterID = body.Chapter
	employee.SeniorityID = body.Seniority

	// 3.3 save to DB
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

func (s *store) UpdatePersonalInfo(db *gorm.DB, body UpdatePersonalInfoInput, id string) (*model.Employee, error) {
	employee := &model.Employee{}

	// 1.2 update infor
	employee.DateOfBirth = body.DoB
	employee.Gender = body.Gender
	employee.Address = body.Address
	employee.PersonalEmail = body.PersonalEmail

	// 1.3 save to DB
	return employee, db.Table("employees").Where("id = ?", id).Updates(&employee).First(&employee).Error
}
