package vault

import (
	"fmt"
	"strconv"

	"github.com/dwarvesf/fortress-api/pkg/config"
	vault "github.com/hashicorp/vault/api"
)

type Vault struct {
	data map[string]interface{}
}

func New(cfg *config.Config) (VaultService, error) {
	config := vault.DefaultConfig()
	config.Address = cfg.Vault.Address

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %v", err)
	}

	client.SetToken(cfg.Vault.Token)

	secret, err := client.Logical().Read(cfg.Vault.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to read secret: %v", err)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unable to read secret data")
	}

	return &Vault{
		data: data,
	}, nil
}

func (v *Vault) LoadConfig() *config.Config {
	return &config.Config{
		Debug:  v.GetBool("DEBUG"),
		APIKey: v.GetString("API_KEY"),

		ApiServer: config.ApiServer{
			Port:           v.GetString("PORT"),
			AllowedOrigins: v.GetString("ALLOWED_ORIGINS"),
		},

		Postgres: config.DBConnection{
			Host:    v.GetString("DB_HOST"),
			Port:    v.GetString("DB_PORT"),
			User:    v.GetString("DB_USER"),
			Name:    v.GetString("DB_NAME"),
			Pass:    v.GetString("DB_PASS"),
			SSLMode: v.GetString("DB_SSL_MODE"),
		},

		Google: config.Google{
			ClientSecret: v.GetString("GOOGLE_API_CLIENT_SECRET"),
			ClientID:     v.GetString("GOOGLE_API_CLIENT_ID"),
			AppName:      v.GetString("GOOGLE_API_APP_NAME"),
		},
	}
}

func (v *Vault) GetString(key string) string {
	value, _ := v.data[key].(string)
	return value
}

func (v *Vault) GetBool(key string) bool {
	data, _ := v.data[key].(string)
	value, _ := strconv.ParseBool(data)
	return value
}
