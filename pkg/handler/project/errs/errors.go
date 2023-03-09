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
	ErrInvalidDeploymentType      = errors.New("invalid deployment type")
	ErrInvalidStartDate           = errors.New("invalid start date")
	ErrInvalidEndDate             = errors.New("invalid end date")
	ErrInvalidMemberID            = errors.New("invalid member ID")
	ErrInvalidSlotID              = errors.New("invalid slot ID")
	ErrInvalidWorkUnitID          = errors.New("invalid work unit ID")
	ErrInvalidWorkUnitType        = errors.New("invalid work unit type")
	ErrInvalidWorkUnitStatus      = errors.New("invalid work unit status")
	ErrInvalidWorkUnitStacks      = errors.New("invalid work unit stacks")
	ErrInvalidInActiveMember      = errors.New("member is not active in work unit")
	ErrMemberIsNotProjectLead     = errors.New("project is not managed by signed-in user")
	ErrInvalidProjectFunction     = errors.New("invalid project function value")

	ErrProjectNotFound         = errors.New("project not found")
	ErrProjectNotionNotFound   = errors.New("project notion not found")
	ErrCountryNotFound         = errors.New("country not found")
	ErrBankAccountNotFound     = errors.New("bank account not found")
	ErrEmployeeNotFound        = errors.New("employee not found")
	ErrSeniorityNotFound       = errors.New("seniority not found")
	ErrProjectSlotNotFound     = errors.New("project slot not found")
	ErrProjectMemberNotFound   = errors.New("project member not found")
	ErrAccountManagerNotFound  = errors.New("account manager not found")
	ErrDeliveryManagerNotFound = errors.New("delivery manager not found")
	ErrWorkUnitNotFound        = errors.New("work unit not found")
	ErrClientNotFound          = errors.New("client not found")
	ErrOrganizationNotFound    = errors.New("organization not found")
	ErrProjectNotExisted       = errors.New("project not existed")

	ErrMemberIsInactive                 = errors.New("member is inactive")
	ErrEmployeeWorkedOnTheProject       = errors.New("employee worked on the project")
	ErrPositionsIsEmpty                 = errors.New("positions is empty")
	ErrMemberIsNotActiveInProject       = errors.New("member is not active in project")
	ErrFailToCheckInputExistence        = errors.New("failed to check input existence")
	ErrFailToDeleteWorkUnitStack        = errors.New("failed to delete work unit stack in database")
	ErrFailedToCreateWorkUnitStack      = errors.New("failed to create work unit stack")
	ErrFailedToGetWorkUnitMember        = errors.New("failed to get work unit member")
	ErrFailedToUpdateWorkUnitMember     = errors.New("failed to update work unit member in database")
	ErrFailedToSoftDeleteWorkUnitMember = errors.New("failed to soft delete work unit member")
	ErrFailedToGetProjectMember         = errors.New("failed to get project member in database")
	ErrFailedToCreateWorkUnitMember     = errors.New("failed to create work unit member")
	ErrSlotAlreadyContainsAnotherMember = errors.New("slot already contains another member")
	ErrDuplicateProjectCode             = errors.New("project code is duplicated")
	ErrAccountManagerCannotEmpty        = errors.New("account manager cannot empty")
	ErrTotalCommissionRateMustBe100     = errors.New("total commission rate must be 100")

	ErrInvalidFileExtension         = errors.New("invalid file extension")
	ErrInvalidFileSize              = errors.New("invalid file size")
	ErrInvalidEmailDomainForClient  = errors.New("invalid email domain for client")
	ErrInvalidEmailDomainForProject = errors.New("invalid email domain for project")
)

// ErrPositionNotFoundWithID returns unauthorized custom error
func ErrPositionNotFoundWithID(id string) error {
	return fmt.Errorf("position not found: %v", id)
}

func ErrStackNotFoundWithID(id string) error {
	return fmt.Errorf("stack not found: %v", id)
}
