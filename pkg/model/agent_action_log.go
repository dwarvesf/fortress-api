package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AgentActionLogStatus represents the status of an agent action
type AgentActionLogStatus string

const (
	AgentActionLogStatusSuccess AgentActionLogStatus = "success"
	AgentActionLogStatusError   AgentActionLogStatus = "error"
	AgentActionLogStatusTimeout AgentActionLogStatus = "timeout"
)

// AgentActionLog represents a log entry for agent actions
type AgentActionLog struct {
	ID           UUID                     `json:"id" gorm:"type:uuid;primary_key;default:uuid()"`
	CreatedAt    time.Time                `json:"createdAt"`
	AgentKeyID   UUID                     `json:"agentKeyId" gorm:"column:agent_key_id"`
	ToolName     string                   `json:"toolName" gorm:"not null;column:tool_name"`
	InputData    datatypes.JSON           `json:"inputData,omitempty" gorm:"type:jsonb;column:input_data"`
	OutputData   datatypes.JSON           `json:"outputData,omitempty" gorm:"type:jsonb;column:output_data"`
	Status       AgentActionLogStatus     `json:"status" gorm:"not null"`
	DurationMs   *int                     `json:"durationMs,omitempty" gorm:"column:duration_ms"`
	ErrorMessage *string                  `json:"errorMessage,omitempty" gorm:"column:error_message"`

	// Relationships
	AgentAPIKey *AgentAPIKey `json:"agentApiKey,omitempty" gorm:"foreignKey:AgentKeyID;references:ID"`
}

// TableName returns the table name for AgentActionLog
func (AgentActionLog) TableName() string {
	return "agent_action_logs"
}

// BeforeCreate sets the ID and CreatedAt before creating
func (a *AgentActionLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID.IsZero() {
		a.ID = NewUUID()
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}
	return nil
}

// IsSuccess checks if the action was successful
func (a *AgentActionLog) IsSuccess() bool {
	return a.Status == AgentActionLogStatusSuccess
}

// HasError checks if the action resulted in an error
func (a *AgentActionLog) HasError() bool {
	return a.Status == AgentActionLogStatusError
}

// GetDurationSeconds returns the duration in seconds
func (a *AgentActionLog) GetDurationSeconds() float64 {
	if a.DurationMs == nil {
		return 0
	}
	return float64(*a.DurationMs) / 1000.0
}