package model

import (
	"time"
)

type OnLeaveRequest struct {
	BaseModel

	Type        string
	StartDate   *time.Time
	EndDate     *time.Time
	Shift       string
	Title       string
	Description string
	CreatorID   UUID
	ApproverID  UUID
	AssigneeIDs JSONArrayString
	NocodbID     *int    `gorm:"column:nocodb_id;index:idx_on_leave_requests_nocodb_id"`
	NotionPageID *string `gorm:"column:notion_page_id;index:idx_on_leave_requests_notion_page_id"`

	Creator  *Employee
	Approver *Employee
}
