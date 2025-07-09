package auth

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/datatypes"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// MockAgentAPIKeyStore is a mock for the agent API key store
type MockAgentAPIKeyStore struct {
	mock.Mock
}

func (m *MockAgentAPIKeyStore) Create(ctx context.Context, key *model.AgentAPIKey) (*model.AgentAPIKey, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AgentAPIKey), args.Error(1)
}

func (m *MockAgentAPIKeyStore) GetByAPIKey(ctx context.Context, apiKey string) (*model.AgentAPIKey, error) {
	args := m.Called(ctx, apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AgentAPIKey), args.Error(1)
}

func (m *MockAgentAPIKeyStore) GetByID(ctx context.Context, id model.UUID) (*model.AgentAPIKey, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AgentAPIKey), args.Error(1)
}

func (m *MockAgentAPIKeyStore) List(ctx context.Context) ([]*model.AgentAPIKey, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.AgentAPIKey), args.Error(1)
}

func (m *MockAgentAPIKeyStore) Update(ctx context.Context, key *model.AgentAPIKey) (*model.AgentAPIKey, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AgentAPIKey), args.Error(1)
}

func (m *MockAgentAPIKeyStore) Delete(ctx context.Context, id model.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAgentAPIKeyStore) Activate(ctx context.Context, id model.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAgentAPIKeyStore) Deactivate(ctx context.Context, id model.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestValidateToolPermission(t *testing.T) {
	tests := []struct {
		name            string
		agent           *model.AgentAPIKey
		toolName        string
		expectedAllowed bool
		expectedError   error
	}{
		{
			name: "Agent with specific tool permission",
			agent: &model.AgentAPIKey{
				BaseModel: model.BaseModel{ID: model.NewUUID()},
				Name:      "test-agent",
				Permissions: createPermissionsJSON(map[string]interface{}{
					"tools": []string{"get_employee", "create_project"},
					"scopes": []string{"read", "write"},
				}),
				IsActive: true,
			},
			toolName:        "get_employee",
			expectedAllowed: true,
			expectedError:   nil,
		},
		{
			name: "Agent without specific tool permission",
			agent: &model.AgentAPIKey{
				BaseModel: model.BaseModel{ID: model.NewUUID()},
				Name:      "test-agent",
				Permissions: createPermissionsJSON(map[string]interface{}{
					"tools": []string{"get_employee"},
					"scopes": []string{"read"},
				}),
				IsActive: true,
			},
			toolName:        "create_project",
			expectedAllowed: false,
			expectedError:   ErrInsufficientPermissions,
		},
		{
			name: "Agent with wildcard permissions",
			agent: &model.AgentAPIKey{
				BaseModel: model.BaseModel{ID: model.NewUUID()},
				Name:      "admin-agent",
				Permissions: createPermissionsJSON(map[string]interface{}{
					"tools": []string{"*"},
					"scopes": []string{"admin"},
				}),
				IsActive: true,
			},
			toolName:        "any_tool",
			expectedAllowed: true,
			expectedError:   nil,
		},
		{
			name: "Agent with scope-based permissions - read tools",
			agent: &model.AgentAPIKey{
				BaseModel: model.BaseModel{ID: model.NewUUID()},
				Name:      "readonly-agent",
				Permissions: createPermissionsJSON(map[string]interface{}{
					"scopes": []string{"read"},
				}),
				IsActive: true,
			},
			toolName:        "get_employee",
			expectedAllowed: true,
			expectedError:   nil,
		},
		{
			name: "Agent with scope-based permissions - write tools denied",
			agent: &model.AgentAPIKey{
				BaseModel: model.BaseModel{ID: model.NewUUID()},
				Name:      "readonly-agent",
				Permissions: createPermissionsJSON(map[string]interface{}{
					"scopes": []string{"read"},
				}),
				IsActive: true,
			},
			toolName:        "create_project",
			expectedAllowed: false,
			expectedError:   ErrInsufficientPermissions,
		},
		{
			name: "Agent with no permissions",
			agent: &model.AgentAPIKey{
				BaseModel:   model.BaseModel{ID: model.NewUUID()},
				Name:        "no-perms-agent",
				Permissions: createPermissionsJSON(map[string]interface{}{}),
				IsActive:    true,
			},
			toolName:        "get_employee",
			expectedAllowed: false,
			expectedError:   ErrInsufficientPermissions,
		},
		{
			name: "Agent with malformed permissions JSON",
			agent: &model.AgentAPIKey{
				BaseModel:   model.BaseModel{ID: model.NewUUID()},
				Name:        "malformed-agent",
				Permissions: datatypes.JSON([]byte(`{"invalid": json}`)),
				IsActive:    true,
			},
			toolName:        "get_employee",
			expectedAllowed: false,
			expectedError:   ErrInvalidPermissions,
		},
		{
			name: "Inactive agent",
			agent: &model.AgentAPIKey{
				BaseModel: model.BaseModel{ID: model.NewUUID()},
				Name:      "inactive-agent",
				Permissions: createPermissionsJSON(map[string]interface{}{
					"tools": []string{"*"},
					"scopes": []string{"admin"},
				}),
				IsActive: false,
			},
			toolName:        "get_employee",
			expectedAllowed: false,
			expectedError:   ErrInactiveAPIKey,
		},
		{
			name: "Expired agent",
			agent: &model.AgentAPIKey{
				BaseModel: model.BaseModel{ID: model.NewUUID()},
				Name:      "expired-agent",
				Permissions: createPermissionsJSON(map[string]interface{}{
					"tools": []string{"*"},
					"scopes": []string{"admin"},
				}),
				IsActive:  true,
				ExpiresAt: func() *time.Time { t := time.Now().Add(-time.Hour); return &t }(),
			},
			toolName:        "get_employee",
			expectedAllowed: false,
			expectedError:   ErrExpiredAPIKey,
		},
		{
			name: "Agent with legacy format - category wildcard permission",
			agent: &model.AgentAPIKey{
				BaseModel: model.BaseModel{ID: model.NewUUID()},
				Name:      "legacy-agent",
				Permissions: createPermissionsJSONArray([]string{"payroll:*", "employee:get_employee"}),
				IsActive: true,
			},
			toolName:        "calculate_monthly_payroll",
			expectedAllowed: true,
			expectedError:   nil,
		},
		{
			name: "Agent with legacy format - specific tool permission",
			agent: &model.AgentAPIKey{
				BaseModel: model.BaseModel{ID: model.NewUUID()},
				Name:      "legacy-agent",
				Permissions: createPermissionsJSONArray([]string{"workflow:calculate_monthly_payroll", "employee:*"}),
				IsActive: true,
			},
			toolName:        "calculate_monthly_payroll",
			expectedAllowed: true,
			expectedError:   nil,
		},
		{
			name: "Agent with legacy format - denied tool",
			agent: &model.AgentAPIKey{
				BaseModel: model.BaseModel{ID: model.NewUUID()},
				Name:      "legacy-agent",
				Permissions: createPermissionsJSONArray([]string{"employee:*"}),
				IsActive: true,
			},
			toolName:        "calculate_monthly_payroll",
			expectedAllowed: false,
			expectedError:   ErrInsufficientPermissions,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock store
			mockStore := &MockAgentAPIKeyStore{}
			authService := New(mockStore)

			// Execute validation
			allowed, err := authService.ValidateToolPermission(context.Background(), tt.agent, tt.toolName)

			// Assertions
			assert.Equal(t, tt.expectedAllowed, allowed)
			assert.Equal(t, tt.expectedError, err)

			// Verify mock expectations
			mockStore.AssertExpectations(t)
		})
	}
}

