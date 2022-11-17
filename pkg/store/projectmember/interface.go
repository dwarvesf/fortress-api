package projectmember

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type IStore interface {
	One(db *gorm.DB, projectID string, employeeID string) (*model.ProjectMember, error)
	Create(db *gorm.DB, member *model.ProjectMember) error
	Upsert(db *gorm.DB, member *model.ProjectMember) error
	HardDelete(db *gorm.DB, id string) (err error)
	Exists(db *gorm.DB, id string) (bool, error)
}
