package errs

import "errors"

var (
	ErrInvalidProjectID = errors.New("invalid project ID")
	ErrProjectNotFound  = errors.New("project not found")
)
