package agentapikey

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
	db *gorm.DB
}

// New creates a new agent API key store
func New(db *gorm.DB) IStore {
	return &store{
		db: db,
	}
}

// Create creates a new agent API key
func (s *store) Create(ctx context.Context, key *model.AgentAPIKey) (*model.AgentAPIKey, error) {
	if err := s.db.WithContext(ctx).Create(key).Error; err != nil {
		return nil, err
	}
	return key, nil
}

// GetByAPIKey retrieves an agent API key by its API key value
func (s *store) GetByAPIKey(ctx context.Context, apiKey string) (*model.AgentAPIKey, error) {
	var key model.AgentAPIKey
	if err := s.db.WithContext(ctx).
		Where("api_key = ? AND is_active = ? AND deleted_at IS NULL", apiKey, true).
		First(&key).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

// GetByID retrieves an agent API key by its ID
func (s *store) GetByID(ctx context.Context, id model.UUID) (*model.AgentAPIKey, error) {
	var key model.AgentAPIKey
	if err := s.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&key).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

// List retrieves all active agent API keys
func (s *store) List(ctx context.Context) ([]*model.AgentAPIKey, error) {
	var keys []*model.AgentAPIKey
	if err := s.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

// Update updates an agent API key
func (s *store) Update(ctx context.Context, key *model.AgentAPIKey) (*model.AgentAPIKey, error) {
	if err := s.db.WithContext(ctx).
		Model(key).
		Where("id = ? AND deleted_at IS NULL", key.ID).
		Updates(key).Error; err != nil {
		return nil, err
	}
	return key, nil
}

// Delete soft deletes an agent API key
func (s *store) Delete(ctx context.Context, id model.UUID) error {
	now := time.Now()
	return s.db.WithContext(ctx).
		Model(&model.AgentAPIKey{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now).Error
}

// Activate activates an agent API key
func (s *store) Activate(ctx context.Context, id model.UUID) error {
	return s.db.WithContext(ctx).
		Model(&model.AgentAPIKey{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("is_active", true).Error
}

// Deactivate deactivates an agent API key
func (s *store) Deactivate(ctx context.Context, id model.UUID) error {
	return s.db.WithContext(ctx).
		Model(&model.AgentAPIKey{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("is_active", false).Error
}