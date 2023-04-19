package projectcommissionconfig

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) GetByProjectID(db *gorm.DB, projectID string) (model.ProjectCommissionConfigs, error) {
	var heads model.ProjectCommissionConfigs
	return heads, db.Where("project_id = ?", projectID).Find(&heads).Error
}
