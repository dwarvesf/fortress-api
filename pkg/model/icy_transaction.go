package model

type IcyTransaction struct {
	BaseModel
	Vault              string `json:"vault"`
	Amount             string `json:"amount"`
	Token              string `json:"token"`
	SenderDiscordId    string `json:"sender_discord_id"`
	RecipientAddress   string `json:"recipient_address"`
	RecipientDiscordId string `json:"recipient_discord_id"`
}
