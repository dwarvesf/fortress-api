package request

import (
	"encoding/json"
)

// DynamicEventRequest represents the incoming request structure
type DynamicEventRequest struct {
	Type string          `json:"type" binding:"required"`
	Data json.RawMessage `json:"data" binding:"required"`
}

// DynamicEventData represents the expected input data structure
type DynamicEventData struct {
	Data json.RawMessage `json:"data" binding:"required"`
}
