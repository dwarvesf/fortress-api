package agentworkflow

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

// New returns a new agent workflow store
func New() IStore {
	return &store{}
}

// Create creates a new agent workflow
func (s *store) Create(db *gorm.DB, workflow *model.AgentWorkflow) error {
	return db.Create(workflow).Error
}

// GetByID retrieves an agent workflow by ID
func (s *store) GetByID(db *gorm.DB, id model.UUID) (*model.AgentWorkflow, error) {
	var workflow model.AgentWorkflow
	err := db.Where("id = ?", id).First(&workflow).Error
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

// UpdateSelectedFieldsByID updates selected fields of an agent workflow by ID
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updates map[string]interface{}) error {
	return db.Model(&model.AgentWorkflow{}).Where("id = ?", id).Updates(updates).Error
}

// List retrieves all agent workflows
func (s *store) List(db *gorm.DB) ([]*model.AgentWorkflow, error) {
	var workflows []*model.AgentWorkflow
	err := db.Find(&workflows).Error
	return workflows, err
}

// ListByAgentKeyID retrieves agent workflows by agent key ID
func (s *store) ListByAgentKeyID(db *gorm.DB, agentKeyID model.UUID) ([]*model.AgentWorkflow, error) {
	var workflows []*model.AgentWorkflow
	err := db.Where("agent_key_id = ?", agentKeyID).Find(&workflows).Error
	return workflows, err
}

// ListByStatus retrieves agent workflows by status
func (s *store) ListByStatus(db *gorm.DB, status model.AgentWorkflowStatus) ([]*model.AgentWorkflow, error) {
	var workflows []*model.AgentWorkflow
	err := db.Where("status = ?", status).Find(&workflows).Error
	return workflows, err
}

// Delete deletes an agent workflow by ID
func (s *store) Delete(db *gorm.DB, id model.UUID) error {
	return db.Where("id = ?", id).Delete(&model.AgentWorkflow{}).Error
}