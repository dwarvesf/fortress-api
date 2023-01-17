package model

type AuditParticipant struct {
	BaseModel

	AuditID    UUID
	EmployeeID UUID
}
