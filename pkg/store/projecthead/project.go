package projecthead

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
}

func New() IStore {
	return &store{}
}

// Create using for insert new data to project head
func (s *store) Create(db *gorm.DB, projectHead *model.ProjectHead) error {
	return db.Create(projectHead).Preload("Employee").First(projectHead).Error
}

// GetByProjectID get all project heads by projectID
func (s *store) GetByProjectID(db *gorm.DB, projectID string) ([]*model.ProjectHead, error) {
	var projectHeads []*model.ProjectHead
	return projectHeads, db.Where("project_id = ? AND deleted_at IS NULL", projectID).Find(&projectHeads).Error
}
