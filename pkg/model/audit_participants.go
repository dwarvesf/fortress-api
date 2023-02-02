package model

type AuditParticipant struct {
	BaseModel

	AuditID    UUID
	EmployeeID UUID
}

func AuditParticipantToMap(auditParticipant []*AuditParticipant) map[UUID]AuditParticipant {
	rs := make(map[UUID]AuditParticipant)
	for _, ap := range auditParticipant {
		rs[ap.EmployeeID] = *ap
	}

	return rs
}
