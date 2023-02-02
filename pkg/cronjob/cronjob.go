package cronjob

import (
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/cronjob/audit"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type Cronjob struct {
	caller *cron.Cron
	logger logger.Logger
	audit  audit.ICronjob
}

func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) *Cronjob {
	return &Cronjob{
		caller: cron.New(),
		logger: logger,
		audit: audit.New(
			store, repo, service, logger, cfg,
		),
	}
}

func (c *Cronjob) Run() error {
	_, err := c.caller.AddFunc("@midnight", c.audit.SyncAuditCycle)
	if err != nil {
		return err
	}

	c.caller.Start()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			time.Sleep(1 * time.Hour)
			c.logger.Info("cronjob is running")
		}
	}()

	wg.Wait()
	return err
}
