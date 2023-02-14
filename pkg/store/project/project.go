package project

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// All get all projects by query and pagination
func (s *store) All(db *gorm.DB, input GetListProjectInput, pagination model.Pagination) ([]*model.Project, int64, error) {
	var projects []*model.Project

	query := db.Table("projects").
		Where("deleted_at IS NULL")

	var total int64

	if input.Name != "" {
		query = query.Where("name ILIKE ?", fmt.Sprintf("%%%s%%", input.Name))
	}

	if len(input.Statuses) > 0 {
		query = query.Where("status IN ?", input.Statuses)
	}

	if input.Type != "" {
		query = query.Where("type = ?", input.Type)
	}

	if input.AllowsSendingSurvey {
		query = query.Where("allows_sending_survey = ?", input.AllowsSendingSurvey)
	}

	if pagination.Sort != "" {
		query = query.Order(pagination.Sort)
	} else {
		query = query.Order("updated_at DESC")
	}

	query = query.Count(&total)

	limit, offset := pagination.ToLimitOffset()
	if pagination.Page > 0 {
		query = query.Limit(limit)
	}

	query = query.Preload("ProjectMembers", func(db *gorm.DB) *gorm.DB {
		return db.Joins("JOIN projects ON project_members.project_id = projects.id").
			Where("project_members.deleted_at IS NULL AND (projects.status = ? OR project_members.status = ?)",
				model.ProjectStatusClosed,
				model.ProjectMemberStatusActive)
	}).
		Preload("ProjectMembers.Employee").
		Preload("ProjectNotion", "deleted_at IS NULL").
		Preload("Organization", "deleted_at IS NULL").
		Preload("Heads", `deleted_at IS NULL AND (end_date IS NULL OR end_date > now())`).
		Preload("Heads.Employee").
		Offset(offset)

	return projects, total, query.Find(&projects).Error
}

// Create use to create new project to database
func (s *store) Create(db *gorm.DB, project *model.Project) error {
	return db.Create(&project).Preload("Country").Error
}

// IsExist check project existence
func (s *store) IsExist(db *gorm.DB, id string) (bool, error) {
	type res struct {
		Result bool
	}

	result := res{}
	query := db.Raw("SELECT EXISTS (SELECT * FROM projects WHERE id = ?) as result", id)

	return result.Result, query.Scan(&result).Error
}

// IsExistByCode check project existence by code
func (s *store) IsExistByCode(db *gorm.DB, code string) (bool, error) {
	type res struct {
		Result bool
	}

	result := res{}
	query := db.Raw("SELECT EXISTS (SELECT * FROM projects WHERE code = ?) as result", code)

	return result.Result, query.Scan(&result).Error
}

// One get 1 project by id
func (s *store) One(db *gorm.DB, id string, preload bool) (*model.Project, error) {
	query := db
	if !model.IsUUIDFromString(id) {
		query = db.Where("code = ?", id)
	} else {
		query = db.Where("id = ?", id)
	}

	if preload {
		query = query.
			Preload("Heads", "deleted_at IS NULL AND (end_date IS NULL OR end_date > now())").
			Preload("Heads.Employee", "deleted_at IS NULL").
			Preload("ProjectStacks", "deleted_at IS NULL").
			Preload("ProjectStacks.Stack", "deleted_at IS NULL").
			Preload("Country", "deleted_at IS NULL").
			Preload("BankAccount", "deleted_at IS NULL").
			Preload("Client", "deleted_at IS NULL").
			Preload("Client.Contacts", "deleted_at IS NULL").
			Preload("CompanyInfo", "deleted_at IS NULL").
			Preload("ProjectNotion", "deleted_at IS NULL").
			Preload("Organization", "deleted_at IS NULL").
			Preload("ProjectMembers", func(db *gorm.DB) *gorm.DB {
				return db.Joins("JOIN seniorities s ON s.id = project_members.seniority_id").
					Joins(`LEFT JOIN project_heads ph ON ph.project_id = project_members.project_id 
						AND ph.employee_id = project_members.employee_id 
						AND ph.position = ?
						AND (ph.end_date IS NULL OR ph.end_date > now())`,
						model.HeadPositionTechnicalLead,
					).
					Where("project_members.deleted_at IS NULL").
					Where("project_members.start_date <= now()").
					Where("(project_members.end_date IS NULL OR project_members.end_date > now())").
					Order("CASE ph.position WHEN 'technical-lead' THEN 1 ELSE 2 END").
					Order("s.level DESC")
			}).
			Preload("ProjectMembers.Employee", "deleted_at IS NULL").
			Preload("ProjectMembers.ProjectMemberPositions", "deleted_at IS NULL").
			Preload("ProjectMembers.ProjectMemberPositions.Position", "deleted_at IS NULL").
			Preload("ProjectMembers.Seniority", "deleted_at IS NULL").
			Preload("ProjectMembers.UpsellPerson", "deleted_at IS NULL")
	}

	var project *model.Project
	return project, query.First(&project).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.Project, updatedFields ...string) (*model.Project, error) {
	project := model.Project{}
	return &project, db.Model(&project).Where("id = ?", id).Select(updatedFields).Updates(&updateModel).Error
}

// GetByEmployeeID get project list by employee id
func (s *store) GetByEmployeeID(db *gorm.DB, employeeID string) ([]*model.Project, error) {
	var projects []*model.Project

	query := db.Table("projects").
		Joins("JOIN project_members pm ON pm.project_id = projects.id").
		Where("pm.start_date <= now() AND (pm.end_date IS NULL OR pm.end_date > now())").
		Where("projects.status = ?", model.ProjectStatusActive).
		Where("projects.deleted_at IS NULL AND pm.employee_id = ?", employeeID).
		Preload("Heads", func(db *gorm.DB) *gorm.DB {
			return db.Joins("JOIN projects p ON project_heads.project_id = p.id").
				Where("(project_heads.end_date IS NULL OR project_heads.end_date > ?) AND project_heads.employee_id = ? AND project_heads.position = ?", time.Now(), employeeID, model.HeadPositionTechnicalLead)
		}).
		Preload("Heads.Employee")

	return projects, query.Find(&projects).Error
}
