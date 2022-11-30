package workunitmember

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, wum *model.WorkUnitMember) error
	GetByWorkUnitID(db *gorm.DB, wuID string) (wuMembers []model.WorkUnitMember, err error)

	UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.WorkUnitMember, updatedFields ...string) (*model.WorkUnitMember, error)
}
