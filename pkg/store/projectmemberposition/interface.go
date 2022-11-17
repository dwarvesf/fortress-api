package projectmemberposition

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type IStore interface {
	Create(db *gorm.DB, pos *model.ProjectMemberPosition) error
	HardDeleteByProjectMemberID(db *gorm.DB, id string) error
}
