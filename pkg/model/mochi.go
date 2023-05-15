package model

type VaultTransactionRequest struct {
	VaultId   string `json:"vault_id"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type VaultTransactionResponse struct {
	Data []VaultTransaction `json:"data"`
}

type VaultTransaction struct {
	Id          int64  `json:"id"`
	GuildId     string `json:"guild_id"`
	VaultId     int64  `json:"vault_id"`
	VaultName   string `json:"vault_name"`
	Action      string `json:"action"`
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Target      string `json:"target"`
	Sender      string `json:"sender"`
	Amount      string `json:"amount"`
	Token       string `json:"token"`
	Threshold   string `json:"threshold"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
