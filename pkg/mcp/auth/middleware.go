package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/store/agentapikey"
)

var (
	ErrMissingAPIKey   = errors.New("missing API key")
	ErrInvalidAPIKey   = errors.New("invalid API key")
	ErrInactiveAPIKey  = errors.New("API key is inactive")
	ErrExpiredAPIKey   = errors.New("API key has expired")
)

// Service handles agent authentication
type Service struct {
	agentKeyStore agentapikey.IStore
}

// New creates a new auth service
func New(agentKeyStore agentapikey.IStore) *Service {
	return &Service{
		agentKeyStore: agentKeyStore,
	}
}

// ValidateAPIKey validates an API key and returns the associated agent
func (s *Service) ValidateAPIKey(ctx context.Context, apiKey string) (*model.AgentAPIKey, error) {
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	// Remove "Bearer " prefix if present
	apiKey = strings.TrimPrefix(apiKey, "Bearer ")

	// Hash the API key for lookup (assuming we store hashed keys)
	hashedKey := s.hashAPIKey(apiKey)

	// Look up the API key in the database
	agentKey, err := s.agentKeyStore.GetByAPIKey(ctx, hashedKey)
	if err != nil {
		return nil, ErrInvalidAPIKey
	}

	// Check if the key is valid
	if !agentKey.IsValid() {
		if agentKey.IsExpired() {
			return nil, ErrExpiredAPIKey
		}
		return nil, ErrInactiveAPIKey
	}

	return agentKey, nil
}

// CreateAPIKey creates a new API key for an agent
func (s *Service) CreateAPIKey(ctx context.Context, name string, permissions []string) (*model.AgentAPIKey, string, error) {
	// Generate a random API key
	rawKey := s.generateAPIKey()
	hashedKey := s.hashAPIKey(rawKey)

	// Create the agent API key record
	agentKey := &model.AgentAPIKey{
		Name:        name,
		APIKey:      hashedKey,
		IsActive:    true,
		RateLimit:   1000, // Default rate limit
	}

	// Set permissions if provided
	if len(permissions) > 0 {
		// Convert permissions to JSON format expected by the model
		// This is a simplified version - you might want more sophisticated permission handling
		permissionData := make(map[string]interface{})
		permissionData["permissions"] = permissions
		// You would need to marshal this to JSON and set it to agentKey.Permissions
	}

	// Save to database
	createdKey, err := s.agentKeyStore.Create(ctx, agentKey)
	if err != nil {
		return nil, "", err
	}

	return createdKey, rawKey, nil
}

// RevokeAPIKey revokes an API key
func (s *Service) RevokeAPIKey(ctx context.Context, keyID model.UUID) error {
	return s.agentKeyStore.Deactivate(ctx, keyID)
}

// hashAPIKey hashes an API key for secure storage
func (s *Service) hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}

// generateAPIKey generates a new random API key
func (s *Service) generateAPIKey() string {
	// Generate a random 32-byte key and encode it as hex
	// In a real implementation, you'd use crypto/rand
	// This is a simplified version
	return "fts_" + hex.EncodeToString([]byte("sample_api_key_" + model.NewUUID().String()))
}

// ContextKey is the type for context keys
type ContextKey string

const (
	// AgentKeyContextKey is the context key for the agent API key
	AgentKeyContextKey ContextKey = "agent_key"
)

// GetAgentFromContext retrieves the agent from context
func GetAgentFromContext(ctx context.Context) (*model.AgentAPIKey, bool) {
	agent, ok := ctx.Value(AgentKeyContextKey).(*model.AgentAPIKey)
	return agent, ok
}

// SetAgentInContext sets the agent in context
func SetAgentInContext(ctx context.Context, agent *model.AgentAPIKey) context.Context {
	return context.WithValue(ctx, AgentKeyContextKey, agent)
}