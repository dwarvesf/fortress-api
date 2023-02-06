package cronjob

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/cronjob/audit"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type syncAudit struct {
	audit audit.ICronjob
}

func NewSyncAuditJob(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) Cronjob {
	return &syncAudit{
		audit: audit.New(
			store, repo, service, logger, cfg,
		),
	}
}

func (j *syncAudit) Run() {
	j.audit.SyncAuditCycle()
}
