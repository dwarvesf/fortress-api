package cachedpayroll

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type IStore interface {
	Set(db *gorm.DB, cachedPayroll *model.CachedPayroll) error
	Get(db *gorm.DB, month, year, batch int) (*model.CachedPayroll, error)
}