func TestParsePermissions(t *testing.T) {
	tests := []struct {
		name                string
		permissionsJSON     datatypes.JSON
		expectedPermissions *AgentPermissions
		expectedError       error
	}{
		{
			name: "Valid permissions with tools and scopes",
			permissionsJSON: createPermissionsJSON(map[string]interface{}{
				"tools": []string{"get_employee", "create_project"},
				"scopes": []string{"read", "write"},
				"restrictions": map[string]interface{}{
					"employee_access": "own_data_only",
				},
			}),
			expectedPermissions: &AgentPermissions{
				Tools:  []string{"get_employee", "create_project"},
				Scopes: []string{"read", "write"},
				Restrictions: map[string]interface{}{
					"employee_access": "own_data_only",
				},
			},
			expectedError: nil,
		},
		{
			name: "Valid permissions with wildcard tools",
			permissionsJSON: createPermissionsJSON(map[string]interface{}{
				"tools": []string{"*"},
				"scopes": []string{"admin"},
			}),
			expectedPermissions: &AgentPermissions{
				Tools:  []string{"*"},
				Scopes: []string{"admin"},
			},
			expectedError: nil,
		},
		{
			name: "Valid permissions with only scopes",
			permissionsJSON: createPermissionsJSON(map[string]interface{}{
				"scopes": []string{"read"},
			}),
			expectedPermissions: &AgentPermissions{
				Tools:  []string{},
				Scopes: []string{"read"},
			},
			expectedError: nil,
		},
		{
			name: "Legacy array format - category:tool",
			permissionsJSON: createPermissionsJSONArray([]string{"workflow:calculate_monthly_payroll", "employee:*"}),
			expectedPermissions: &AgentPermissions{
				Tools:  []string{"calculate_monthly_payroll", "employee:*"},
				Scopes: []string{},
			},
			expectedError: nil,
		},
		{
			name: "Legacy array format - direct tools",
			permissionsJSON: createPermissionsJSONArray([]string{"get_employee", "create_project"}),
			expectedPermissions: &AgentPermissions{
				Tools:  []string{"get_employee", "create_project"},
				Scopes: []string{},
			},
			expectedError: nil,
		},
		{
			name: "Legacy array format - mixed",
			permissionsJSON: createPermissionsJSONArray([]string{"workflow:*", "get_employee", "project:create_project"}),
			expectedPermissions: &AgentPermissions{
				Tools:  []string{"workflow:*", "get_employee", "create_project"},
				Scopes: []string{},
			},
			expectedError: nil,
		},
		{
			name: "Legacy array format - wildcard only",
			permissionsJSON: createPermissionsJSONArray([]string{"*"}),
			expectedPermissions: &AgentPermissions{
				Tools:  []string{"*"},
				Scopes: []string{},
			},
			expectedError: nil,
		},
		{
			name:                "Empty permissions",
			permissionsJSON:     createPermissionsJSON(map[string]interface{}{}),
			expectedPermissions: &AgentPermissions{Tools: []string{}, Scopes: []string{}},
			expectedError:       nil,
		},
		{
			name:                "Invalid JSON",
			permissionsJSON:     datatypes.JSON([]byte(`{"invalid": json}`)),
			expectedPermissions: nil,
			expectedError:       ErrInvalidPermissions,
		},
		{
			name:                "Nil permissions",
			permissionsJSON:     nil,
			expectedPermissions: &AgentPermissions{Tools: []string{}, Scopes: []string{}},
			expectedError:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock store
			mockStore := &MockAgentAPIKeyStore{}
			authService := New(mockStore)

			// Execute parsing
			permissions, err := authService.parsePermissions(tt.permissionsJSON)

			// Assertions
			assert.Equal(t, tt.expectedError, err)
			if tt.expectedError == nil {
				assert.Equal(t, tt.expectedPermissions.Tools, permissions.Tools)
				assert.Equal(t, tt.expectedPermissions.Scopes, permissions.Scopes)
				if tt.expectedPermissions.Restrictions != nil {
					assert.Equal(t, tt.expectedPermissions.Restrictions, permissions.Restrictions)
				}
			} else {
				assert.Nil(t, permissions)
			}

			// Verify mock expectations
			mockStore.AssertExpectations(t)
		})
	}
}

