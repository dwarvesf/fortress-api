package googlestorage

import (
	"io"
)

type IService interface {
	UploadContentGCS(file io.Reader, fileName string) error
}
