package agentactionlog

import (
	"context"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// IStore is the interface for agent action log store
type IStore interface {
	Create(ctx context.Context, log *model.AgentActionLog) (*model.AgentActionLog, error)
	GetByID(ctx context.Context, id model.UUID) (*model.AgentActionLog, error)
	ListByAgentKey(ctx context.Context, agentKeyID model.UUID, limit int) ([]*model.AgentActionLog, error)
	ListByToolName(ctx context.Context, toolName string, limit int) ([]*model.AgentActionLog, error)
	ListByDateRange(ctx context.Context, start, end time.Time, limit int) ([]*model.AgentActionLog, error)
	GetStatsByAgentKey(ctx context.Context, agentKeyID model.UUID) (*ActionStats, error)
	GetStatsByToolName(ctx context.Context, toolName string) (*ActionStats, error)
}

// ActionStats represents statistics for agent actions
type ActionStats struct {
	TotalActions     int64   `json:"totalActions"`
	SuccessfulActions int64   `json:"successfulActions"`
	FailedActions    int64   `json:"failedActions"`
	AverageDurationMs float64 `json:"averageDurationMs"`
	SuccessRate      float64 `json:"successRate"`
}