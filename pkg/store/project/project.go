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

	if input.Status != "" {
		query = query.Where("status = ?", input.Status)
	}

	if input.Type != "" {
		query = query.Where("type = ?", input.Type)
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

	query = query.Preload("ProjectMembers", "deleted_at IS NULL and left_date IS NULL AND status = ?", model.ProjectMemberStatusActive).
		Preload("ProjectMembers.Employee").
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

// One get 1 project by id
func (s *store) One(db *gorm.DB, id string) (*model.Project, error) {
	var project *model.Project
	return project, db.Where("id = ?", id).
		Preload("ProjectMembers", "deleted_at IS NULL and left_date IS NULL AND status IN ?",
			[]model.ProjectMemberStatus{model.ProjectMemberStatusActive, model.ProjectMemberStatusOnBoarding}).
		Preload("ProjectMembers.Employee").
		Preload("ProjectMembers.ProjectMemberPositions", "deleted_at IS NULL").
		Preload("ProjectMembers.ProjectMemberPositions.Position").
		Preload("ProjectMembers.Seniority", "deleted_at IS NULL").
		Preload("Heads", "deleted_at IS NULL and left_date IS NULL").
		Preload("Heads.Employee").
		Preload("ProjectStacks", "deleted_at IS NULL").
		Preload("ProjectStacks.Stack", "deleted_at IS NULL").
		Preload("Country").
		First(&project).
		Error
}

// Update update all value (including nested model)
func (s *store) Update(db *gorm.DB, project *model.Project) (*model.Project, error) {
	return project, db.Model(&project).Where("id = ?", project.ID).Updates(&project).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.Project, updatedFields ...string) (*model.Project, error) {
	project := model.Project{}
	return &project, db.Model(&project).Where("id = ?", id).Select(updatedFields).Updates(&updateModel).Error
}
