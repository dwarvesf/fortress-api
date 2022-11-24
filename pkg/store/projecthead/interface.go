package projecthead

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, head *model.ProjectHead) error
	Upsert(db *gorm.DB, head *model.ProjectHead) error
	GetByProjectID(db *gorm.DB, projectID string) (projectHeads []*model.ProjectHead, err error)
	DeleteByProjectIDAndPosition(db *gorm.DB, projectID string, pos string) (err error)
	DeleteByPositionInProject(db *gorm.DB, projectID string, employeeID string, position string) (err error)
	One(db *gorm.DB, projectID string, position model.HeadPosition) (projectHead *model.ProjectHead, err error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.ProjectHead, updatedFields ...string) (*model.ProjectHead, error)
}
