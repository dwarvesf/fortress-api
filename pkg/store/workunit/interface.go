package workunit

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, workUnit *model.WorkUnit) error
	GetByProjectID(db *gorm.DB, projectID string, status model.WorkUnitStatus) (workUnits []*model.WorkUnit, err error)
	One(db *gorm.DB, id string) (*model.WorkUnit, error)
	IsExists(db *gorm.DB, id string) (bool, error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.WorkUnit, updatedFields ...string) (workUnit *model.WorkUnit, err error)
	GetAllWorkUnitByEmployeeID(db *gorm.DB, employeeID string) (workUnits []*model.WorkUnit, err error)
}
