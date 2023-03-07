package currency

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type IService interface {
	USDToVND(usd float64) (float64, error)
	VNDToUSD(vnd float64) (float64, error)

	Convert(value float64, target, dest string) (float64, float64, error)

	GetByName(db *gorm.DB, name string) (*model.Currency, error)
	GetByID(db *gorm.DB, id model.UUID) (*model.Currency, error)
	GetCurrencyOption(db *gorm.DB) ([]model.Currency, error)

	GetRate(target string) (float64, error)
}

const (
	// VNDCurrency : VietNam dong
	VNDCurrency = "VND"

	// USDCurrency : US dollar
	USDCurrency = "USD"

	// GBPCurrency :  British Pound
	GBPCurrency = "GBP"

	// SGDCurrency : Singapore Dollar
	SGDCurrency = "SGD"

	// EURCurrency : Europe dollar
	EURCurrency = "EUR"

	// EURCurrency : Canadian dollar
	CADCurrency = "CAD"
)
