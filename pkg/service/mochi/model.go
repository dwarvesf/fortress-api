package mochi

// Vault is the model for mochi vault that is fetched from mochi-api
type Vault struct {
	ID      int64  `json:"id"`
	GuildID string `json:"guild_id"`
}
