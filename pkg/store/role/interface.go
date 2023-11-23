package role

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB) (roles []*model.Role, err error)
	One(db *gorm.DB, id model.UUID) (role *model.Role, err error)
	GetByLevel(db *gorm.DB, level int64) ([]*model.Role, error)
	GetByCode(db *gorm.DB, code string) (*model.Role, error)
	GetByIDs(db *gorm.DB, ids []model.UUID) ([]*model.Role, error)
	IsExist(db *gorm.DB, id string) (exists bool, err error)
}
