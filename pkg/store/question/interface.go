package question

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	AllByCategory(db *gorm.DB, category model.EventType, subcategory model.EventSubtype) (questions []*model.Question, err error)
}
