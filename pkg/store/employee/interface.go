package employee

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Search(db *gorm.DB, query SearchFilter, pagination model.Pagination) (employees []*model.Employee, total int64, err error)
	One(db *gorm.DB, id string) (employee *model.Employee, err error)
	Create(db *gorm.DB, e *model.Employee) (employee *model.Employee, err error)
	OneByTeamEmail(db *gorm.DB, teamEmail string) (employee *model.Employee, err error)
	UpdateEmployeeStatus(db *gorm.DB, employeeID string, accountStatusID model.AccountStatus) (employee *model.Employee, err error)
	UpdateGeneralInfo(db *gorm.DB, body UpdateGeneralInfoInput, id string) (employee *model.Employee, err error)
	UpdatePersonalInfo(db *gorm.DB, body UpdatePersonalInfoInput, id string) (employee *model.Employee, err error)
	Update(db *gorm.DB, id string, employee *model.Employee) (*model.Employee, error)
	Exists(db *gorm.DB, id string) (bool, error)
	UpdateProfileInfo(db *gorm.DB, body UpdateProfileInforInput, id string) (*model.Employee, error)
}
