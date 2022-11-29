package workunit

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// GetManyByProjectID get all work units of a project and having the status as required
func (s *store) GetAllByProjectID(db *gorm.DB, projectID string, status model.WorkUnitStatus) ([]*model.WorkUnit, error) {
	var workUnits []*model.WorkUnit

	return workUnits, db.Where("project_id = ? and status = ?", projectID, status).
		Preload("WorkUnitMembers", "deleted_at IS NULL and status = 'active'").
		Preload("WorkUnitMembers.Employee", "deleted_at IS NULL").
		Preload("WorkUnitStacks", "deleted_at IS NULL").
		Preload("WorkUnitStacks.Stack", "deleted_at IS NULL").
		Find(&workUnits).Error
}

// Create create new WorkUnit
func (s *store) Create(db *gorm.DB, workUnit *model.WorkUnit) error {
	return db.Create(&workUnit).Error
}
