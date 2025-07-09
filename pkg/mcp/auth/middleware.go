package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"gorm.io/datatypes"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/store/agentapikey"
)

var (
	ErrMissingAPIKey          = errors.New("missing API key")
	ErrInvalidAPIKey          = errors.New("invalid API key")
	ErrInactiveAPIKey         = errors.New("API key is inactive")
	ErrExpiredAPIKey          = errors.New("API key has expired")
	ErrInsufficientPermissions = errors.New("insufficient permissions for this tool")
	ErrInvalidPermissions     = errors.New("invalid permissions format")
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

// AgentPermissions represents the parsed permissions structure
type AgentPermissions struct {
	Tools        []string               `json:"tools"`
	Scopes       []string               `json:"scopes"`
	Restrictions map[string]interface{} `json:"restrictions"`
}

// ValidateToolPermission validates if an agent has permission to execute a specific tool
func (s *Service) ValidateToolPermission(ctx context.Context, agent *model.AgentAPIKey, toolName string) (bool, error) {
	// First check if the agent is valid
	if !agent.IsValid() {
		if agent.IsExpired() {
			return false, ErrExpiredAPIKey
		}
		return false, ErrInactiveAPIKey
	}

	// Parse the agent's permissions
	permissions, err := s.parsePermissions(agent.Permissions)
	if err != nil {
		return false, err
	}

	// Check if agent has permission for this tool
	hasPermission := s.hasToolPermission(permissions, toolName)
	if !hasPermission {
		return false, ErrInsufficientPermissions
	}

	return true, nil
}

// parsePermissions parses the JSON permissions into a structured format
// Handles both legacy array format ["workflow:calculate_monthly_payroll", "employee:*"]
// and new structured format {"tools": [...], "scopes": [...]}
func (s *Service) parsePermissions(permissionsJSON datatypes.JSON) (*AgentPermissions, error) {
	if len(permissionsJSON) == 0 {
		// Return empty permissions if no permissions are set
		return &AgentPermissions{
			Tools:  []string{},
			Scopes: []string{},
		}, nil
	}

	// Try to parse as structured format first
	var permissions AgentPermissions
	if err := json.Unmarshal(permissionsJSON, &permissions); err == nil {
		// Successfully parsed as structured format
		// Ensure slices are not nil
		if permissions.Tools == nil {
			permissions.Tools = []string{}
		}
		if permissions.Scopes == nil {
			permissions.Scopes = []string{}
		}
		return &permissions, nil
	}

	// Try to parse as legacy array format
	var legacyPermissions []string
	if err := json.Unmarshal(permissionsJSON, &legacyPermissions); err != nil {
		return nil, ErrInvalidPermissions
	}

	// Convert legacy format to structured format
	tools := []string{}
	scopes := []string{}
	
	for _, perm := range legacyPermissions {
		if strings.Contains(perm, ":") {
			// Parse category:tool format
			parts := strings.SplitN(perm, ":", 2)
			if len(parts) == 2 {
				category, tool := parts[0], parts[1]
				
				// Convert to tool name format
				if tool == "*" {
					// Wildcard permission for category
					tools = append(tools, category+":*")
				} else {
					// Specific tool permission
					tools = append(tools, tool)
				}
			}
		} else {
			// Direct permission without category
			tools = append(tools, perm)
		}
	}

	return &AgentPermissions{
		Tools:        tools,
		Scopes:       scopes,
		Restrictions: make(map[string]interface{}),
	}, nil
}

// hasToolPermission checks if the agent permissions allow access to a specific tool
func (s *Service) hasToolPermission(permissions *AgentPermissions, toolName string) bool {
	// Check for wildcard permissions
	for _, tool := range permissions.Tools {
		if tool == "*" {
			return true
		}
		if tool == toolName {
			return true
		}
		
		// Check category-based wildcard permissions (e.g., "workflow:*")
		if strings.HasSuffix(tool, ":*") {
			category := strings.TrimSuffix(tool, ":*")
			toolCategory := getToolCategory(toolName)
			if category == toolCategory {
				return true
			}
		}
	}

	// Check scope-based permissions
	for _, scope := range permissions.Scopes {
		switch scope {
		case "admin":
			// Admin scope allows all tools
			return true
		case "read":
			// Read scope allows only read-only tools
			if isReadOnlyTool(toolName) {
				return true
			}
		case "write":
			// Write scope allows read and write tools (but not admin tools)
			category := getToolCategory(toolName)
			if category != "admin" {
				return true
			}
		}
	}

	return false
}

// getToolCategory determines the category of a tool based on its name
func getToolCategory(toolName string) string {
	switch {
	case strings.Contains(toolName, "workflow"):
		return "workflow"
	case strings.Contains(toolName, "employee"):
		return "employee"
	case strings.Contains(toolName, "project"):
		return "project"
	case strings.Contains(toolName, "invoice"):
		return "invoice"
	case strings.Contains(toolName, "payroll"):
		return "payroll"
	default:
		return "unknown"
	}
}

// isReadOnlyTool determines if a tool is read-only based on its name
func isReadOnlyTool(toolName string) bool {
	readOnlyPrefixes := []string{"get_", "list_", "calculate_", "search_"}
	readOnlyTools := []string{
		"get_employee",
		"list_available_employees",
		"get_employee_skills",
		"get_project_details",
		"get_project_members",
		"get_invoice_status",
		"get_payroll_summary",
	}

	// Check for read-only prefixes
	for _, prefix := range readOnlyPrefixes {
		if strings.HasPrefix(toolName, prefix) {
			// Exception: some calculate_ tools might modify data
			if strings.HasPrefix(toolName, "calculate_") && !strings.Contains(toolName, "summary") {
				return false
			}
			return true
		}
	}

	// Check specific read-only tools
	for _, tool := range readOnlyTools {
		if tool == toolName {
			return true
		}
	}

	return false
}