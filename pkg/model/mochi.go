package model

type VaultTransaction struct {
	ID          int64
	GuildID     string
	VaultID     int64
	VaultName   string
	Action      string
	FromAddress string
	ToAddress   string
	Target      string
	Sender      string
	Amount      string
	Token       string
	Threshold   string
	CreatedAt   string
	UpdatedAt   string
}
