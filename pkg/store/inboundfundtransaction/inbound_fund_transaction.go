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

// DeleteUnpaidByInvoiceID delete all inbound fund transactions which is not paid and by invoice id
func (s *store) DeleteUnpaidByInvoiceID(db *gorm.DB, invoiceID string) error {
	return db.Where("invoice_id = ? AND paid_at IS NULL", invoiceID).Delete(&model.InboundFundTransaction{}).Error
}

func (s *store) GetByInvoiceID(db *gorm.DB, invoiceID string) (*model.InboundFundTransaction, error) {
	var ift model.InboundFundTransaction
	if err := db.Where("invoice_id = ?", invoiceID).First(&ift).Error; err != nil {
		return nil, err
	}
	return &ift, nil
}