func TestGetToolCategory(t *testing.T) {
	tests := []struct {
		name             string
		toolName         string
		expectedCategory string
	}{
		{
			name:             "Employee tool",
			toolName:         "get_employee",
			expectedCategory: "employee",
		},
		{
			name:             "Project tool",
			toolName:         "create_project",
			expectedCategory: "project",
		},
		{
			name:             "Invoice tool",
			toolName:         "generate_invoice",
			expectedCategory: "invoice",
		},
		{
			name:             "Payroll tool",
			toolName:         "calculate_payroll",
			expectedCategory: "payroll",
		},
		{
			name:             "Payroll tool (monthly)",
			toolName:         "calculate_monthly_payroll",
			expectedCategory: "payroll",
		},
		{
			name:             "Unknown tool",
			toolName:         "unknown_tool",
			expectedCategory: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category := getToolCategory(tt.toolName)
			assert.Equal(t, tt.expectedCategory, category)
		})
	}
}

func TestIsReadOnlyTool(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		expected bool
	}{
		{
			name:     "Read tool - get_employee",
			toolName: "get_employee",
			expected: true,
		},
		{
			name:     "Read tool - list_available_employees",
			toolName: "list_available_employees",
			expected: true,
		},
		{
			name:     "Read tool - get_project_details",
			toolName: "get_project_details",
			expected: true,
		},
		{
			name:     "Write tool - create_project",
			toolName: "create_project",
			expected: false,
		},
		{
			name:     "Write tool - update_employee_status",
			toolName: "update_employee_status",
			expected: false,
		},
		{
			name:     "Write tool - generate_invoice",
			toolName: "generate_invoice",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isReadOnlyTool(tt.toolName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to create permissions JSON
func createPermissionsJSON(data map[string]interface{}) datatypes.JSON {
	jsonData, _ := json.Marshal(data)
	return datatypes.JSON(jsonData)
}

// Helper function to create permissions JSON array (legacy format)
func createPermissionsJSONArray(data []string) datatypes.JSON {
	jsonData, _ := json.Marshal(data)
	return datatypes.JSON(jsonData)
}