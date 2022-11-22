package projectstack

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, projectStack *model.ProjectStack) (*model.ProjectStack, error)
	HardDelete(db *gorm.DB, projectID string) (err error)
}
