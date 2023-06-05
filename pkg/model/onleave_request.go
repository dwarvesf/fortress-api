package model

import (
	"time"
)

type OnLeaveRequest struct {
	BaseModel

	Name        string    `json:"name" gorm:"column:name"`
	OffType     string    `json:"off_type" gorm:"column:off_type"`
	StartDate   time.Time `json:"start_date" gorm:"column:start_date"`
	EndDate     time.Time `json:"end_date" gorm:"column:end_date"`
	Shift       string    `json:"shift" gorm:"column:shift"`
	Title       string    `json:"title" gorm:"column:title"`
	Description string    `json:"description" gorm:"column:description"`
	CreatorID   UUID      `json:"creator_id" gorm:"column:creator_id"`
	ApprovedID  UUID      `json:"approver_id" gorm:"column:approver_id"`
	AssigneeIDs []string  `json:"assignee_ids" gorm:"column:assignee_ids"`
}
