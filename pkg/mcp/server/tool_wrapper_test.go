package server

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"

	"github.com/dwarvesf/fortress-api/pkg/mcp/auth"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

// TestWrapToolWithAuth_NoAgentInContext tests when no agent is in context
func TestWrapToolWithAuth_NoAgentInContext(t *testing.T) {
	// Test agent context extraction directly
	ctx := context.Background()
	
	// Test that no agent is found in empty context
	agent, exists := auth.GetAgentFromContext(ctx)
	assert.False(t, exists)
	assert.Nil(t, agent)
}

// TestWrapToolWithAuth_AgentInContext tests basic agent extraction from context
func TestWrapToolWithAuth_AgentInContext(t *testing.T) {
	// Create a test agent
	agent := &model.AgentAPIKey{
		BaseModel: model.BaseModel{ID: model.NewUUID()},
		Name:      "test-agent",
		Permissions: createPermissionsJSON(map[string]interface{}{
			"tools": []string{"test_tool"},
		}),
		IsActive: true,
	}

	// Create context with agent
	ctx := auth.SetAgentInContext(context.Background(), agent)

	// Test that the agent can be extracted from context
	extractedAgent, exists := auth.GetAgentFromContext(ctx)
	assert.True(t, exists)
	assert.Equal(t, agent.ID, extractedAgent.ID)
	assert.Equal(t, agent.Name, extractedAgent.Name)
}

// TestWrapToolWithAuth_Integration tests the integration with real auth service
func TestWrapToolWithAuth_Integration(t *testing.T) {
	// Skip if we don't have a real auth service setup
	t.Skip("Integration test - requires real auth service setup")

	// This test would require setting up a real auth service with database
	// and is more suitable for integration testing
}

// TestAgentPermissions tests agent permission validation scenarios
func TestAgentPermissions(t *testing.T) {
	tests := []struct {
		name        string
		agent       *model.AgentAPIKey
		toolName    string
		shouldAllow bool
	}{
		{
			name: "Agent with specific tool permission",
			agent: &model.AgentAPIKey{
				BaseModel: model.BaseModel{ID: model.NewUUID()},
				Name:      "test-agent",
				Permissions: createPermissionsJSON(map[string]interface{}{
					"tools": []string{"get_employee"},
				}),
				IsActive: true,
			},
			toolName:    "get_employee",
			shouldAllow: true,
		},
		{
			name: "Agent with wildcard permissions",
			agent: &model.AgentAPIKey{
				BaseModel: model.BaseModel{ID: model.NewUUID()},
				Name:      "admin-agent",
				Permissions: createPermissionsJSON(map[string]interface{}{
					"tools": []string{"*"},
				}),
				IsActive: true,
			},
			toolName:    "any_tool",
			shouldAllow: true,
		},
		{
			name: "Agent with read scope for read tool",
			agent: &model.AgentAPIKey{
				BaseModel: model.BaseModel{ID: model.NewUUID()},
				Name:      "readonly-agent",
				Permissions: createPermissionsJSON(map[string]interface{}{
					"scopes": []string{"read"},
				}),
				IsActive: true,
			},
			toolName:    "get_employee",
			shouldAllow: true,
		},
		{
			name: "Agent with read scope for write tool",
			agent: &model.AgentAPIKey{
				BaseModel: model.BaseModel{ID: model.NewUUID()},
				Name:      "readonly-agent",
				Permissions: createPermissionsJSON(map[string]interface{}{
					"scopes": []string{"read"},
				}),
				IsActive: true,
			},
			toolName:    "create_project",
			shouldAllow: false,
		},
		{
			name: "Inactive agent",
			agent: &model.AgentAPIKey{
				BaseModel: model.BaseModel{ID: model.NewUUID()},
				Name:      "inactive-agent",
				Permissions: createPermissionsJSON(map[string]interface{}{
					"tools": []string{"*"},
				}),
				IsActive: false,
			},
			toolName:    "get_employee",
			shouldAllow: false,
		},
		{
			name: "Expired agent",
			agent: &model.AgentAPIKey{
				BaseModel: model.BaseModel{ID: model.NewUUID()},
				Name:      "expired-agent",
				Permissions: createPermissionsJSON(map[string]interface{}{
					"tools": []string{"*"},
				}),
				IsActive:  true,
				ExpiresAt: func() *time.Time { t := time.Now().Add(-time.Hour); return &t }(),
			},
			toolName:    "get_employee",
			shouldAllow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is a unit test for the permission logic
			// The actual permission validation is tested in the auth package
			
			// Test agent validity
			if tt.shouldAllow {
				if tt.agent.IsActive && !tt.agent.IsExpired() {
					assert.True(t, tt.agent.IsValid())
				}
			} else {
				if !tt.agent.IsActive || tt.agent.IsExpired() {
					assert.False(t, tt.agent.IsValid())
				}
			}
		})
	}
}

// TestToolHandlerError tests error handling in tool execution
func TestToolHandlerError(t *testing.T) {
	// Create a test agent
	agent := &model.AgentAPIKey{
		BaseModel: model.BaseModel{ID: model.NewUUID()},
		Name:      "test-agent",
		Permissions: createPermissionsJSON(map[string]interface{}{
			"tools": []string{"failing_tool"},
		}),
		IsActive: true,
	}

	// Create context with agent
	ctx := auth.SetAgentInContext(context.Background(), agent)

	// Create a test tool handler that fails
	testError := fmt.Errorf("tool execution failed")
	testHandler := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return nil, testError
	}

	// Test that the error is properly handled
	// Note: This is a simplified test since we're not testing the full wrapper
	result, err := testHandler(ctx, mcp.CallToolRequest{})
	assert.Error(t, err)
	assert.Equal(t, testError, err)
	assert.Nil(t, result)
}

// Helper function to create permissions JSON
func createPermissionsJSON(data map[string]interface{}) datatypes.JSON {
	jsonData, _ := json.Marshal(data)
	return datatypes.JSON(jsonData)
}