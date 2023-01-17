package stack

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB, keyword string, pagination *model.Pagination) (int64, []*model.Stack, error)
	One(db *gorm.DB, id string) (*model.Stack, error)
	GetByIDs(db *gorm.DB, ids []model.UUID) (stacks []*model.Stack, err error)
	Update(db *gorm.DB, stack *model.Stack) (s *model.Stack, err error)
	Create(db *gorm.DB, stack *model.Stack) (s *model.Stack, err error)
	Delete(db *gorm.DB, id string) (err error)
}
