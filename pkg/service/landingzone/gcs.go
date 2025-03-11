package landingzone

import (
	"cloud.google.com/go/storage"
)

type CloudStorage struct {
	client     *storage.Client
	bucketName string
}
