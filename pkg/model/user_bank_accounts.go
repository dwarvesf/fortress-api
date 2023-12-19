package model

type UserBankAccount struct {
	BaseModel

	EmployerID       UUID
	DiscordAccountID UUID
	BankID           UUID
	AccountNumber    string
	Branch           string
}
