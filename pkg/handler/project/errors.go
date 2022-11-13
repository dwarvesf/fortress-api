package project

import "errors"

var (
	ErrInvalidProjectType   = errors.New("invalid project type")
	ErrInvalidProjectStatus = errors.New("invalid project status")
	ErrInvalidStartDate     = errors.New("invalid start date")
)
