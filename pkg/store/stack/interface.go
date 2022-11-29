package stack

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB) ([]*model.Stack, error)
	One(db *gorm.DB, id string) (*model.Stack, error)
	GetByIDs(db *gorm.DB, ids []string) (stacks []*model.Stack, err error)
}
