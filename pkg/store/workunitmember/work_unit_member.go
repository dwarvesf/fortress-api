package workunitmember

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Create create new WorkUnitMember
func (s *store) Create(db *gorm.DB, wum *model.WorkUnitMember) error {
	return db.Create(&wum).Error
}

// GetByWorkUnitID return list member of a work unit
func (s *store) GetByWorkUnitID(db *gorm.DB, wuID string) (wuMembers []model.WorkUnitMember, err error) {
	var members []model.WorkUnitMember
	return members, db.Where("work_unit_id = ?", wuID).Find(&members).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.WorkUnitMember, updatedFields ...string) (*model.WorkUnitMember, error) {
	member := model.WorkUnitMember{}
	return &member, db.Model(&member).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}
