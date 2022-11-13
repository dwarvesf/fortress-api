package project

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

// All get all projects with filter and pagination
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

	query = query.Preload("Members", "deleted_at IS NULL and left_date IS NULL AND status = ?", model.ProjectMemberStatusActive).
		Preload("Members.Employee").
		Preload("Heads", "deleted_at IS NULL and left_date IS NULL").
		Preload("Heads.Employee").
		Offset(offset)

	return projects, total, query.Find(&projects).Error
}

// UpdateStatus use to update project to database
func (s *store) UpdateStatus(db *gorm.DB, projectID string, projectStatus model.ProjectStatus) (*model.Project, error) {
	project := &model.Project{}
	return project, db.Model(&project).Where("id = ?", projectID).Update("status", string(projectStatus)).First(&project).Error
}

// Create use to create new project to database
func (s *store) Create(db *gorm.DB, project *model.Project) error {
	return db.Create(&project).Error
}

// Exists return true/false if project exist or not
func (s *store) Exists(db *gorm.DB, id string) (bool, error) {
	var record struct {
		Result bool
	}

	query := db.Raw("SELECT EXISTS (SELECT * FROM projects WHERE id = ?) as result", id)
	return record.Result, query.Scan(&record).Error
}
