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

func New(cfg *config.Config) (IService, error) {
	defaultConfig := vault.DefaultConfig()
	defaultConfig.Address = cfg.Vault.Address

	client, err := vault.NewClient(defaultConfig)
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

func (v *Vault) GetString(key string) string {
	value, _ := v.data[key].(string)
	return value
}

func (v *Vault) GetBool(key string) bool {
	data, _ := v.data[key].(string)
	value, _ := strconv.ParseBool(data)
	return value
}
