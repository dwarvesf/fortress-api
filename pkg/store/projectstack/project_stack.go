package projectstack

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Create create new one by id
func (s *store) Create(db *gorm.DB, projectStack *model.ProjectStack) (*model.ProjectStack, error) {
	return projectStack, db.Create(&projectStack).Error
}

// DeleteByProjectID delete many ProjectStacks by projectID
func (s *store) DeleteByProjectID(db *gorm.DB, projectID string) error {
	return db.Unscoped().Where("project_id = ?", projectID).Delete(&model.ProjectStack{}).Error
}
