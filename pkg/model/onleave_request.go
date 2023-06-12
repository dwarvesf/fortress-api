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

	Creator  *Employee
	Approver *Employee
}
