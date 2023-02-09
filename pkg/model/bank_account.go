package model

// BankAccount contain company information
type BankAccount struct {
	BaseModel

	AccountNumber string
	BankName      string
	OwnerName     string
	Address       *string
	SwiftCode     string
	RoutingNumber string
	Name          string
	UKSortCode    string

	CurrencyID UUID
	Currency   Currency
}
