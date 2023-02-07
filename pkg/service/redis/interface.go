package redis

type RedisService interface {
	AddTokenBlacklist(token string) error
	GetAllBlacklistToken() ([]string, error)
}
