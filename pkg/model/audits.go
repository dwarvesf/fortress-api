package model

import (
	"strings"
	"time"

	"github.com/dstotijn/go-notion"
)

type AuditType string

// values for audit_type
const (
	AuditTypeHealth     AuditType = "engineering-health"
	AuditTypeProcess    AuditType = "engineering-process"
	AuditTypeFrontend   AuditType = "frontend"
	AuditTypeBackend    AuditType = "backend"
	AuditTypeSystem     AuditType = "system"
	AuditTypeMobile     AuditType = "mobile"
	AuditTypeBlockchain AuditType = "blockchain"
)

// IsValid validation for AuditType
func (e AuditType) IsValid() bool {
	switch e {
	case
		AuditTypeHealth,
		AuditTypeProcess,
		AuditTypeFrontend,
		AuditTypeBackend,
		AuditTypeSystem,
		AuditTypeMobile,
		AuditTypeBlockchain:
		return true
	}
	return false
}

// String returns the string representation
func (e AuditType) String() string {
	return string(e)
}

type Audit struct {
	BaseModel

	ProjectID  UUID
	NotionDBID UUID
	AuditorID  UUID
	Name       string
	Type       AuditType
	Score      float64
	Status     AuditStatus
	Flag       AuditFlag
	ActionItem int64
	Duration   float64
	AuditedAt  *time.Time
	SyncAt     *time.Time
}

func NewAuditFromNotionPage(page notion.Page, projectID string, auditorID UUID, flag AuditFlag, notionDBID string) *Audit {
	properties := page.Properties.(notion.DatabasePageProperties)
	now := time.Now()

	rs := &Audit{
		BaseModel:  BaseModel{ID: MustGetUUIDFromString(page.ID)},
		NotionDBID: MustGetUUIDFromString(notionDBID),
		Name:       properties["Name"].Title[0].PlainText,
		SyncAt:     &now,
		Flag:       flag,
	}

	if !auditorID.IsZero() {
		rs.AuditorID = auditorID
	}

	if properties["Score"].Number != nil {
		rs.Score = *properties["Score"].Number
		rs.Status = AuditStatusAudited
	} else {
		rs.Status = AuditStatusPending
	}

	if properties["Duration (hours)"].Number != nil {
		rs.Duration = *properties["Duration (hours)"].Number
	}

	if properties["Date"].Date != nil {
		rs.AuditedAt = &properties["Date"].Date.Start.Time
	}

	if len(properties["Name"].Title) > 0 {
		if MappingAuditType(properties["Name"].Title[0].PlainText) != "" {
			rs.Type = MappingAuditType(properties["Name"].Title[0].PlainText)
		} else {
			return nil
		}
	} else {
		return nil
	}

	if projectID != "" {
		rs.ProjectID = MustGetUUIDFromString(projectID)
	}

	return rs
}

func MappingAuditType(auditType string) AuditType {
	switch strings.ToLower(auditType) {
	case "engineering health checklist":
		return AuditTypeHealth
	case "engineering process checklist":
		return AuditTypeProcess
	case "frontend checklist":
		return AuditTypeFrontend
	case "backend checklist":
		return AuditTypeBackend
	case "system checklist":
		return AuditTypeSystem
	case "mobile checklist":
		return AuditTypeMobile
	case "blockchain checklist":
		return AuditTypeBlockchain
	}

	return ""
}

func CompareAudit(currAudit Audit, newAudit Audit) bool {
	return ((currAudit.AuditedAt == nil && newAudit.AuditedAt == nil) ||
		(currAudit.AuditedAt != nil && newAudit.AuditedAt != nil && currAudit.AuditedAt.Equal(*newAudit.AuditedAt))) &&
		currAudit.ProjectID == newAudit.ProjectID &&
		currAudit.NotionDBID == newAudit.NotionDBID && currAudit.AuditorID == newAudit.AuditorID &&
		currAudit.Name == newAudit.Name && currAudit.Type == newAudit.Type &&
		currAudit.Score == newAudit.Score && currAudit.Status == newAudit.Status &&
		currAudit.Flag == newAudit.Flag && currAudit.Duration == newAudit.Duration
}
