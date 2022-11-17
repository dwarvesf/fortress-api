package project

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidProjectID             = errors.New("invalid project ID")
	ErrInvalidProjectType           = errors.New("invalid project type")
	ErrInvalidProjectStatus         = errors.New("invalid project status")
	ErrInvalidProjectMemberStatus   = errors.New("invalid project member status")
	ErrInvalidStartDate             = errors.New("invalid start date")
	ErrInvalidDeploymentType        = errors.New("invalid deployment type")
	ErrInvalidJoinedDate            = errors.New("invalid joined date")
	ErrInvalidLeftDate              = errors.New("invalid left date")
	ErrInvalidMemberID              = errors.New("invalid member ID")
	ErrProjectNotFound              = errors.New("project not found")
	ErrEmployeeNotFound             = errors.New("employee not found")
	ErrSeniorityNotFound            = errors.New("seniority not found")
	ErrProjectSlotNotFound          = errors.New("project slot not found")
	ErrMemberIsInactive             = errors.New("member is inactive")
	ErrSlotIsInactive               = errors.New("slot is inactive")
	ErrEmployeeIDCannotBeChanged    = errors.New("employeeID cannot be changed")
	ErrPositionsIsEmpty             = errors.New("positions is empty")
	ErrProjectMemberNotFound        = errors.New("project member not found")
	ErrCouldNotDeleteInactiveMember = errors.New("could not delete inactive member")
)

// errPositionNotFound returns unauthorized custom error
func errPositionNotFound(id string) error {
	return fmt.Errorf("position not found: %v", id)
}
