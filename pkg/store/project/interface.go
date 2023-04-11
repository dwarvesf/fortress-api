package project

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB, input GetListProjectInput, pagination model.Pagination) ([]*model.Project, int64, error)
	Create(db *gorm.DB, project *model.Project) error
	IsExist(db *gorm.DB, id string) (bool, error)
	IsExistByCode(db *gorm.DB, code string) (bool, error)
	One(db *gorm.DB, id string, preload bool) (*model.Project, error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.Project, updatedFields ...string) (*model.Project, error)
	GetByEmployeeID(db *gorm.DB, employeeID string) ([]*model.Project, error)
	GetProjectByAlias(db *gorm.DB, alias string) (*model.Project, error)
}
