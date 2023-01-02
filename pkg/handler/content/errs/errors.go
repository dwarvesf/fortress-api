package errs

import (
	"errors"
)

var (
	ErrInvalidFileExtension = errors.New("invalid file extension")
	ErrInvalidFileSize      = errors.New("invalid file size")
	ErrFileAlreadyExisted   = errors.New("file already existed")
	ErrEmployeeIDRequired   = errors.New("employeeID is required")
)
