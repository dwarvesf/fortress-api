package employee

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// One get 1 employee by id
func (s *store) One(db *gorm.DB, id string, preload bool) (*model.Employee, error) {
	query := db.Where("id = ?", id).
		Preload("EmployeeRoles", func(db *gorm.DB) *gorm.DB {
			return db.Joins("employee_roles JOIN roles ON roles.id = employee_roles.role_id").
				Where("employee_roles.deleted_at IS NULL").
				Order("roles.level")
		}).
		Preload("EmployeeRoles.Role", "deleted_at IS NULL")

	if preload {
		query = query.
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
			Preload("Seniority").
			Preload("LineManager").
			Preload("Mentees", "deleted_at IS NULL").
			Preload("Mentees.Mentee", "deleted_at IS NULL AND NOT working_status = 'left'").
			Preload("Mentees.Mentee.EmployeePositions", "deleted_at IS NULL").
			Preload("Mentees.Mentee.EmployeePositions.Position", "deleted_at IS NULL").
			Preload("Mentees.Mentee.Seniority")
	}

	var employee *model.Employee
	return employee, query.First(&employee).Error
}

// OneByUsername get 1 employee by username
func (s *store) OneByUsername(db *gorm.DB, username string, preload bool) (*model.Employee, error) {
	query := db.Where("username = ?", username)

	if preload {
		query = query.
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
			Preload("Mentees", "deleted_at IS NULL").
			Preload("Mentees.Mentee", "deleted_at IS NULL AND NOT working_status = 'left'").
			Preload("Mentees.Mentee.EmployeePositions", "deleted_at IS NULL").
			Preload("Mentees.Mentee.EmployeePositions.Position", "deleted_at IS NULL").
			Preload("Mentees.Mentee.Seniority")
	}

	var employee *model.Employee
	return employee, query.First(&employee).Error
}

// OneByTeamEmail get 1 employee by team email
func (s *store) OneByTeamEmail(db *gorm.DB, teamEmail string) (*model.Employee, error) {
	var employee *model.Employee
	return employee, db.Where("team_email = ?", teamEmail).First(&employee).Error
}

// All get employees by query and pagination
func (s *store) All(db *gorm.DB, filter EmployeeFilter, pagination model.Pagination) ([]*model.Employee, int64, error) {
	var total int64
	var employees []*model.Employee

	query := db.Table("employees")

	query = getByWhereConditions(query, filter)
	query = getByFieldSort(query, "employees.joined_date", filter.JoinedDateSort)

	query = query.Count(&total)

	if filter.Preload {
		query = query.Preload("ProjectMembers", "deleted_at IS NULL AND status != 'inactive'").
			Preload("LineManager", "deleted_at IS NULL").
			Preload("Seniority", "deleted_at IS NULL").
			Preload("ProjectMembers.Project", "deleted_at IS NULL").
			Preload("ProjectMembers.Project.Heads", "deleted_at IS NULL").
			Preload("EmployeePositions", "deleted_at IS NULL").
			Preload("EmployeePositions.Position", "deleted_at IS NULL").
			Preload("EmployeeRoles", "deleted_at IS NULL").
			Preload("EmployeeRoles.Role", "deleted_at IS NULL").
			Preload("EmployeeChapters", "deleted_at IS NULL").
			Preload("EmployeeChapters.Chapter", "deleted_at IS NULL").
			Preload("EmployeeStacks", "deleted_at IS NULL").
			Preload("EmployeeStacks.Stack", "deleted_at IS NULL")
	}

	limit, offset := pagination.ToLimitOffset()
	if pagination.Page > 0 {
		query = query.Limit(limit)
	}

	return employees, total, query.
		Limit(limit).
		Offset(offset).
		Find(&employees).Error
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
	return employee, db.Model(&employee).Where("id = ?", employee.ID).Updates(&employee).First(&employee).Error
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

// GetByWorkingStatus return list employee by working status
func (s *store) GetByWorkingStatus(db *gorm.DB, workingStatus model.WorkingStatus, preload bool) ([]*model.Employee, error) {
	var employees []*model.Employee
	return employees, db.Where("working_status = ?", workingStatus).Preload("WorkUnitMembers", "deleted_at IS NULL AND status = ?", model.ProjectStatusActive).Preload("WorkUnitMembers.WorkUnit", "deleted_at IS NULL").Find(&employees).Error
}
