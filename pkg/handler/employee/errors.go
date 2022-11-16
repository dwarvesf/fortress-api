package employee

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
)

// errStackNotFound returns unauthorized custom error
func errStackNotFound(id string) error {
	return fmt.Errorf("stack not found: %v", id)
}

// errPositionNotFound returns unauthorized custom error
func errPositionNotFound(id string) error {
	return fmt.Errorf("position not found: %v", id)
}
