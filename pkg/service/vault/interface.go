package vault

type VaultService interface {
	GetString(key string) string
	GetBool(key string) bool
}
