package model

type ProjectCommissionObject struct {
	BaseModel

	ProjectCommissionID UUID
	ProjectMemberID     UUID

	ProjectCommission         ProjectCommission
	ProjectCommissionReceiver ProjectCommissionReceiver `gorm:"foreignkey:project_commission_id;references:project_commission_id"`

	InvoiceItem InvoiceItem `gorm:"-"`
}
