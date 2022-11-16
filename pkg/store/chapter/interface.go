package chapter

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB) (chapters []*model.Chapter, err error)
	Exists(db *gorm.DB, id string) (bool, error)
}
