package project

import "errors"

var (
	ErrInvalidProjectID           = errors.New("invalid project ID")
	ErrInvalidProjectType         = errors.New("invalid project type")
	ErrInvalidProjectStatus       = errors.New("invalid project status")
	ErrInvalidProjectMemberStatus = errors.New("invalid project member status")
	ErrInvalidStartDate           = errors.New("invalid start date")
	ErrInvalidDeploymentType      = errors.New("invalid deployment type")
	ErrInvalidJoinedDate          = errors.New("invalid joined date")
	ErrInvalidLeftDate            = errors.New("invalid left date")
	ErrInvalidMemberID            = errors.New("invalid member ID")
	ErrProjectNotFound            = errors.New("project not found")
	ErrProjectSlotNotFound        = errors.New("project slot not found")
	ErrMemberIsInactive           = errors.New("member is inactive")
	ErrSlotIsInactive             = errors.New("slot is inactive")
	ErrEmployeeIDCannotBeChanged  = errors.New("employeeID cannot be changed")
)
