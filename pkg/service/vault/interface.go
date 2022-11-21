package vault

import "github.com/dwarvesf/fortress-api/pkg/config"

type VaultService interface {
	LoadConfig() *config.Config
}
