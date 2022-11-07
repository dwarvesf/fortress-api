package accountstatus

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
	db *gorm.DB
}

func New(db *gorm.DB) IStore {
	return &store{
		db: db,
	}
}

// One get all Senitorities
func (s *store) All() ([]*model.AccountStatus, error) {
	var accountStatuses []*model.AccountStatus
	return accountStatuses, s.db.Find(&accountStatuses).Error
}
