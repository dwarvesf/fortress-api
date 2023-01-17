package model

import (
	"strings"

	"github.com/dstotijn/go-notion"
)

type ActionItemStatus string

// values for action_item_status
const (
	ActionItemStatusPending   ActionItemStatus = "pending"
	ActionItemStatusInProgess ActionItemStatus = "in-progress"
	ActionItemStatusDone      ActionItemStatus = "done"
)

// IsValid validation for ActionItemStatus
func (e ActionItemStatus) IsValid() bool {
	switch e {
	case
		ActionItemStatusPending,
		ActionItemStatusInProgess,
		ActionItemStatusDone:
		return true
	}
	return false
}

// String returns the string representation
func (e ActionItemStatus) String() string {
	return string(e)
}

type ActionItemPriority string

// values for action_item_priority
const (
	ActionItemPriorityHigh   ActionItemPriority = "high"
	ActionItemPriorityLow    ActionItemPriority = "low"
	ActionItemPriorityMedium ActionItemPriority = "medium"
)

// IsValid validation for ActionItemPriority
func (e ActionItemPriority) IsValid() bool {
	switch e {
	case
		ActionItemPriorityHigh,
		ActionItemPriorityLow,
		ActionItemPriorityMedium:
		return true
	}
	return false
}

// String returns the string representation
func (e ActionItemPriority) String() string {
	return string(e)
}

type ActionItem struct {
	BaseModel

	ProjectID    UUID
	NotionDBID   UUID
	PICID        UUID
	AuditCycleID UUID
	Name         string
	Description  string
	NeedHelp     bool
	Priority     *ActionItemPriority
	Status       ActionItemStatus
}

func ActionItemToMap(actionItems []*ActionItem) map[UUID]*ActionItem {
	rs := map[UUID]*ActionItem{}
	for _, s := range actionItems {
		rs[s.ID] = s
	}

	return rs
}

func NewActionItemFromNotionPage(page notion.Page, picID UUID) *ActionItem {
	properties := page.Properties.(notion.DatabasePageProperties)

	rs := &ActionItem{
		BaseModel:  BaseModel{ID: MustGetUUIDFromString(page.ID)},
		NotionDBID: MustGetUUIDFromString(page.ID),
		Status:     ActionItemStatus(strings.ToLower(properties["Status"].Status.Name)),

		// TODO:Description:

	}

	if properties["Project"].Relation != nil && len(properties["Project"].Relation) > 0 {
		rs.ProjectID = MustGetUUIDFromString(properties["Project"].Relation[0].ID)
	}

	if !picID.IsZero() {
		rs.PICID = picID
	}

	if properties["NEED HELP???"].Checkbox != nil {
		rs.NeedHelp = *properties["NEED HELP???"].Checkbox
	}

	if properties["Name"].Title != nil {
		rs.Name = properties["Name"].Title[0].PlainText
	}

	if properties["Priority"].Select != nil {
		priority := MappingAuditActionPriority(properties["Priority"].Select.Name)
		if priority.IsValid() {
			rs.Priority = &priority
		}
	}

	if len(properties["👏 Audit Changelog"].Relation) > 0 {
		rs.AuditCycleID = MustGetUUIDFromString(properties["👏 Audit Changelog"].Relation[0].ID)
	}

	return rs
}

func MappingAuditActionPriority(auditGrade string) ActionItemPriority {
	switch auditGrade {
	case "1 - Critical":
		return ActionItemPriorityHigh
	case "2 - Medium":
		return ActionItemPriorityMedium
	case "3- Low":
		return ActionItemPriorityLow
	default:
		return ""
	}
}
