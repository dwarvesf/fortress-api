package model

import (
	"time"

	"gorm.io/datatypes"
)

const (
	AccountingSE     = "SE"
	AccountingOP     = "OP"
	AccountingOV     = "OV"
	AccountingCA     = "CA"
	AccountingIncome = "In"

	AccountingOfficeSupply   = "Office Supply"
	AccountingOfficeServices = "Office Services"
	AccountingOfficeSpace    = "Office Space"
	AccountingIn             = "In"
	AccountingTools          = "Tools"
	AccountingAssets         = "Assets"

	AccountingOps         = "Payroll for Operation"
	AccountingRec         = "Payroll for Recruit"
	AccountingEng         = "Payroll for Engineer"
	AccountingMar         = "Payroll for Marketing"
	AccountingVen         = "Payroll for Venture"
	AccountingMng         = "Payroll for Middle Mngr"
	AccountingSal         = "Payroll for Sales"
	AccountingDsg         = "Payroll for Design"
	AccountingCommLead    = "Commission for Lead"
	AccountingCommSales   = "Commission for Sales"
	AccountingCommAccount = "Commission for Account"
	AccountingCommHiring  = "Commission for Hiring"
)

// AccountingTransaction --
type AccountingTransaction struct {
	BaseModel

	Name             string         `json:"name"`
	Date             *time.Time     `json:"date"`
	Amount           float64        `json:"amount"`
	ConversionAmount VietnamDong    `json:"conversion_amount"`
	Organization     string         `json:"organization"`
	Category         string         `json:"category_name"`
	Type             string         `json:"type"`
	Currency         string         `json:"currency_name"`
	CurrencyID       *UUID          `json:"-"`
	ConversionRate   float64        `json:"conversion_rate"`
	Metadata         datatypes.JSON `json:"metadata"`

	CurrencyInfo       *Currency           `json:"currency" gorm:"foreignkey:ID;association_foreignkey:CurrencyID"`
	AccountingCategory *AccountingCategory `json:"category" gorm:"foreignkey:Type;association_foreignkey:Type"`
}

type AccountingCategory struct {
	BaseModel
	Name string `json:"name"`
	Type string `json:"type"`
}

type SheetExpense struct {
	Name     string `json:"name"`
	Amount   string `json:"amount"`
	Category string `json:"category"`
	Currency string `json:"currency"`
	Date     string `json:"date"`
}

type AccountingMetadata struct {
	Source string `json:"source"`
	ID     string `json:"id"`
}
