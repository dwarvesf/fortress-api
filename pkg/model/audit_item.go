package model

import (
	"strings"

	"github.com/dstotijn/go-notion"
)

type AuditItemSeverity string

// values for audit_item_severity
const (
	AuditItemSeverityHigh   AuditItemSeverity = "high"
	AuditItemSeverityLow    AuditItemSeverity = "low"
	AuditItemSeverityMedium AuditItemSeverity = "medium"
)

// IsValid validation for AuditItemSeverity
func (e AuditItemSeverity) IsValid() bool {
	switch e {
	case
		AuditItemSeverityHigh,
		AuditItemSeverityLow,
		AuditItemSeverityMedium:
		return true
	}
	return false
}

// String returns the string representation
func (e AuditItemSeverity) String() string {
	return string(e)
}

// values for audit_area field
const (
	AuditItemAreaDelivery      string = "Delivery performance"
	AuditItemAreaQuality       string = "Quality assurance"
	AuditItemAreaCollaborating string = "Collaborating"
	AuditItemAreaFeedback      string = "Engineering feedback"
)

type AuditItem struct {
	BaseModel

	AuditID      UUID
	NotionDBID   UUID
	Name         string
	Area         string
	Requirements string
	Grade        int64
	Severity     *AuditItemSeverity
	Notes        string
	ActionItemID *UUID
}

func NewAuditItemFromNotionPage(page notion.Page, auditID string, notionDBID string) *AuditItem {
	properties := page.Properties.(notion.DatabasePageProperties)
	rs := &AuditItem{
		BaseModel:  BaseModel{ID: MustGetUUIDFromString(page.ID)},
		AuditID:    MustGetUUIDFromString(auditID),
		Name:       properties["Name"].Title[0].PlainText,
		NotionDBID: MustGetUUIDFromString(notionDBID),
	}

	if properties["Area"].Select != nil {
		rs.Area = properties["Area"].Select.Name
	}

	if properties["Grade"].Select != nil {
		rs.Grade = MappingAuditItemGrade(properties["Grade"].Select.Name)
	}

	if properties["Severity"].Select != nil {
		severity := AuditItemSeverity(strings.ToLower(properties["Severity"].Select.Name))
		if severity.IsValid() {
			rs.Severity = &severity
		}
	}

	if len(properties["Requirements"].RichText) > 0 {
		rs.Requirements = properties["Requirements"].RichText[0].PlainText
	}

	if len(properties["Notes"].RichText) > 0 {
		rs.Notes = properties["Notes"].RichText[0].PlainText
	}

	return rs
}

func MappingAuditItemGrade(auditGrade string) int64 {
	switch auditGrade {
	case "Very Good":
		return 5
	case "Good":
		return 4
	case "Acceptable":
		return 3
	case "Poor":
		return 2
	case "Very Poor":
		return 1
	default:
		return 0
	}
}

func CompareAuditItem(currAuditItem *AuditItem, newAuditItem *AuditItem) bool {
	return ((currAuditItem.Severity == nil && newAuditItem.Severity == nil) ||
		(currAuditItem.Severity != nil && newAuditItem.Severity != nil && *currAuditItem.Severity == *newAuditItem.Severity)) &&
		currAuditItem.Name == newAuditItem.Name && currAuditItem.Area == newAuditItem.Area &&
		currAuditItem.Requirements == newAuditItem.Requirements && currAuditItem.Grade == newAuditItem.Grade &&
		currAuditItem.Notes == newAuditItem.Notes && currAuditItem.AuditID == newAuditItem.AuditID && currAuditItem.NotionDBID == newAuditItem.NotionDBID
}

func AuditItemToMap(ai []*AuditItem) map[UUID]AuditItem {
	rs := make(map[UUID]AuditItem)
	for _, item := range ai {
		rs[item.ID] = *item
	}
	return rs
}
