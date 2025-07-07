package view

import (
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// FormatJSONResponse formats any data structure into a JSON MCP tool result
// This utility function can be reused across all MCP tools for consistent JSON formatting
func FormatJSONResponse(data any) (*mcp.CallToolResult, error) {
	resultJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}

// FormatJSONResponseWithPrefix formats data with a text prefix for better context
func FormatJSONResponseWithPrefix(prefix string, data any) (*mcp.CallToolResult, error) {
	resultJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format result: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("%s:\n\n%s", prefix, string(resultJSON))), nil
}

// FormatErrorResponse creates a consistent error response for MCP tools
func FormatErrorResponse(operation string, err error) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError(fmt.Sprintf("%s failed: %v", operation, err)), nil
}
