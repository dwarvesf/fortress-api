package model

type ActionItemSnapshot struct {
	BaseModel

	ProjectID    UUID
	AuditCycleID UUID
	High         int64
	Medium       int64
	Low          int64
}

func CompareActionItemSnapshot(old, new *ActionItemSnapshot) bool {
	return old.High == new.High &&
		old.Medium == new.Medium &&
		old.Low == new.Low
}
