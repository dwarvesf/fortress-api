package agentapikey

import (
	"context"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// IStore is the interface for agent API key store
type IStore interface {
	Create(ctx context.Context, key *model.AgentAPIKey) (*model.AgentAPIKey, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*model.AgentAPIKey, error)
	GetByID(ctx context.Context, id model.UUID) (*model.AgentAPIKey, error)
	List(ctx context.Context) ([]*model.AgentAPIKey, error)
	Update(ctx context.Context, key *model.AgentAPIKey) (*model.AgentAPIKey, error)
	Delete(ctx context.Context, id model.UUID) error
	Activate(ctx context.Context, id model.UUID) error
	Deactivate(ctx context.Context, id model.UUID) error
}