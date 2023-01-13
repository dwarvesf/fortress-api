package audit

type ICronjob interface {
	SyncAuditCycle()
}
