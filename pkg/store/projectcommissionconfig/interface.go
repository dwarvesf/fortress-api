package projectcommissionconfig

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type IStore interface {
	GetByProjectID(db *gorm.DB, projectID string) (heads model.ProjectCommissionConfigs, err error)
}
