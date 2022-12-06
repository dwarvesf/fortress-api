package chapter

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB) (chapters []*model.Chapter, err error)
	IsExist(db *gorm.DB, id string) (isExist bool, err error)
	UpdateChapterLead(db *gorm.DB, id string, lead *model.UUID) (err error)
	GetAllByLeadID(db *gorm.DB, leadID string) (chapters []*model.Chapter, err error)
}
