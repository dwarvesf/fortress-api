package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type ListBankAccountResponse struct {
	Data []model.BankAccount `json:"data"`
}
