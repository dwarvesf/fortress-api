package employee

import (
	"errors"
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
)
