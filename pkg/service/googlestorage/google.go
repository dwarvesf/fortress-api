package googlestorage

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type service struct {
	gcs *CloudStorage
}

// New function return Google service
func New(BucketName string, GCSProjectID string, GCSCredentials string) (IService, error) {
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
			client:     client,
			projectID:  GCSProjectID,
			bucketName: BucketName,
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
