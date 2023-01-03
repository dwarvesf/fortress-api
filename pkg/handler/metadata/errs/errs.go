package errs

import "errors"

var (
	ErrInvalidStackID    = errors.New("invalid stack ID")
	ErrStackNotFound     = errors.New("stack not found")
	ErrInvalidPositionID = errors.New("invalid Position ID")
	ErrPositionNotFound  = errors.New("Position not found")
)
