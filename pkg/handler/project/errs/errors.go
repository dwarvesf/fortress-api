package errs

import (
	"errors"
	"fmt"
)

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
	ErrInvalidWorkUnitID          = errors.New("invalid work unit ID")
	ErrInvalidWorkUnitType        = errors.New("invalid work unit type")
	ErrInvalidWorkUnitStatus      = errors.New("invalid work unit status")
	ErrInvalidStackID             = errors.New("invalid stack ID")
	ErrInvalidWorkUnitStacks      = errors.New("invalid work unit stacks")
	ErrInvalidInActiveMember      = errors.New("member is not active in work unit")
	ErrInvalidProjectFunction     = errors.New("invalid project function value")

	ErrProjectNotFound         = errors.New("project not found")
	ErrCountryNotFound         = errors.New("country not found")
	ErrEmployeeNotFound        = errors.New("employee not found")
	ErrSeniorityNotFound       = errors.New("seniority not found")
	ErrProjectSlotNotFound     = errors.New("project slot not found")
	ErrProjectMemberNotFound   = errors.New("project member not found")
	ErrProjectHeadNotFound     = errors.New("project head not found")
	ErrAccountManagerNotFound  = errors.New("account manager not found")
	ErrDeliveryManagerNotFound = errors.New("delivery manager not found")
	ErrStackNotFound           = errors.New("stack not found")
	ErrWorkUnitNotFound        = errors.New("work unit not found")
	ErrProjectNotExisted       = errors.New("project not existed")

	ErrMemberIsInactive                 = errors.New("member is inactive")
	ErrSlotIsInactive                   = errors.New("slot is inactive")
	ErrEmployeeIDCannotBeChanged        = errors.New("employeeID cannot be changed")
	ErrSlotIDCannotBeChanged            = errors.New("slotID cannot be changed")
	ErrPositionsIsEmpty                 = errors.New("positions is empty")
	ErrCouldNotDeleteInactiveMember     = errors.New("could not delete inactive member")
	ErrFailToCreateProjectHead          = errors.New("fail to create project head")
	ErrFailToFindProjectHead            = errors.New("fail to find project head")
	ErrFailToUpdateLeftDate             = errors.New("fail to update left date for project head")
	ErrProjectMemberExists              = errors.New("project member already exists")
	ErrMemberIsNotActiveInProject       = errors.New("member is not active in project")
	ErrFailToCheckInputExistence        = errors.New("failed to check input existance")
	ErrFailToDeleteWorkUnitStack        = errors.New("failed to delete work unit stack in database")
	ErrFailedToCreateWorkUnitStack      = errors.New("failed to create work unit stack")
	ErrFailedToGetWorkUnitMember        = errors.New("failed to get work unit member")
	ErrFailedToUpdateWorkUnitMember     = errors.New("failed to update work unit member in database")
	ErrFailedToSoftDeleteWorkUnitMember = errors.New("failed to soft delete work unit member")
	ErrFailedToGetProjectMember         = errors.New("failed to get project member in database")
	ErrFailedToCreateWorkUnitMember     = errors.New("failed to create work unit member")
	ErrSlotAlreadyContainsAnotherMember = errors.New("slot already contains another member")
	ErrDuplicateProjectCode             = errors.New("project code is duplicated")

	ErrInvalidFileExtension = errors.New("invalid file extension")
	ErrInvalidFileSize      = errors.New("invalid file size")
)

// ErrPositionNotFoundWithID returns unauthorized custom error
func ErrPositionNotFoundWithID(id string) error {
	return fmt.Errorf("position not found: %v", id)
}

func ErrStackNotFoundWithID(id string) error {
	return fmt.Errorf("stack not found: %v", id)
}
