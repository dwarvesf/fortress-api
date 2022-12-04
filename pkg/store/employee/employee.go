package employee

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

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
		Preload("EmployeeChapters", "deleted_at IS NULL").
		Preload("EmployeeChapters.Chapter", "deleted_at IS NULL").
		Preload("EmployeeRoles", "deleted_at IS NULL").
		Preload("EmployeeRoles.Role", "deleted_at IS NULL").
		Preload("Seniority").
		Preload("LineManager").
		First(&employee).
		Error
}

// OneByTeamEmail get 1 employee by team email
func (s *store) OneByTeamEmail(db *gorm.DB, teamEmail string) (*model.Employee, error) {
	var employee *model.Employee
	return employee, db.Where("team_email = ?", teamEmail).First(&employee).Error
}

// All get employees by query and pagination
func (s *store) All(db *gorm.DB, input GetAllInput, pagination model.Pagination) ([]*model.Employee, int64, error) {
	var total int64
	var employees []*model.Employee

	query := db.Table("employees")

	if len(input.WorkingStatuses) > 0 {
		query = query.Where("working_status IN ?", input.WorkingStatuses)
	}

	if input.PositionID != "" {
		query = query.Joins("JOIN employee_positions ON employees.id = employee_positions.employee_id AND employee_positions.position_id = ?",
			input.PositionID)
	}

	if input.StackID != "" {
		query = query.Joins("JOIN employee_stacks ON employees.id = employee_stacks.employee_id AND employee_stacks.stack_id = ?",
			input.StackID)
	}

	if input.ProjectID != "" {
		query = query.Joins("JOIN project_members ON employees.id = project_members.employee_id AND project_members.project_id = ?",
			input.ProjectID)
	}

	if input.Keyword != "" {
		keywork := fmt.Sprintf("%%%s%%", input.Keyword)

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

	if input.Preload {
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

func (s *store) Create(db *gorm.DB, e *model.Employee) (employee *model.Employee, err error) {
	return e, db.Create(e).Error
}

// IsExist check the existence of employee
func (s *store) IsExist(db *gorm.DB, id string) (bool, error) {
	type res struct {
		Result bool
	}

	result := res{}
	query := db.Raw("SELECT EXISTS (SELECT * FROM employees WHERE id = ?) as result", id)

	return result.Result, query.Scan(&result).Error
}

// Update update all value (including nested model)
func (s *store) Update(db *gorm.DB, employee *model.Employee) (*model.Employee, error) {
	return employee, db.Model(&employee).Where("id = ?", employee.ID).Updates(&employee).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.Employee, updatedFields ...string) (*model.Employee, error) {
	employee := model.Employee{}
	return &employee, db.Model(&employee).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}

// GetByIDs return list employee by IDs
func (s *store) GetByIDs(db *gorm.DB, ids []model.UUID) ([]*model.Employee, error) {
	var employees []*model.Employee
	return employees, db.Where("id IN ?", ids).Find(&employees).Error
}
