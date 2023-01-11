package model

import (
	"strings"
	"time"

	"github.com/dstotijn/go-notion"
)

type AuditStatus string

// values for audit_status
const (
	AuditStatusPending AuditStatus = "pending"
	AuditStatusAudited AuditStatus = "audited"
)

// IsValid validation for AuditStatus
func (e AuditStatus) IsValid() bool {
	switch e {
	case
		AuditStatusPending,
		AuditStatusAudited:
		return true
	}
	return false
}

// String returns the string representation
func (e AuditStatus) String() string {
	return string(e)
}

type AuditFlag string

// values for audit_flag
const (
	AuditFlagRed    AuditFlag = "red"
	AuditFlagYellow AuditFlag = "yellow"
	AuditFlagGreen  AuditFlag = "green"
	AuditFlagNone   AuditFlag = "none"
)

// IsValid validation for AuditFlag
func (e AuditFlag) IsValid() bool {
	switch e {
	case
		AuditFlagRed,
		AuditFlagYellow,
		AuditFlagGreen,
		AuditFlagNone:
		return true
	}
	return false
}

// String returns the string representation
func (e AuditFlag) String() string {
	return string(e)
}

type AuditCycle struct {
	BaseModel

	ProjectID         UUID
	NotionDBID        UUID
	HealthAuditID     UUID
	ProcessAuditID    UUID
	BackendAuditID    UUID
	FrontendAuditID   UUID
	SystemAuditID     UUID
	MobileAuditID     UUID
	BlockchainAuditID UUID
	Cycle             int64
	AverageScore      float64
	Status            AuditStatus
	Flag              AuditFlag
	Quarter           string
	ActionItemHigh    int64
	ActionItemMedium  int64
	ActionItemLow     int64
	SyncAt            *time.Time
}

func AuditCycleToMap(auditCycles []*AuditCycle) map[UUID]*AuditCycle {
	rs := map[UUID]*AuditCycle{}
	for _, s := range auditCycles {
		rs[s.ID] = s
	}

	return rs
}

func NewAuditCycleFromNotionPage(page *notion.Page) *AuditCycle {
	properties := page.Properties.(notion.DatabasePageProperties)
	now := time.Now()

	rs := &AuditCycle{
		BaseModel:  BaseModel{ID: MustGetUUIDFromString(page.ID)},
		ProjectID:  MustGetUUIDFromString(properties["Project"].Relation[0].ID),
		NotionDBID: MustGetUUIDFromString(page.ID),
		Status:     AuditStatusPending, //TODO: quarter
		Flag:       AuditFlag(strings.ToLower(properties["Flag"].Status.Name)),
		SyncAt:     &now,
	}

	if properties["Score"].Number != nil {
		rs.AverageScore = *properties["Score"].Number
	}

	if properties["Cycle"].Number != nil {
		rs.Cycle = int64(*properties["Cycle"].Number)
	}

	return rs
}
