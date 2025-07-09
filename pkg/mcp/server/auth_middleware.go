package server

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/dwarvesf/fortress-api/pkg/mcp/auth"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

var (
	ErrMissingAuthHeader = errors.New("missing authorization header")
	ErrInvalidAuthHeader = errors.New("invalid authorization header format")
)

// AuthService interface for dependency injection in tests
type AuthService interface {
	ValidateAPIKey(ctx context.Context, apiKey string) (*model.AgentAPIKey, error)
	CreateAPIKey(ctx context.Context, name string, permissions []string) (*model.AgentAPIKey, string, error)
	RevokeAPIKey(ctx context.Context, keyID model.UUID) error
}

// HTTPAuthMiddleware creates an HTTP middleware for agent authentication
func HTTPAuthMiddleware(authService AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract Bearer token from Authorization header
			authHeader := r.Header.Get("Authorization")
			token, err := ExtractBearerToken(authHeader)
			if err != nil {
				writeAuthError(w, err)
				return
			}

			// Validate the API key
			agent, err := authService.ValidateAPIKey(r.Context(), token)
			if err != nil {
				writeAuthError(w, err)
				return
			}

			// Set the agent in the request context
			ctx := auth.SetAgentInContext(r.Context(), agent)
			r = r.WithContext(ctx)

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// ExtractBearerToken extracts the Bearer token from the Authorization header
func ExtractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", ErrMissingAuthHeader
	}

	// Check for Bearer prefix
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", ErrInvalidAuthHeader
	}

	// Extract token (everything after "Bearer ")
	token := strings.TrimSpace(authHeader[len(bearerPrefix):])
	if token == "" {
		return "", ErrInvalidAuthHeader
	}

	return token, nil
}

// writeAuthError writes an authentication error response
func writeAuthError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	
	var message string
	switch {
	case errors.Is(err, ErrMissingAuthHeader):
		message = "missing authorization header"
	case errors.Is(err, ErrInvalidAuthHeader):
		message = "invalid authorization header format"
	case errors.Is(err, auth.ErrInvalidAPIKey):
		message = "invalid API key"
	case errors.Is(err, auth.ErrExpiredAPIKey):
		message = "API key has expired"
	case errors.Is(err, auth.ErrInactiveAPIKey):
		message = "API key is inactive"
	default:
		message = "authentication failed"
	}
	
	w.Write([]byte(`{"error":"` + message + `"}`))
}