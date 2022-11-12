package project

import (
	"fmt"

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

// All get all projects with filter and pagination
func (s *store) All(input GetListProjectInput, pagination model.Pagination) ([]*model.Project, int64, error) {
	var projects []*model.Project

	query := s.db.Debug().Table("projects").
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

type GetListProjectInput struct {
	Status string `json:"status"`
	Name   string `json:"name"`
	Type   string `json:"type"`
}
