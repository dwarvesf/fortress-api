package projectnotion

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	OneByProjectID(db *gorm.DB, projectID string) (projectNotion *model.ProjectNotion, err error)
	OneByAuditNotionID(db *gorm.DB, projectID string) (projectNotion *model.ProjectNotion, err error)
	Create(db *gorm.DB, e *model.ProjectNotion) (projectNotion *model.ProjectNotion, err error)
	Update(db *gorm.DB, projectNotion *model.ProjectNotion) (a *model.ProjectNotion, err error)
	IsExistByAuditNotionID(db *gorm.DB, id string) (exists bool, err error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.ProjectNotion, updatedFields ...string) (projectNotion *model.ProjectNotion, err error)
}
