package redis

import (
	"fmt"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/utils/authutils"
	"github.com/go-redis/redis"
)

type Redis struct {
	client *redis.Client
	cfg    *config.Config
}

func New(cfg *config.Config) (RedisService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		DB:       0,
		Password: cfg.Redis.Password,
	})
	_, err := client.Ping().Result()
	return &Redis{
		client: client,
		cfg:    cfg,
	}, err
}

func (r *Redis) AddTokenBlacklist(token string) error {
	expiredAt, err := authutils.GetExpireAtFromToken(r.cfg, token)
	if err != nil {
		return err
	}
	_, err = r.client.Set(token, expiredAt, 24*time.Hour).Result()
	return err
}

func (r *Redis) GetAllBlacklistToken() ([]string, error) {
	return r.client.Keys("*").Result()
}
