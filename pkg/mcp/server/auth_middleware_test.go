package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/dwarvesf/fortress-api/pkg/mcp/auth"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

// MockAuthService is a mock for the auth service
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) ValidateAPIKey(ctx context.Context, apiKey string) (*model.AgentAPIKey, error) {
	args := m.Called(ctx, apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AgentAPIKey), args.Error(1)
}

func (m *MockAuthService) CreateAPIKey(ctx context.Context, name string, permissions []string) (*model.AgentAPIKey, string, error) {
	args := m.Called(ctx, name, permissions)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*model.AgentAPIKey), args.String(1), args.Error(2)
}

func (m *MockAuthService) RevokeAPIKey(ctx context.Context, keyID model.UUID) error {
	args := m.Called(ctx, keyID)
	return args.Error(0)
}

func TestHTTPAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		mockSetup      func(*MockAuthService)
		expectedStatus int
		expectedAgent  bool
		expectedError  string
	}{
		{
			name:       "Valid Bearer Token",
			authHeader: "Bearer valid_token_123",
			mockSetup: func(m *MockAuthService) {
				agent := &model.AgentAPIKey{
					BaseModel: model.BaseModel{ID: model.MustGetUUIDFromString("550e8400-e29b-41d4-a716-446655440000")},
					Name:      "test-agent",
					IsActive:  true,
				}
				m.On("ValidateAPIKey", mock.Anything, "valid_token_123").Return(agent, nil)
			},
			expectedStatus: http.StatusOK,
			expectedAgent:  true,
		},
		{
			name:       "Invalid Bearer Token",
			authHeader: "Bearer invalid_token",
			mockSetup: func(m *MockAuthService) {
				m.On("ValidateAPIKey", mock.Anything, "invalid_token").Return(nil, auth.ErrInvalidAPIKey)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedAgent:  false,
			expectedError:  "invalid API key",
		},
		{
			name:       "Missing Authorization Header",
			authHeader: "",
			mockSetup: func(m *MockAuthService) {
				// No mock setup needed as validation shouldn't be called
			},
			expectedStatus: http.StatusUnauthorized,
			expectedAgent:  false,
			expectedError:  "missing authorization header",
		},
		{
			name:       "Invalid Authorization Header Format",
			authHeader: "InvalidFormat token123",
			mockSetup: func(m *MockAuthService) {
				// No mock setup needed as validation shouldn't be called
			},
			expectedStatus: http.StatusUnauthorized,
			expectedAgent:  false,
			expectedError:  "invalid authorization header format",
		},
		{
			name:       "Expired API Key",
			authHeader: "Bearer expired_token",
			mockSetup: func(m *MockAuthService) {
				m.On("ValidateAPIKey", mock.Anything, "expired_token").Return(nil, auth.ErrExpiredAPIKey)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedAgent:  false,
			expectedError:  "API key has expired",
		},
		{
			name:       "Inactive API Key",
			authHeader: "Bearer inactive_token",
			mockSetup: func(m *MockAuthService) {
				m.On("ValidateAPIKey", mock.Anything, "inactive_token").Return(nil, auth.ErrInactiveAPIKey)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedAgent:  false,
			expectedError:  "API key is inactive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock auth service
			mockAuthService := &MockAuthService{}
			tt.mockSetup(mockAuthService)

			// Create middleware
			middleware := HTTPAuthMiddleware(mockAuthService)

			// Create test handler that checks for agent in context
			var contextAgent *model.AgentAPIKey
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				agent, exists := auth.GetAgentFromContext(r.Context())
				if exists {
					contextAgent = agent
				}
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write([]byte("OK")); err != nil {
					t.Errorf("Failed to write response: %v", err)
				}
			}))

			// Create test request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute middleware
			handler.ServeHTTP(rr, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedAgent {
				assert.NotNil(t, contextAgent, "Expected agent to be set in context")
				assert.Equal(t, "test-agent", contextAgent.Name)
			} else {
				assert.Nil(t, contextAgent, "Expected no agent in context")
			}

			if tt.expectedError != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedError)
			}

			// Verify mock expectations
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name          string
		authHeader    string
		expectedToken string
		expectedError error
	}{
		{
			name:          "Valid Bearer Token",
			authHeader:    "Bearer abc123",
			expectedToken: "abc123",
			expectedError: nil,
		},
		{
			name:          "Valid Bearer Token with Extra Spaces",
			authHeader:    "Bearer   token_with_spaces   ",
			expectedToken: "token_with_spaces",
			expectedError: nil,
		},
		{
			name:          "Missing Bearer Prefix",
			authHeader:    "token123",
			expectedToken: "",
			expectedError: ErrInvalidAuthHeader,
		},
		{
			name:          "Empty Authorization Header",
			authHeader:    "",
			expectedToken: "",
			expectedError: ErrMissingAuthHeader,
		},
		{
			name:          "Bearer with No Token",
			authHeader:    "Bearer",
			expectedToken: "",
			expectedError: ErrInvalidAuthHeader,
		},
		{
			name:          "Bearer with Empty Token",
			authHeader:    "Bearer ",
			expectedToken: "",
			expectedError: ErrInvalidAuthHeader,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := ExtractBearerToken(tt.authHeader)
			
			assert.Equal(t, tt.expectedToken, token)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestSetAgentInHTTPContext(t *testing.T) {
	// Create test agent
	agent := &model.AgentAPIKey{
		BaseModel: model.BaseModel{ID: model.MustGetUUIDFromString("550e8400-e29b-41d4-a716-446655440000")},
		Name:      "test-agent",
		IsActive:  true,
	}

	// Create context and set agent
	ctx := context.Background()
	ctxWithAgent := auth.SetAgentInContext(ctx, agent)

	// Retrieve agent from context
	retrievedAgent, exists := auth.GetAgentFromContext(ctxWithAgent)

	// Assertions
	assert.True(t, exists, "Agent should exist in context")
	assert.Equal(t, agent.ID, retrievedAgent.ID)
	assert.Equal(t, agent.Name, retrievedAgent.Name)
	assert.Equal(t, agent.IsActive, retrievedAgent.IsActive)
}

func TestAuthMiddlewareIntegration(t *testing.T) {
	// Create a mock auth service
	mockAuthService := &MockAuthService{}
	
	// Setup valid agent
	validAgent := &model.AgentAPIKey{
		BaseModel: model.BaseModel{ID: model.MustGetUUIDFromString("550e8400-e29b-41d4-a716-446655440000")},
		Name:      "integration-test-agent",
		IsActive:  true,
	}
	
	mockAuthService.On("ValidateAPIKey", mock.Anything, "valid_integration_token").Return(validAgent, nil)

	// Create middleware
	middleware := HTTPAuthMiddleware(mockAuthService)

	// Create a handler that requires authentication
	authenticatedHandler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agent, exists := auth.GetAgentFromContext(r.Context())
		if !exists {
			http.Error(w, "No agent in context", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"agent_id":"` + agent.ID.String() + `","agent_name":"` + agent.Name + `"}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))

	// Test successful authentication
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid_integration_token")
	rr := httptest.NewRecorder()

	authenticatedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "integration-test-agent")
	assert.Contains(t, rr.Body.String(), validAgent.ID.String())

	// Verify mock expectations
	mockAuthService.AssertExpectations(t)
}