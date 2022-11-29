package employee

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB, input GetAllInput, pagination model.Pagination) (employees []*model.Employee, total int64, err error)
	Create(db *gorm.DB, e *model.Employee) (employee *model.Employee, err error)

	One(db *gorm.DB, id string) (employee *model.Employee, err error)
	OneByTeamEmail(db *gorm.DB, teamEmail string) (employee *model.Employee, err error)
	GetByIDs(db *gorm.DB, ids []string) (employees []*model.Employee, err error)

	IsExist(db *gorm.DB, id string) (bool, error)

	Update(db *gorm.DB, employee *model.Employee) (*model.Employee, error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.Employee, updatedFields ...string) (*model.Employee, error)
}
