package agentactionlog

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
	db *gorm.DB
}

// New creates a new agent action log store
func New(db *gorm.DB) IStore {
	return &store{
		db: db,
	}
}

// Create creates a new agent action log
func (s *store) Create(ctx context.Context, log *model.AgentActionLog) (*model.AgentActionLog, error) {
	if err := s.db.WithContext(ctx).Create(log).Error; err != nil {
		return nil, err
	}
	return log, nil
}

// GetByID retrieves an agent action log by its ID
func (s *store) GetByID(ctx context.Context, id model.UUID) (*model.AgentActionLog, error) {
	var log model.AgentActionLog
	if err := s.db.WithContext(ctx).
		Preload("AgentAPIKey").
		Where("id = ?", id).
		First(&log).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

// ListByAgentKey retrieves agent action logs by agent key ID
func (s *store) ListByAgentKey(ctx context.Context, agentKeyID model.UUID, limit int) ([]*model.AgentActionLog, error) {
	var logs []*model.AgentActionLog
	query := s.db.WithContext(ctx).
		Where("agent_key_id = ?", agentKeyID).
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// ListByToolName retrieves agent action logs by tool name
func (s *store) ListByToolName(ctx context.Context, toolName string, limit int) ([]*model.AgentActionLog, error) {
	var logs []*model.AgentActionLog
	query := s.db.WithContext(ctx).
		Where("tool_name = ?", toolName).
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// ListByDateRange retrieves agent action logs within a date range
func (s *store) ListByDateRange(ctx context.Context, start, end time.Time, limit int) ([]*model.AgentActionLog, error) {
	var logs []*model.AgentActionLog
	query := s.db.WithContext(ctx).
		Where("created_at BETWEEN ? AND ?", start, end).
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// GetStatsByAgentKey retrieves statistics for actions by a specific agent key
func (s *store) GetStatsByAgentKey(ctx context.Context, agentKeyID model.UUID) (*ActionStats, error) {
	var stats ActionStats
	
	// Get total actions
	if err := s.db.WithContext(ctx).
		Model(&model.AgentActionLog{}).
		Where("agent_key_id = ?", agentKeyID).
		Count(&stats.TotalActions).Error; err != nil {
		return nil, err
	}
	
	// Get successful actions
	if err := s.db.WithContext(ctx).
		Model(&model.AgentActionLog{}).
		Where("agent_key_id = ? AND status = ?", agentKeyID, model.AgentActionLogStatusSuccess).
		Count(&stats.SuccessfulActions).Error; err != nil {
		return nil, err
	}
	
	// Get failed actions
	if err := s.db.WithContext(ctx).
		Model(&model.AgentActionLog{}).
		Where("agent_key_id = ? AND status = ?", agentKeyID, model.AgentActionLogStatusError).
		Count(&stats.FailedActions).Error; err != nil {
		return nil, err
	}
	
	// Get average duration
	var avgDuration *float64
	if err := s.db.WithContext(ctx).
		Model(&model.AgentActionLog{}).
		Where("agent_key_id = ? AND duration_ms IS NOT NULL", agentKeyID).
		Select("AVG(duration_ms)").
		Scan(&avgDuration).Error; err != nil {
		return nil, err
	}
	
	if avgDuration != nil {
		stats.AverageDurationMs = *avgDuration
	}
	
	// Calculate success rate
	if stats.TotalActions > 0 {
		stats.SuccessRate = float64(stats.SuccessfulActions) / float64(stats.TotalActions) * 100
	}
	
	return &stats, nil
}

// GetStatsByToolName retrieves statistics for actions by a specific tool
func (s *store) GetStatsByToolName(ctx context.Context, toolName string) (*ActionStats, error) {
	var stats ActionStats
	
	// Get total actions
	if err := s.db.WithContext(ctx).
		Model(&model.AgentActionLog{}).
		Where("tool_name = ?", toolName).
		Count(&stats.TotalActions).Error; err != nil {
		return nil, err
	}
	
	// Get successful actions
	if err := s.db.WithContext(ctx).
		Model(&model.AgentActionLog{}).
		Where("tool_name = ? AND status = ?", toolName, model.AgentActionLogStatusSuccess).
		Count(&stats.SuccessfulActions).Error; err != nil {
		return nil, err
	}
	
	// Get failed actions
	if err := s.db.WithContext(ctx).
		Model(&model.AgentActionLog{}).
		Where("tool_name = ? AND status = ?", toolName, model.AgentActionLogStatusError).
		Count(&stats.FailedActions).Error; err != nil {
		return nil, err
	}
	
	// Get average duration
	var avgDuration *float64
	if err := s.db.WithContext(ctx).
		Model(&model.AgentActionLog{}).
		Where("tool_name = ? AND duration_ms IS NOT NULL", toolName).
		Select("AVG(duration_ms)").
		Scan(&avgDuration).Error; err != nil {
		return nil, err
	}
	
	if avgDuration != nil {
		stats.AverageDurationMs = *avgDuration
	}
	
	// Calculate success rate
	if stats.TotalActions > 0 {
		stats.SuccessRate = float64(stats.SuccessfulActions) / float64(stats.TotalActions) * 100
	}
	
	return &stats, nil
}