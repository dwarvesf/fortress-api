package employee

import (
	"errors"
)

var (
	ErrCannotSelfReferral                         = errors.New("cannot self referral")
	ErrCantFindLineManager                        = errors.New("can't find line manager with the input id")
	ErrChapterNotFound                            = errors.New("chapter not found")
	ErrCouldNotAssignRoleForSameLevelEmployee     = errors.New("could not assign role for the same level employee")
	ErrCouldNotMentorTheirMentor                  = errors.New("employee could not be mentor of their mentor")
	ErrCouldNotMentorThemselves                   = errors.New("employee could not be their own mentor")
	ErrCurrencyNotFound                           = errors.New("currency not found")
	ErrEmailExisted                               = errors.New("email already exists")
	ErrTeamEmailExisted                           = errors.New("team email already exists")
	ErrPersonalEmailExisted                       = errors.New("personal email already exists")
	ErrEmployeeExisted                            = errors.New("can't create existed employee")
	ErrEmployeeLeft                               = errors.New("employee is left")
	ErrEmployeeMenteeNotFound                     = errors.New("employee mentee not found")
	ErrEmployeeNotFound                           = errors.New("employee not found")
	ErrSalaryAdvanceNotPayBack                    = errors.New("employee not pay back salary advance")
	ErrSalaryAdvanceExceedAmount                  = errors.New("your request is exceed amount you can advance")
	ErrSalaryAdvanceMaxCapInvalid                 = errors.New("max cap salary invalid")
	ErrFileAlreadyExisted                         = errors.New("file already existed")
	ErrInvalidAccountRole                         = errors.New("invalid account role")
	ErrInvalidChapterCode                         = errors.New("invalid chapter code")
	ErrInvalidCountryOrCity                       = errors.New("invalid country or city")
	ErrInvalidEmailDomain                         = errors.New("invalid email domain")
	ErrInvalidEmployeeID                          = errors.New("invalid employee ID")
	ErrInvalidEmployeeStatus                      = errors.New("invalid value for employee status")
	ErrInvalidFileExtension                       = errors.New("invalid file extension")
	ErrInvalidFileSize                            = errors.New("invalid file size")
	ErrInvalidJoinedDate                          = errors.New("invalid joined date")
	ErrInvalidLeftDate                            = errors.New("invalid left date")
	ErrInvalidMenteeID                            = errors.New("invalid mentee ID")
	ErrInvalidMentorID                            = errors.New("invalid mentor ID")
	ErrInvalidOrganizationCode                    = errors.New("invalid organization code")
	ErrInvalidPositionCode                        = errors.New("invalid position code")
	ErrInvalidPositionID                          = errors.New("invalid position ID")
	ErrInvalidProjectCode                         = errors.New("invalid project code")
	ErrInvalidProjectID                           = errors.New("invalid project ID")
	ErrInvalidSeniorityCode                       = errors.New("invalid seniority code")
	ErrInvalidStackCode                           = errors.New("invalid stack code")
	ErrInvalidStackID                             = errors.New("invalid stack ID")
	ErrLeftDateBeforeJoinedDate                   = errors.New("left date could not be before joined date")
	ErrLineManagerNotFound                        = errors.New("line manager not found")
	ErrMenteeLeft                                 = errors.New("mentee is left")
	ErrMenteeNotFound                             = errors.New("mentee not found")
	ErrOrganizationNotFound                       = errors.New("organization not found")
	ErrPositionNotFound                           = errors.New("position not found")
	ErrPositionNotfound                           = errors.New("position not found")
	ErrReferrerNotFound                           = errors.New("referrer not found")
	ErrRoleNotFound                               = errors.New("role not found")
	ErrRoleNotfound                               = errors.New("role not found")
	ErrSeniorityNotFound                          = errors.New("seniority not found")
	ErrSeniorityNotfound                          = errors.New("seniority not found")
	ErrStackNotFound                              = errors.New("stack not found")
	ErrDiscordAccountNotFound                     = errors.New("discord account not found")
	ErrDiscordAccountAlreadyUsedByAnotherEmployee = errors.New("discord account already used by another employee")
	ErrCouldNotFoundDiscordMemberInGuild          = errors.New("could not found discord member in the guild")
	ErrEmployeeNotFullTime                        = errors.New("employee is not full time")
)
