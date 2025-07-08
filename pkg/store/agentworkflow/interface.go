package agentworkflow

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, workflow *model.AgentWorkflow) error
	GetByID(db *gorm.DB, id model.UUID) (*model.AgentWorkflow, error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, updates map[string]interface{}) error
	List(db *gorm.DB) ([]*model.AgentWorkflow, error)
	ListByAgentKeyID(db *gorm.DB, agentKeyID model.UUID) ([]*model.AgentWorkflow, error)
	ListByStatus(db *gorm.DB, status model.AgentWorkflowStatus) ([]*model.AgentWorkflow, error)
	Delete(db *gorm.DB, id model.UUID) error
}