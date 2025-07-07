package errors

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	ErrorTypeValidation     ErrorType = "validation_error"
	ErrorTypeNotFound       ErrorType = "not_found_error"
	ErrorTypeDatabase       ErrorType = "database_error"
	ErrorTypePermission     ErrorType = "permission_error"
	ErrorTypeRateLimit      ErrorType = "rate_limit_error"
	ErrorTypeInternal       ErrorType = "internal_error"
	ErrorTypeExternalAPI    ErrorType = "external_api_error"
	ErrorTypeConfiguration ErrorType = "configuration_error"
)

// MCPError represents a structured MCP error
type MCPError struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Code    string    `json:"code,omitempty"`
	Details string    `json:"details,omitempty"`
}

// Error implements the error interface
func (e *MCPError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Type, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// NewValidationError creates a validation error
func NewValidationError(message string, details ...string) *mcp.CallToolResult {
	err := &MCPError{
		Type:    ErrorTypeValidation,
		Message: message,
		Code:    "VALIDATION_FAILED",
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return mcp.NewToolResultError(err.Error())
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource, identifier string) *mcp.CallToolResult {
	err := &MCPError{
		Type:    ErrorTypeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Code:    "RESOURCE_NOT_FOUND",
		Details: fmt.Sprintf("Identifier: %s", identifier),
	}
	return mcp.NewToolResultError(err.Error())
}

// NewDatabaseError creates a database error
func NewDatabaseError(operation string, underlying error) *mcp.CallToolResult {
	err := &MCPError{
		Type:    ErrorTypeDatabase,
		Message: fmt.Sprintf("Database operation failed: %s", operation),
		Code:    "DATABASE_ERROR",
		Details: underlying.Error(),
	}
	return mcp.NewToolResultError(err.Error())
}

// NewPermissionError creates a permission error
func NewPermissionError(action, resource string) *mcp.CallToolResult {
	err := &MCPError{
		Type:    ErrorTypePermission,
		Message: fmt.Sprintf("Permission denied for %s on %s", action, resource),
		Code:    "PERMISSION_DENIED",
	}
	return mcp.NewToolResultError(err.Error())
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(limit string, retryAfter string) *mcp.CallToolResult {
	err := &MCPError{
		Type:    ErrorTypeRateLimit,
		Message: fmt.Sprintf("Rate limit exceeded: %s", limit),
		Code:    "RATE_LIMIT_EXCEEDED",
		Details: fmt.Sprintf("Retry after: %s", retryAfter),
	}
	return mcp.NewToolResultError(err.Error())
}

// NewInternalError creates an internal error
func NewInternalError(component string, underlying error) *mcp.CallToolResult {
	err := &MCPError{
		Type:    ErrorTypeInternal,
		Message: fmt.Sprintf("Internal error in %s", component),
		Code:    "INTERNAL_ERROR",
		Details: underlying.Error(),
	}
	return mcp.NewToolResultError(err.Error())
}

// NewExternalAPIError creates an external API error
func NewExternalAPIError(service string, underlying error) *mcp.CallToolResult {
	err := &MCPError{
		Type:    ErrorTypeExternalAPI,
		Message: fmt.Sprintf("External API error: %s", service),
		Code:    "EXTERNAL_API_ERROR",
		Details: underlying.Error(),
	}
	return mcp.NewToolResultError(err.Error())
}

// NewConfigurationError creates a configuration error
func NewConfigurationError(setting string, issue string) *mcp.CallToolResult {
	err := &MCPError{
		Type:    ErrorTypeConfiguration,
		Message: fmt.Sprintf("Configuration error: %s", setting),
		Code:    "CONFIGURATION_ERROR",
		Details: issue,
	}
	return mcp.NewToolResultError(err.Error())
}

// WrapError wraps a generic error with context
func WrapError(errType ErrorType, context string, underlying error) *mcp.CallToolResult {
	err := &MCPError{
		Type:    errType,
		Message: context,
		Details: underlying.Error(),
	}
	return mcp.NewToolResultError(err.Error())
}