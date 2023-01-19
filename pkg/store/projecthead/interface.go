package projecthead

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, head *model.ProjectHead) error
	GetActiveLeadsByProjectID(db *gorm.DB, projectID string) (projectHeads []*model.ProjectHead, err error)
	DeleteByPositionInProject(db *gorm.DB, projectID string, employeeID string, position string) (err error)
	One(db *gorm.DB, projectID string, position model.HeadPosition) (projectHead *model.ProjectHead, err error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.ProjectHead, updatedFields ...string) (*model.ProjectHead, error)
	UpdateEndDateOfEmployee(db *gorm.DB, employeeID string, projectID string, position string, endDate time.Time) (*model.ProjectHead, error)
}
