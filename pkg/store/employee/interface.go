package employee

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB, input EmployeeFilter, pagination model.Pagination) (employees []*model.Employee, total int64, err error)
	Create(db *gorm.DB, e *model.Employee) (employee *model.Employee, err error)

	One(db *gorm.DB, id string, preload bool) (employee *model.Employee, err error)
	OneByUsername(db *gorm.DB, username string, preload bool) (employee *model.Employee, err error)
	OneByTeamEmail(db *gorm.DB, teamEmail string) (employee *model.Employee, err error)
	OneByEmail(db *gorm.DB, email string) (*model.Employee, error)
	OneByNotionID(db *gorm.DB, notionID string) (employee *model.Employee, err error)
	GetByIDs(db *gorm.DB, ids []model.UUID) (employees []*model.Employee, err error)
	GetByWorkingStatus(db *gorm.DB, workingStatus model.WorkingStatus) ([]*model.Employee, error)
	GetLineManagers(db *gorm.DB) ([]*model.Employee, error)
	GetLineManagersOfPeers(db *gorm.DB, employeeID string) ([]*model.Employee, error)

	IsExist(db *gorm.DB, id string) (bool, error)

	Update(db *gorm.DB, employee *model.Employee) (*model.Employee, error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.Employee, updatedFields ...string) (*model.Employee, error)
}
