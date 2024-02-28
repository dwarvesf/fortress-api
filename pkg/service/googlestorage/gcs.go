package googlestorage

import (
	"cloud.google.com/go/storage"
)

type CloudStorage struct {
	client     *storage.Client
	projectID  string
	bucketName string
}
