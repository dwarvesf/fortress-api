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
	ErrCountryNotFound              = errors.New("country not found")
	ErrEmployeeNotFound             = errors.New("employee not found")
	ErrSeniorityNotFound            = errors.New("seniority not found")
	ErrProjectSlotNotFound          = errors.New("project slot not found")
	ErrMemberIsInactive             = errors.New("member is inactive")
	ErrSlotIsInactive               = errors.New("slot is inactive")
	ErrEmployeeIDCannotBeChanged    = errors.New("employeeID cannot be changed")
	ErrPositionsIsEmpty             = errors.New("positions is empty")
	ErrProjectMemberNotFound        = errors.New("project member not found")
	ErrCouldNotDeleteInactiveMember = errors.New("could not delete inactive member")
	ErrAccountManagerNotFound       = errors.New("account manager not found")
	ErrDeliveryManagerNotFound      = errors.New("delivery manager not found")
	ErrFailToCreateProjectHead      = errors.New("fail to create project head")
	ErrFailToFindProjectHead        = errors.New("fail to find project head")
	ErrFailToUpdateLeftDate         = errors.New("fail to update left date for project head")
)

// errPositionNotFound returns unauthorized custom error
func errPositionNotFound(id string) error {
	return fmt.Errorf("position not found: %v", id)
}

func errStackNotFound(id string) error {
	return fmt.Errorf("stack not found: %v", id)
}
