package workunitstack

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Create create new WorkUnitStack
func (s *store) Create(db *gorm.DB, wus *model.WorkUnitStack) error {
	return db.Create(&wus).Error
}

// DeleteByWorkUnitID delete many workUnitStack by workUnitID
func (s *store) DeleteByWorkUnitID(db *gorm.DB, workUnitID string) error {
	return db.Unscoped().Where("work_unit_id = ?", workUnitID).Delete(&model.WorkUnitStack{}).Error
}
