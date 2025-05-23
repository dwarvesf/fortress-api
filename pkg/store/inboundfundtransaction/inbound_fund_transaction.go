package inboundfundtransaction

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) Create(db *gorm.DB, ift *model.InboundFundTransaction) (*model.InboundFundTransaction, error) {
	return ift, db.Create(ift).Error
}
