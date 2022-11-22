package project

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB, input GetListProjectInput, pagination model.Pagination) ([]*model.Project, int64, error)
	UpdateStatus(db *gorm.DB, projectID string, projectStatus model.ProjectStatus) (*model.Project, error)
	Create(db *gorm.DB, project *model.Project) error
	Exists(db *gorm.DB, id string) (bool, error)
	One(db *gorm.DB, id string) (*model.Project, error)
	UpdateGeneralInfo(db *gorm.DB, body UpdateGeneralInfoInput, id string) (*model.Project, error)
}
