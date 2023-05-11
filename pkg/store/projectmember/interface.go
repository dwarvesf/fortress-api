package projectmember

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type IStore interface {
	Create(db *gorm.DB, member *model.ProjectMember) error
	Delete(db *gorm.DB, id string) (err error)
	GetActiveByProjectIDs(db *gorm.DB, projectIDs []string) ([]*model.ProjectMember, error)
	GetActiveMemberInProject(db *gorm.DB, projectID string, employeeID string) (*model.ProjectMember, error)
	GetActiveMembersBySlotID(db *gorm.DB, slotID string) ([]*model.ProjectMember, error)
	GetAssignedMembers(db *gorm.DB, projectID string, status string, preload bool) ([]*model.ProjectMember, error)
	IsExist(db *gorm.DB, id string) (bool, error)
	IsExistsByEmployeeID(db *gorm.DB, projectID string, employeeID string) (bool, error)
	OneByID(db *gorm.DB, id string) (*model.ProjectMember, error)
	OneBySlotID(db *gorm.DB, slotID string) (*model.ProjectMember, error)
	UpdateEndDateByProjectID(db *gorm.DB, projectID string) error
	UpdateEndDateOverdueMemberToInActive(db *gorm.DB) error
	UpdateLeftMemberToInActive(db *gorm.DB) error
	UpdateMemberInClosedProjectToInActive(db *gorm.DB) error
	UpdateSelectedFieldByProjectID(db *gorm.DB, projectID string, updateModel model.ProjectMember, updatedField string) error
	UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.ProjectMember, updatedFields ...string) (*model.ProjectMember, error)
}
