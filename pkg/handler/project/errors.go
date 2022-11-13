package project

import "errors"

var (
	ErrInvalidProjectType         = errors.New("invalid project type")
	ErrInvalidProjectStatus       = errors.New("invalid project status")
	ErrInvalidProjectMemberStatus = errors.New("invalid project member status")
	ErrInvalidStartDate           = errors.New("invalid start date")
	ErrProjectNotFound            = errors.New("project not found")
)
