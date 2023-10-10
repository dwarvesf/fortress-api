package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type ListBankAccountResponse struct {
	Data []BankAccount `json:"data"`
} // @name ListBankAccountResponse

type BankAccount struct {
	ID                      string   `json:"id"`
	AccountNumber           string   `json:"accountNumber"`
	BankName                string   `json:"bankName"`
	OwnerName               string   `json:"ownerName"`
	Address                 *string  `json:"address"`
	SwiftCode               string   `json:"swiftCode"`
	IntermediaryBankAddress string   `json:"intermediaryBankAddress"`
	IntermediaryBankName    string   `json:"intermediaryBankName"`
	RoutingNumber           string   `json:"routingNumber"`
	Name                    string   `json:"name"`
	UKSortCode              string   `json:"ukSortCode"`
	CurrencyID              string   `json:"currencyID"`
	Currency                Currency `json:"currency"`
} // @name BankAccount

func ToBankAccount(account *model.BankAccount) *BankAccount {
	return &BankAccount{
		ID:                      account.ID.String(),
		AccountNumber:           account.AccountNumber,
		BankName:                account.BankName,
		OwnerName:               account.OwnerName,
		Address:                 account.Address,
		SwiftCode:               account.SwiftCode,
		RoutingNumber:           account.RoutingNumber,
		Name:                    account.Name,
		UKSortCode:              account.UKSortCode,
		IntermediaryBankName:    account.IntermediaryBankName,
		IntermediaryBankAddress: account.IntermediaryBankAddress,
		CurrencyID:              account.CurrencyID.String(),
		Currency:                *toCurrency(account.Currency),
	}
}

func ToListBankAccount(accounts []*model.BankAccount) []BankAccount {
	res := make([]BankAccount, 0)

	for _, acc := range accounts {
		res = append(res, BankAccount{
			ID:                      acc.ID.String(),
			AccountNumber:           acc.AccountNumber,
			BankName:                acc.BankName,
			OwnerName:               acc.OwnerName,
			Address:                 acc.Address,
			SwiftCode:               acc.SwiftCode,
			RoutingNumber:           acc.RoutingNumber,
			Name:                    acc.Name,
			UKSortCode:              acc.UKSortCode,
			IntermediaryBankName:    acc.IntermediaryBankName,
			IntermediaryBankAddress: acc.IntermediaryBankAddress,
			CurrencyID:              acc.CurrencyID.String(),
			Currency:                *toCurrency(acc.Currency),
		})
	}

	return res
}
