package employee

import (
	"errors"
)

var (
	ErrInvalidEmployeeStatus = errors.New("invalid value for employee status")
	ErrInvalidEditType       = errors.New("invalid value for query type")
)
