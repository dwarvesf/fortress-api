package project

import (
	"fmt"

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

	query = query.Preload("Slots", "deleted_at IS NULL").
		Preload("Slots.ProjectMember", "deleted_at IS NULL and left_date IS NULL AND status = ?", model.ProjectMemberStatusActive).
		Preload("Slots.ProjectMember.Employee").
		Preload("Heads", "deleted_at IS NULL and left_date IS NULL").
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
			Preload("Heads", "deleted_at IS NULL and left_date IS NULL").
			Preload("Heads.Employee").
			Preload("ProjectStacks", "deleted_at IS NULL").
			Preload("ProjectStacks.Stack", "deleted_at IS NULL").
			Preload("Country").
			Preload("Slots", func(db *gorm.DB) *gorm.DB {
				return db.Joins("JOIN project_members pm ON pm.project_slot_id = project_slots.id").
					Joins("JOIN seniorities s ON s.id = pm.seniority_id").
					Joins("LEFT JOIN project_heads ph ON ph.project_id = pm.project_id AND ph.employee_id = pm.employee_id").
					Where("project_slots.deleted_at IS NULL").
					Where("pm.deleted_at IS NULL").
					Where("pm.status IN ?", []model.ProjectMemberStatus{model.ProjectMemberStatusActive, model.ProjectMemberStatusOnBoarding}).
					Where("project_slots.status = ?", model.ProjectMemberStatusActive).
					Order("pm.left_date ASC").
					Order("CASE ph.position WHEN 'technical-lead' THEN 1 ELSE 2 END").
					Order("s.level DESC")
			}).
			Preload("Slots.ProjectMember", "deleted_at IS NULL AND status IN ?",
				[]model.ProjectMemberStatus{model.ProjectMemberStatusActive, model.ProjectMemberStatusOnBoarding}).
			Preload("Slots.ProjectMember.Employee", "deleted_at IS NULL").
			Preload("Slots.ProjectMember.ProjectMemberPositions", "deleted_at IS NULL").
			Preload("Slots.ProjectMember.ProjectMemberPositions.Position", "deleted_at IS NULL").
			Preload("Slots.ProjectMember.Seniority", "deleted_at IS NULL").
			Preload("Slots.ProjectSlotPositions", "deleted_at IS NULL").
			Preload("Slots.ProjectSlotPositions.Position", "deleted_at IS NULL").
			Preload("Slots.ProjectSlotPositions.Position", "deleted_at IS NULL").
			Preload("Slots.Seniority", "deleted_at IS NULL")
	}

	var project *model.Project
	return project, query.First(&project).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.Project, updatedFields ...string) (*model.Project, error) {
	project := model.Project{}
	return &project, db.Model(&project).Where("id = ?", id).Select(updatedFields).Updates(&updateModel).Error
}
