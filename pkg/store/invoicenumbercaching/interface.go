package invoicenumbercaching

import (
	"time"

	"gorm.io/gorm"
)

type IStore interface {
	Set(db *gorm.DB, key string) error
	Decrease(db *gorm.DB, key string) error
	GetNext(db *gorm.DB, key string) (int, error)

	// NextProjectTemplateNumber TODO: (hnh), might be we just need GetNext(key) fn to get these things ?
	NextProjectTemplateNumber(db *gorm.DB, pid string) (int, error)

	UpdateInvoiceCachingNumber(db *gorm.DB, issuedAt time.Time, projectAlias string) error
	UnCountErrorInvoice(db *gorm.DB, issuedAt time.Time) error
}
