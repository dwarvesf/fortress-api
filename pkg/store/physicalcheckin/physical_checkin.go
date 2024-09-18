package physicalcheckin

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) Save(db *gorm.DB, tx *model.PhysicalCheckinTransaction) error {
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "employee_id"}, {Name: "date"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"icy_amount": tx.IcyAmount,
		}),
	}).Create(tx).Error
}

func (s *store) One(db *gorm.DB, id int) (*model.PhysicalCheckinTransaction, error) {
	var tx model.PhysicalCheckinTransaction
	err := db.Where("id = ?", id).First(&tx).Error
	return &tx, err
}

func (s *store) GetByEmployeeIDAndDate(db *gorm.DB, employeeID string, date string) (*model.PhysicalCheckinTransaction, error) {
	var tx model.PhysicalCheckinTransaction
	err := db.Where("employee_id = ? AND date = ?", employeeID, date).First(&tx).Error
	return &tx, err
}
