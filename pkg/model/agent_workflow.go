package model

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AgentWorkflowStatus represents the status of an agent workflow
type AgentWorkflowStatus string

const (
	AgentWorkflowStatusPending     AgentWorkflowStatus = "pending"
	AgentWorkflowStatusInProgress  AgentWorkflowStatus = "in_progress"
	AgentWorkflowStatusCompleted   AgentWorkflowStatus = "completed"
	AgentWorkflowStatusFailed      AgentWorkflowStatus = "failed"
)

// AgentWorkflow represents a multi-step workflow executed by an agent
type AgentWorkflow struct {
	BaseModel

	WorkflowType    string              `json:"workflowType" gorm:"not null;column:workflow_type"`
	Status          AgentWorkflowStatus `json:"status" gorm:"not null"`
	InputData       datatypes.JSON      `json:"inputData" gorm:"type:jsonb;not null;column:input_data"`
	OutputData      datatypes.JSON      `json:"outputData,omitempty" gorm:"type:jsonb;column:output_data"`
	StepsCompleted  int                 `json:"stepsCompleted" gorm:"default:0;column:steps_completed"`
	TotalSteps      *int                `json:"totalSteps,omitempty" gorm:"column:total_steps"`
	AgentKeyID      UUID                `json:"agentKeyId" gorm:"column:agent_key_id"`
	ErrorMessage    *string             `json:"errorMessage,omitempty" gorm:"column:error_message"`

	// Relationships
	AgentAPIKey *AgentAPIKey `json:"agentApiKey,omitempty" gorm:"foreignKey:AgentKeyID;references:ID"`
}

// TableName returns the table name for AgentWorkflow
func (AgentWorkflow) TableName() string {
	return "agent_workflows"
}

// BeforeCreate sets the ID before creating
func (a *AgentWorkflow) BeforeCreate(tx *gorm.DB) error {
	if a.ID.IsZero() {
		a.ID = NewUUID()
	}
	return nil
}

// IsCompleted checks if the workflow is completed
func (a *AgentWorkflow) IsCompleted() bool {
	return a.Status == AgentWorkflowStatusCompleted
}

// IsFailed checks if the workflow has failed
func (a *AgentWorkflow) IsFailed() bool {
	return a.Status == AgentWorkflowStatusFailed
}

// IsInProgress checks if the workflow is currently running
func (a *AgentWorkflow) IsInProgress() bool {
	return a.Status == AgentWorkflowStatusInProgress
}

// GetCompletionPercentage returns the completion percentage (0-100)
func (a *AgentWorkflow) GetCompletionPercentage() float64 {
	if a.TotalSteps == nil || *a.TotalSteps == 0 {
		return 0
	}
	return (float64(a.StepsCompleted) / float64(*a.TotalSteps)) * 100
}