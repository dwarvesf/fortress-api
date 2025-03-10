package landingzone

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

const (
	landingZoneBucketName = "df-landing-zone"
)

type service struct {
	gcs *CloudStorage
}

// New function return Google service
func New(GCSCredentials string) (IService, error) {
	decoded, err := base64.StdEncoding.DecodeString(GCSCredentials)
	if err != nil {
		return nil, fmt.Errorf("failed to decode gcs credentials: %v", err)
	}

	client, err := storage.NewClient(context.Background(), option.WithCredentialsJSON(decoded))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	return &service{
		gcs: &CloudStorage{
			client: client,
		},
	}, nil
}

func (g *service) UploadContentGCS(file io.Reader, filePath string) error {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer.
	wc := g.gcs.client.Bucket(g.gcs.bucketName).Object(filePath).NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
}

func (g *service) CreateOrUpdateEvents(ctx context.Context, filePath string, newEvent map[string]interface{}) error {
	bucket := g.gcs.client.Bucket(landingZoneBucketName)
	obj := bucket.Object(filePath)

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
