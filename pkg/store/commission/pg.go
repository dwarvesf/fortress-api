package commission

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

// New initilize new store for commission
func New() IStore {
	return &store{}
}

func (s *store) Create(db *gorm.DB, us []model.EmployeeCommission) error {
	for i := range us {
		err := db.Create(&us[i]).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *store) Get(db *gorm.DB, q Query) ([]model.EmployeeCommission, error) {
	var res []model.EmployeeCommission
	if q.EmployeeID != "" {
		db = db.Where("employee_id = ?", q.EmployeeID)
	}
	if q.FromDate != nil {
		db = db.Where("created_at > ?", q.FromDate)
	}
	if q.ToDate != nil {
		db = db.Where("created_at < ?", q.ToDate)
	}
	return res, db.Preload("Invoice").Where("is_paid = ?", q.IsPaid).Find(&res).Error
}

func (s *store) MarkPaid(db *gorm.DB, id model.UUID) error {
	var cms model.EmployeeCommission
	err := db.Where("id = ?", id).Find(&cms).Error
	if err != nil {
		return err
	}
	return db.Model(&cms).Updates(map[string]interface{}{"is_paid": true, "paid_at": time.Now()}).Error
}
