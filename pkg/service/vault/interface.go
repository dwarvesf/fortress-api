package vault

type IService interface {
	GetString(key string) string
	GetBool(key string) bool
}
