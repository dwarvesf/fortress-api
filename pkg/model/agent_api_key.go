package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AgentAPIKey represents an API key for agent authentication
type AgentAPIKey struct {
	BaseModel

	Name        string                 `json:"name" gorm:"not null"`
	APIKey      string                 `json:"-" gorm:"unique;not null;column:api_key"` // Hidden from JSON
	Permissions datatypes.JSON         `json:"permissions" gorm:"type:jsonb;default:'[]'"`
	RateLimit   int                    `json:"rateLimit" gorm:"default:1000;column:rate_limit"`
	IsActive    bool                   `json:"isActive" gorm:"default:true;column:is_active"`
	ExpiresAt   *time.Time             `json:"expiresAt,omitempty" gorm:"column:expires_at"`

	// Relationships
	ActionLogs []AgentActionLog `json:"actionLogs,omitempty" gorm:"foreignKey:AgentKeyID"`
	Workflows  []AgentWorkflow  `json:"workflows,omitempty" gorm:"foreignKey:AgentKeyID"`
}

// TableName returns the table name for AgentAPIKey
func (AgentAPIKey) TableName() string {
	return "agent_api_keys"
}

// IsExpired checks if the API key has expired
func (a *AgentAPIKey) IsExpired() bool {
	if a.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*a.ExpiresAt)
}

// IsValid checks if the API key is active and not expired
func (a *AgentAPIKey) IsValid() bool {
	return a.IsActive && !a.IsExpired()
}

// BeforeCreate sets the ID before creating
func (a *AgentAPIKey) BeforeCreate(tx *gorm.DB) error {
	if a.ID.IsZero() {
		a.ID = NewUUID()
	}
	return nil
}