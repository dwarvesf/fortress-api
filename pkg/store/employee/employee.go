package employee

import (
	"fmt"

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
		Preload("ProjectMembers", "deleted_at IS NULL").
		Preload("ProjectMembers.Project", "deleted_at IS NULL").
		Preload("ProjectMembers.ProjectMemberPositions", "deleted_at IS NULL").
		Preload("ProjectMembers.ProjectMemberPositions.Position", "deleted_at IS NULL").
		Preload("EmployeePositions", "deleted_at IS NULL").
		Preload("EmployeePositions.Position", "deleted_at IS NULL").
		Preload("EmployeeStacks", "deleted_at IS NULL").
		Preload("EmployeeStacks.Stack", "deleted_at IS NULL").
		Preload("EmployeeRoles", "deleted_at IS NULL").
		Preload("EmployeeRoles.Role", "deleted_at IS NULL").
		Preload("Seniority").
		Preload("Chapter").
		Preload("LineManager").
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

	if filter.PositionID != "" {
		query = query.Joins("JOIN employee_positions ON employees.id = employee_positions.employee_id AND employee_positions.position_id = ?",
			filter.PositionID)
	}

	if filter.StackID != "" {
		query = query.Joins("JOIN employee_stacks ON employees.id = employee_stacks.employee_id AND employee_stacks.stack_id = ?",
			filter.StackID)
	}

	if filter.ProjectID != "" {
		query = query.Joins("JOIN project_members ON employees.id = project_members.employee_id AND project_members.project_id = ?",
			filter.ProjectID)
	}

	if filter.Keyword != "" {
		keywork := fmt.Sprintf("%%%s%%", filter.Keyword)

		query = query.Where("github_id like ?", keywork).
			Or("discord_id like ?", keywork).
			Or("notion_id like ?", keywork).
			Or("full_name like ?", keywork).
			Or("team_email like ?", keywork)
	}

	query = query.Count(&total)

	query = query.Order(pagination.Sort)
	limit, offset := pagination.ToLimitOffset()
	if pagination.Page > 0 {
		query = query.Limit(limit)
	}

	query = query.Offset(offset)

	if filter.Preload {
		query = query.Preload("ProjectMembers", "deleted_at IS NULL").
			Preload("ProjectMembers.Project", "deleted_at IS NULL").
			Preload("ProjectMembers.Project.Heads", "deleted_at IS NULL").
			Preload("EmployeePositions", "deleted_at IS NULL").
			Preload("EmployeePositions.Position", "deleted_at IS NULL").
			Preload("EmployeeRoles", "deleted_at IS NULL").
			Preload("EmployeeRoles.Role", "deleted_at IS NULL").
			Preload("EmployeeStacks", "deleted_at IS NULL").
			Preload("EmployeeStacks.Stack", "deleted_at IS NULL")
	}

	return employees, total, query.Find(&employees).Error
}

func (s *store) UpdateEmployeeStatus(db *gorm.DB, employeeID string, status model.WorkingStatus) (*model.Employee, error) {
	employee := &model.Employee{}
	return employee, db.Model(&employee).Where("id = ?", employeeID).Update("working_status", string(status)).Find(&employee).Error
}

func (s *store) UpdateGeneralInfo(db *gorm.DB, body UpdateGeneralInfoInput, id string) (*model.Employee, error) {
	employee := &model.Employee{}

	employee.FullName = body.FullName
	employee.TeamEmail = body.Email
	employee.PhoneNumber = body.Phone
	employee.DiscordID = body.DiscordID
	employee.GithubID = body.GithubID
	employee.NotionID = body.NotionID
	employee.LineManagerID = body.LineManagerID

	return employee, db.Table("employees").Where("id = ?", id).Updates(&employee).
		Preload("LineManager").
		First(&employee).Error
}

func (s *store) UpdateProfileInfo(db *gorm.DB, body UpdateProfileInforInput, id string) (*model.Employee, error) {
	employee := &model.Employee{}
	updateInfo := map[string]interface{}{}

	updateInfo["team_email"] = body.TeamEmail
	updateInfo["personal_email"] = body.PersonalEmail
	updateInfo["phone_number"] = body.PhoneNumber
	updateInfo["discord_id"] = body.DiscordID
	updateInfo["github_id"] = body.GithubID
	updateInfo["notion_id"] = body.NotionID

	return employee, db.Model(&employee).Where("id = ?", id).Updates(updateInfo).First(&employee).Error
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
