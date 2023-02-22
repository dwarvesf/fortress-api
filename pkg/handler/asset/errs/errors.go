package errs

import "errors"

var (
	ErrEmployeeNotFound     = errors.New("employee not found")
	ErrProjectNotFound      = errors.New("project not found")
	ErrInvalidFileExtension = errors.New("invalid file extension")
	ErrInvalidFileType      = errors.New("invalid file type")
	ErrInvalidFileSize      = errors.New("invalid file size")
	ErrFileAlreadyExisted   = errors.New("file already existed")
)
