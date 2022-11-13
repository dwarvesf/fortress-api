package projecthead

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, projectHead *model.ProjectHead) error
	GetByProjectID(db *gorm.DB, projectID string) (projectHeads []*model.ProjectHead, err error)
}
