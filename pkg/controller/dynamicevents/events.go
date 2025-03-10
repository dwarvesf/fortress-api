package dynamicevents

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

func (c *controller) CreateEvents(ctx context.Context, input model.DynamicEvent) error {
	// Parse the new event data
	var newEvent map[string]interface{}
	if err := json.Unmarshal(input.Data, &newEvent); err != nil {
		return fmt.Errorf("failed to parse event data: %w", err)
	}

	// Add event type to the event data
	newEvent["event_type"] = input.EventType

	// Construct the file path
	jsonPath := fmt.Sprintf("dynamic_events/%s_events.json", input.EventType)

	// Use the landing zone service to create/update events
	if err := c.service.LandingZone.CreateOrUpdateEvents(ctx, jsonPath, newEvent); err != nil {
		return fmt.Errorf("failed to create/update events: %w", err)
	}

	return nil
}
