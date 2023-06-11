package icytransaction

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) Create(db *gorm.DB, model []model.IcyTransaction) error {
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "sender"},
			{Name: "target"},
			{Name: "category"},
			{Name: "amount"},
			{Name: "txn_time"},
		},
		UpdateAll: true,
	}).Create(&model).Error
}
