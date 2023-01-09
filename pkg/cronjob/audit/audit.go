package audit

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type cronjob struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
	repo    store.DBRepo
	config  *config.Config
}

func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) ICronjob {
	return &cronjob{store: store, repo: repo, service: service, logger: logger, config: cfg}
}

func (c *cronjob) SyncAuditCycle() {
	//TODO: sync audit cycle with notion here
}

func (c *cronjob) SyncActionItem() {
	//TODO: sync audit action item with notion here
}
