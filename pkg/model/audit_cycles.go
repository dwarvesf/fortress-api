package model

import (
	"fmt"
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
	HealthAuditID     *UUID
	ProcessAuditID    *UUID
	BackendAuditID    *UUID
	FrontendAuditID   *UUID
	SystemAuditID     *UUID
	MobileAuditID     *UUID
	BlockchainAuditID *UUID
	Cycle             int64
	AverageScore      float64
	Status            AuditStatus
	Flag              AuditFlag
	Quarter           string
	ActionItemHigh    int64
	ActionItemMedium  int64
	ActionItemLow     int64
	SyncAt            *time.Time

	Project *Project
}

func AuditCycleToMap(auditCycles []*AuditCycle) map[UUID]*AuditCycle {
	rs := map[UUID]*AuditCycle{}
	for _, s := range auditCycles {
		rs[s.ID] = s
	}

	return rs
}

func NewAuditCycleFromNotionPage(page *notion.Page, notionDBID string) *AuditCycle {
	properties := page.Properties.(notion.DatabasePageProperties)
	now := time.Now()

	rs := &AuditCycle{
		BaseModel:  BaseModel{ID: MustGetUUIDFromString(page.ID)},
		NotionDBID: MustGetUUIDFromString(notionDBID),
		Status:     AuditStatusPending,
		Flag:       AuditFlag(strings.ToLower(properties["Flag"].Status.Name)),
		SyncAt:     &now,
	}

	if properties["Score"].Number != nil {
		rs.AverageScore = *properties["Score"].Number
	}

	if properties["Cycle"].Number != nil {
		rs.Cycle = int64(*properties["Cycle"].Number)
	}

	if properties["Date"].Date != nil {
		date := properties["Date"].Date.Start.Time
		rs.Quarter = fmt.Sprintf("%d/Q%d", date.Year(), (date.Month()-1)/3+1)
	} else {
		date := time.Now()
		rs.Quarter = fmt.Sprintf("%d/Q%d", date.Year(), (date.Month()-1)/3+1)
	}

	if len(properties["Project"].Relation) > 0 {
		rs.ProjectID = MustGetUUIDFromString(properties["Project"].Relation[0].ID)
	}

	return rs
}

func AuditMap(ac AuditCycle) map[UUID]AuditType {
	rs := make(map[UUID]AuditType)

	if !ac.HealthAuditID.IsZero() {
		rs[*ac.HealthAuditID] = AuditTypeHealth
	}

	if !ac.ProcessAuditID.IsZero() {
		rs[*ac.ProcessAuditID] = AuditTypeProcess
	}

	if !ac.BackendAuditID.IsZero() {
		rs[*ac.BackendAuditID] = AuditTypeBackend
	}

	if !ac.FrontendAuditID.IsZero() {
		rs[*ac.FrontendAuditID] = AuditTypeFrontend
	}

	if !ac.SystemAuditID.IsZero() {
		rs[*ac.SystemAuditID] = AuditTypeSystem
	}

	if !ac.MobileAuditID.IsZero() {
		rs[*ac.MobileAuditID] = AuditTypeMobile
	}

	if !ac.BlockchainAuditID.IsZero() {
		rs[*ac.BlockchainAuditID] = AuditTypeBlockchain
	}

	return rs
}

func CheckTypeExists(auditMap map[UUID]AuditType, auditType AuditType) UUID {
	for k, v := range auditMap {
		if v == auditType {
			return k
		}
	}

	return UUID{}
}

func CompareAuditCycle(currAC *AuditCycle, newAC *AuditCycle) bool {
	return currAC.ProjectID == newAC.ProjectID && currAC.NotionDBID == newAC.NotionDBID &&
		currAC.Cycle == newAC.Cycle && currAC.AverageScore == newAC.AverageScore &&
		currAC.Flag == newAC.Flag && currAC.Quarter == newAC.Quarter
}
