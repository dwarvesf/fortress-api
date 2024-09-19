package physicalcheckin

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, id int) (pc *model.PhysicalCheckinTransaction, err error)
	Save(db *gorm.DB, pc *model.PhysicalCheckinTransaction) (err error)
	GetByEmployeeIDAndDate(db *gorm.DB, employeeID string, date string) (*model.PhysicalCheckinTransaction, error)
}
