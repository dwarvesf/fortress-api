package projectmember

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type IStore interface {
	One(db *gorm.DB, projectID string, employeeID string, preload bool) (*model.ProjectMember, error)
	OneBySlotID(db *gorm.DB, slotID string) (*model.ProjectMember, error)
	Create(db *gorm.DB, member *model.ProjectMember) error
	Delete(db *gorm.DB, id string) (err error)
	IsExist(db *gorm.DB, id string) (bool, error)
	IsExistsByEmployeeID(db *gorm.DB, projectID string, employeeID string) (bool, error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.ProjectMember, updatedFields ...string) (*model.ProjectMember, error)
	UpdateSelectedFieldByProjectID(db *gorm.DB, projectID string, updateModel model.ProjectMember, updatedField string) error
	UpdateLeftDateByProjectID(db *gorm.DB, projectID string) error
	GetByProjectIDs(db *gorm.DB, projectIDs []string) ([]*model.ProjectMember, error)
}
