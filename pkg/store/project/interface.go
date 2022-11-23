package project

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB, input GetListProjectInput, pagination model.Pagination) ([]*model.Project, int64, error)
	Create(db *gorm.DB, project *model.Project) error
	IsExist(db *gorm.DB, id string) (bool, error)
	One(db *gorm.DB, id string) (*model.Project, error)

	// TODO: dùng chung 1 interface
	UpdateStatus(db *gorm.DB, projectID string, projectStatus model.ProjectStatus) (*model.Project, error)
	UpdateGeneralInfo(db *gorm.DB, body UpdateGeneralInfoInput, id string) (*model.Project, error)
	UpdateContactInfo(db *gorm.DB, body UpdateContactInfoInput, id string) (*model.Project, error)
}
