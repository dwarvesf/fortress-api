package landingzone

import (
	"context"
	"io"
)

type IService interface {
	UploadContentGCS(file io.Reader, fileName string) error
	CreateOrUpdateEvents(ctx context.Context, filePath string, newEvent map[string]interface{}) error
}
