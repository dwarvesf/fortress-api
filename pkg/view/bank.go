package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type ListBankAccountResponse struct {
	Data []BankAccount `json:"data"`
}

type BankAccount struct {
	ID            string   `json:"id"`
	AccountNumber string   `json:"accountNumber"`
	BankName      string   `json:"bankName"`
	OwnerName     string   `json:"ownerName"`
	Address       *string  `json:"address"`
	SwiftCode     string   `json:"swiftCode"`
	RoutingNumber string   `json:"routingNumber"`
	Name          string   `json:"name"`
	UKSortCode    string   `json:"ukSortCode"`
	CurrencyID    string   `json:"currencyID"`
	Currency      Currency `json:"currency"`
}

func ToListBankAccount(accounts []*model.BankAccount) []BankAccount {
	res := make([]BankAccount, 0)

	for _, acc := range accounts {
		res = append(res, BankAccount{
			ID:            acc.ID.String(),
			AccountNumber: acc.AccountNumber,
			BankName:      acc.BankName,
			OwnerName:     acc.OwnerName,
			Address:       acc.Address,
			SwiftCode:     acc.SwiftCode,
			RoutingNumber: acc.RoutingNumber,
			Name:          acc.Name,
			UKSortCode:    acc.UKSortCode,
			CurrencyID:    acc.CurrencyID.String(),
			Currency:      toCurrency(acc.Currency),
		})
	}

	return res
}
