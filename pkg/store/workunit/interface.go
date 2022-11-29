package workunit

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, workUnit *model.WorkUnit) error
	GetAllByProjectID(db *gorm.DB, projectID string, status model.WorkUnitStatus) (workUnits []*model.WorkUnit, err error)
}
