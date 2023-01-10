package audit

type ICronjob interface {
	SyncAuditCycle()
	SyncActionItem()
}
