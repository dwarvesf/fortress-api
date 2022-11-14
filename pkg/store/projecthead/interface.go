package projecthead

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, head *model.ProjectHead) error
	Upsert(db *gorm.DB, head *model.ProjectHead) error
	GetByProjectID(db *gorm.DB, projectID string) (projectHeads []*model.ProjectHead, err error)
	DeleteByProjectIDAndPosition(db *gorm.DB, projectID string, pos string) error
}
