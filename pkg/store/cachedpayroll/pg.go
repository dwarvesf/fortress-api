package cachedpayroll

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

// New create new pg service
func New() IStore {
	return &store{}
}

func (s *store) Set(db *gorm.DB, cachedPayroll *model.CachedPayroll) error {
	return db.Save(cachedPayroll).Error
}

func (s *store) Get(db *gorm.DB, month, year, batch int) (*model.CachedPayroll, error) {
	p := model.CachedPayroll{}
	return &p, db.Where("month = ? AND year = ? AND batch = ?", month, year, batch).First(&p).Error
}
