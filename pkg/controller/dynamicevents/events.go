package dynamicevents

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/storage"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"google.golang.org/api/option"
)

func (c *controller) CreateEvents(ctx context.Context, input model.DynamicEvent) error {
	// Setup GCS client
	var client *storage.Client
	var err error

	if gcsCredentials := c.config.Google.GCSLandingZoneCredentials; gcsCredentials != "" {
		decodedCreds, decodeErr := base64.StdEncoding.DecodeString(gcsCredentials)
		if decodeErr != nil {
			decodedCreds = []byte(gcsCredentials)
		}

		var credJSON map[string]interface{}
		if err = json.Unmarshal(decodedCreds, &credJSON); err != nil {
			return fmt.Errorf("invalid GCP credentials JSON: %w", err)
		}

		client, err = storage.NewClient(ctx, option.WithCredentialsJSON(decodedCreds))
	} else {
		client, err = storage.NewClient(ctx)
	}

	if err != nil {
		return fmt.Errorf("failed to create GCP client: %w", err)
	}
	defer client.Close()
	bucket := client.Bucket("df-landing-zone")

	// First, try to read existing events file
	jsonPath := fmt.Sprintf("dynamic_events/%s_events.json", input.EventType)
	obj := bucket.Object(jsonPath)

	var events []map[string]interface{}

	// Try to read existing file
	reader, err := obj.NewReader(ctx)
	if err != nil {
		if err != storage.ErrObjectNotExist {
			return fmt.Errorf("failed to read events file: %w", err)
		}
		// File doesn't exist yet, start with empty array
		events = make([]map[string]interface{}, 0)
	} else {
		defer reader.Close()
		if err := json.NewDecoder(reader).Decode(&events); err != nil {
			return fmt.Errorf("failed to decode existing events: %w", err)
		}
	}

	// Parse the new event data
	var newEvent map[string]interface{}
	if err := json.Unmarshal(input.Data, &newEvent); err != nil {
		return fmt.Errorf("failed to parse event data: %w", err)
	}

	// Add timestamp to the event
	newEvent["timestamp"] = time.Now()

	// Append new event
	events = append(events, newEvent)

	// Write updated events back to GCS
	writer := obj.NewWriter(ctx)
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(events); err != nil {
		writer.Close()
		return fmt.Errorf("failed to write events: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return nil
}
