package employee

import (
	"errors"
)

var (
	ErrInvalidEmployeeStatus = errors.New("invalid value for employee status")
	ErrCantFindLineManager   = errors.New("can't find line manager with the input id")
)
