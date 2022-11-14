package projectmember

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type IStore interface {
	GetByProjectIDAndEmployeeID(db *gorm.DB, projectID string, employeeID string) (*model.ProjectMember, error)
	Create(db *gorm.DB, member *model.ProjectMember) error
	Upsert(db *gorm.DB, member *model.ProjectMember) error
}
