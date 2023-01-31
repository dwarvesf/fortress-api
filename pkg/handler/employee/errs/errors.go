package errs

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidLeftDate                        = errors.New("invalid left date")
	ErrInvalidJoinedDate                      = errors.New("invalid joined date")
	ErrLeftDateBeforeJoinedDate               = errors.New("left date could not be before joined date")
	ErrInvalidEmployeeStatus                  = errors.New("invalid value for employee status")
	ErrInvalidEmailDomain                     = errors.New("invalid email domain")
	ErrCantFindLineManager                    = errors.New("can't find line manager with the input id")
	ErrEmployeeExisted                        = errors.New("can't create existed employee")
	ErrPositionNotfound                       = errors.New("position not found")
	ErrSeniorityNotfound                      = errors.New("seniority not found")
	ErrRoleNotfound                           = errors.New("role not found")
	ErrLineManagerNotFound                    = errors.New("line manager not found")
	ErrOrganizationNotFound                   = errors.New("organization not found")
	ErrEmployeeMenteeNotFound                 = errors.New("employee mentee not found")
	ErrMenteeNotFound                         = errors.New("mentee not found")
	ErrEmployeeNotFound                       = errors.New("employee not found")
	ErrRoleNotFound                           = errors.New("role not found")
	ErrStackNotFound                          = errors.New("stack not found")
	ErrPositionNotFound                       = errors.New("position not found")
	ErrChapterNotFound                        = errors.New("chapter not found")
	ErrSeniorityNotFound                      = errors.New("seniority not found")
	ErrInvalidEmployeeID                      = errors.New("invalid employee ID")
	ErrInvalidMentorID                        = errors.New("invalid mentor ID")
	ErrInvalidMenteeID                        = errors.New("invalid mentee ID")
	ErrInvalidPositionID                      = errors.New("invalid position ID")
	ErrInvalidStackID                         = errors.New("invalid stack ID")
	ErrInvalidProjectID                       = errors.New("invalid project ID")
	ErrInvalidFileExtension                   = errors.New("invalid file extension")
	ErrInvalidFileSize                        = errors.New("invalid file size")
	ErrFileAlreadyExisted                     = errors.New("file already existed")
	ErrInvalidPositionCode                    = errors.New("invalid position code")
	ErrInvalidStackCode                       = errors.New("invalid stack code")
	ErrInvalidProjectCode                     = errors.New("invalid project code")
	ErrInvalidChapterCode                     = errors.New("invalid chapter code")
	ErrInvalidSeniorityCode                   = errors.New("invalid seniority code")
	ErrInvalidOrganizationCode                = errors.New("invalid organization code")
	ErrInvalidCountryOrCity                   = errors.New("invalid country or city")
	ErrCouldNotMentorThemselves               = errors.New("employee could not be their own mentor")
	ErrCouldNotMentorTheirMentor              = errors.New("employee could not be mentor of their mentor")
	ErrMenteeLeft                             = errors.New("mentee is left")
	ErrEmployeeLeft                           = errors.New("employee is left")
	ErrInvalidAccountRole                     = errors.New("invalid account role")
	ErrCouldNotAssignRoleForSameLevelEmployee = errors.New("could not assign role for the same level employee")
)

// ErrPositionNotFoundWithID returns bad request custom error
func ErrPositionNotFoundWithID(id string) error {
	return fmt.Errorf("position not found: %v", id)
}

// ErrChapterNotFoundWithID returns bad request custom error
func ErrChapterNotFoundWithID(id string) error {
	return fmt.Errorf("chapter not found: %v", id)
}

// ErrStackNotFoundWithID returns bad request custom error
func ErrStackNotFoundWithID(id string) error {
	return fmt.Errorf("stack not found: %v", id)
}

// ErrOrganizationNotFoundWithID returns bad request custom error
func ErrOrganizationNotFoundWithID(id string) error {
	return fmt.Errorf("organization not found: %v", id)
}
