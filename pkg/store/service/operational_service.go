package service

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
	"time"
)

type store struct{}

func New() IStore {
	return &store{}
}
func (s store) FindOperationByMonth(db *gorm.DB, month time.Month) ([]*model.OperationalService, error) {
	var res []*model.OperationalService
	err := db.Table("operational_services").Preload("Currency").Where("is_active is true and date_part('month',register_date) = ?", month).Find(&res).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return res, nil
}
