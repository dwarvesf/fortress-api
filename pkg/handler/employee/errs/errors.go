package errs

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidEmployeeStatus = errors.New("invalid value for employee status")
	ErrCantFindLineManager   = errors.New("can't find line manager with the input id")
	ErrEmployeeExisted       = errors.New("can't create existed employee")
	ErrPositionNotfound      = errors.New("position not found")
	ErrSeniorityNotfound     = errors.New("seniority not found")
	ErrRoleNotfound          = errors.New("role not found")
	ErrLineManagerNotFound   = errors.New("line manager not found")
	ErrEmployeeNotFound      = errors.New("employee not found")
	ErrStackNotFound         = errors.New("stack not found")
	ErrPositionNotFound      = errors.New("position not found")
	ErrChapterNotFound       = errors.New("chapter not found")
	ErrSeniorityNotFound     = errors.New("seniority not found")
	ErrInvalidEmployeeID     = errors.New("invalid employee ID")
	ErrInvalidPositionID     = errors.New("invalid position ID")
	ErrInvalidStackID        = errors.New("invalid stack ID")
	ErrInvalidProjectID      = errors.New("invalid project ID")
	ErrInvalidFileExtension  = errors.New("invalid file extension")
	ErrInvalidFileSize       = errors.New("invalid file size")
	ErrFileAlreadyExisted    = errors.New("file already existed")
)

// ErrPositionNotFoundWithID returns bad request custom error
func ErrPositionNotFoundWithID(id string) error {
	return fmt.Errorf("position not found: %v", id)
}

// ErrChapterNotFoundWithID returns bad request custom error
func ErrChapterNotFoundWithID(id string) error {
	return fmt.Errorf("chapter not found: %v", id)
}

// ErrStackNotFoundWithID returns bad request custom error
func ErrStackNotFoundWithID(id string) error {
	return fmt.Errorf("stack not found: %v", id)
}
