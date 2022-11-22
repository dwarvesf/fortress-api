package projecthead

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, head *model.ProjectHead) error
	Upsert(db *gorm.DB, head *model.ProjectHead) error
	GetByProjectID(db *gorm.DB, projectID string) (projectHeads []*model.ProjectHead, err error)
	DeleteByProjectIDAndPosition(db *gorm.DB, projectID string, pos string) (err error)
	HardDeleteByPosition(db *gorm.DB, projectID string, employeeID string, position string) (err error)
	One(db *gorm.DB, projectID string, position model.HeadPosition) (projectHead *model.ProjectHead, err error)
	UpdateLeftDate(db *gorm.DB, projectID string, employeeID string, position string, timeNow time.Time) (err error)
}
