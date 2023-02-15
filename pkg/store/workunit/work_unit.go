package workunit

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// GetByProjectID get all work units of a project and having the status as required
func (s *store) GetByProjectID(db *gorm.DB, projectID string, status model.WorkUnitStatus) ([]*model.WorkUnit, error) {
	var workUnits []*model.WorkUnit
	query := db.Where("project_id = ?", projectID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if status == model.WorkUnitStatusActive {
		query = query.Preload("WorkUnitMembers", "deleted_at IS NULL and status = 'active'")
	} else {
		query = query.Preload("WorkUnitMembers", "deleted_at IS NULL")
	}

	return workUnits, query.Preload("WorkUnitMembers.Employee", "deleted_at IS NULL").
		Preload("WorkUnitStacks", "deleted_at IS NULL").
		Preload("WorkUnitStacks.Stack", "deleted_at IS NULL").
		Find(&workUnits).Error
}

// Create create new WorkUnit
func (s *store) Create(db *gorm.DB, workUnit *model.WorkUnit) error {
	return db.Create(&workUnit).Error
}

// One get 1 WorkUnit by ID
func (s *store) One(db *gorm.DB, id string) (*model.WorkUnit, error) {
	var workUnit *model.WorkUnit
	return workUnit, db.Where("id = ?", id).First(&workUnit).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.WorkUnit, updatedFields ...string) (*model.WorkUnit, error) {
	workUnit := model.WorkUnit{}
	return &workUnit, db.Model(&workUnit).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}

// IsExists check work unit existence
func (s *store) IsExists(db *gorm.DB, id string) (bool, error) {
	type res struct {
		Result bool
	}

	result := res{}
	query := db.Raw("SELECT EXISTS (SELECT * FROM work_units WHERE id = ?) as result", id)

	return result.Result, query.Scan(&result).Error
}

func (s *store) GetAllWorkUnitByEmployeeID(db *gorm.DB, employeeID string) ([]*model.WorkUnit, error) {
	var workUnits []*model.WorkUnit
	return workUnits, db.Where("id IN (SELECT work_unit_id FROM work_unit_members JOIN projects ON work_unit_members.project_id = projects.id WHERE employee_id = ? AND projects.status <> ?)", employeeID, model.ProjectStatusClosed).
		Preload("WorkUnitMembers", "deleted_at IS NULL").
		Preload("WorkUnitMembers.Employee", "deleted_at IS NULL").
		Preload("Project", "deleted_at IS NULL").
		Find(&workUnits).Error
}
