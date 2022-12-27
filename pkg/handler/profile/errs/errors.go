package errs

import "errors"

var (
	ErrEmployeeNotFound      = errors.New("employee not found")
	ErrCountryNotFound       = errors.New("country not found")
	ErrInvalidCountryAndCity = errors.New("invalid country and city")
	ErrInvalidFileExtension  = errors.New("invalid file extension")
	ErrInvalidFileSize       = errors.New("invalid file size")
	ErrFileAlreadyExisted    = errors.New("file already existed")
)
