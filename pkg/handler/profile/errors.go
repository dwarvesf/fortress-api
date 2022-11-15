package profile

import "errors"

var (
	ErrEmployeeNotFound     = errors.New("employee not found")
	ErrInvalidFileExtension = errors.New("invalid file extension")
	ErrInvalidFileSize      = errors.New("invalid file size")
	ErrFileAlreadyExisted   = errors.New("file already existed")
)
