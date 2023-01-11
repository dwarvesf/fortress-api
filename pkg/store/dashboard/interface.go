package dashboard

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	GetProjectSizes(db *gorm.DB) (res []*model.ProjectSize, err error)
}
