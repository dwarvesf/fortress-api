package model

type ProjectNotion struct {
	BaseModel

	ProjectID     UUID
	AuditNotionID UUID

	Project *Project `gorm:"foreignKey:project_id"`
}
