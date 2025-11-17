package nocodb

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	nocodbsvc "github.com/dwarvesf/fortress-api/pkg/service/nocodb"
	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"
	"github.com/dwarvesf/fortress-api/pkg/service/wise"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

// BaseProvider contains shared functionality for all NocoDB providers
type BaseProvider struct {
	svc        *nocodbsvc.Service
	expenseCfg config.ExpenseIntegration
	store      *store.Store
	repo       store.DBRepo
	cfg        *config.Config
	wise       wise.IService
	l          logger.Logger
}

// NewBaseProvider creates a new base provider with shared dependencies
func NewBaseProvider(svc *nocodbsvc.Service, cfg *config.Config, s *store.Store, repo store.DBRepo, wiseService wise.IService, l logger.Logger) *BaseProvider {
	if svc == nil {
		return nil
	}
	var expCfg config.ExpenseIntegration
	if cfg != nil {
		expCfg = cfg.ExpenseIntegration
	}
	return &BaseProvider{
		svc:        svc,
		expenseCfg: expCfg,
		store:      s,
		repo:       repo,
		cfg:        cfg,
		wise:       wiseService,
		l:          l,
	}
}

// Type returns the provider type for all NocoDB providers
func (p *BaseProvider) Type() taskprovider.ProviderType {
	return taskprovider.ProviderNocoDB
}