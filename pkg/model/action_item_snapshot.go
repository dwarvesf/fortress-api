package model

type ActionItemSnapshot struct {
	BaseModel

	ProjectID    UUID
	AuditCycleID UUID
	High         int64
	Medium       int64
	Low          int64
}
