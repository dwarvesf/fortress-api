package inboundfundtransaction

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, ift *model.InboundFundTransaction) (*model.InboundFundTransaction, error)
	DeleteUnpaidByInvoiceID(db *gorm.DB, invoiceID string) error
	GetByInvoiceID(db *gorm.DB, invoiceID string) (*model.InboundFundTransaction, error)
	Get(db *gorm.DB, q Query) ([]model.InboundFundTransaction, error)
}
