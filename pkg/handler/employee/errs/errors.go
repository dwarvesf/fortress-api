package errs

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/controller/employee"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

var (
	ErrInvalidEmployeeID          = errors.New("invalid employee ID")
	ErrInvalidEmployeeStatus      = errors.New("invalid value for employee status")
	ErrInvalidJoinedDate          = errors.New("invalid join date")
	ErrInvalidPositionCode        = errors.New("invalid position code")
	ErrInvalidStackCode           = errors.New("invalid stack code")
	ErrInvalidProjectCode         = errors.New("invalid project code")
	ErrInvalidChapterCode         = errors.New("invalid chapter code")
	ErrInvalidSeniorityCode       = errors.New("invalid seniority code")
	ErrInvalidOrganizationCode    = errors.New("invalid organization code")
	ErrInvalidEmailDomain         = errors.New("invalid email domain")
	ErrRoleCannotBeEmpty          = errors.New("role cannot be empty")
	ErrCountryNotFound            = errors.New("country not found")
	ErrCityDoesNotBelongToCountry = errors.New("city does not belong to country")
	ErrInvalidLineManagerID       = errors.New("invalid line manager ID")
	ErrInvalidOrganizationID      = errors.New("invalid organization ID")
	ErrInvalidReferredBy          = errors.New("invalid referred by")
	ErrInvalidSortType            = errors.New("invalid sort type")
)

func ConvertControllerErr(c *gin.Context, err error) {
	if err == nil {
		return
	}

	var status int

	switch {
	case errors.Is(err, employee.ErrEmployeeNotFound),
		errors.Is(err, employee.ErrLineManagerNotFound),
		errors.Is(err, employee.ErrRoleNotfound),
		errors.Is(err, employee.ErrSeniorityNotfound),
		errors.Is(err, employee.ErrReferrerNotFound),
		errors.Is(err, employee.ErrOrganizationNotFound),
		errors.Is(err, employee.ErrStackNotFound),
		errors.Is(err, employee.ErrPositionNotFound):
		status = http.StatusNotFound

	case errors.Is(err, employee.ErrInvalidJoinedDate),
		errors.Is(err, employee.ErrInvalidLeftDate),
		errors.Is(err, employee.ErrLeftDateBeforeJoinedDate),
		errors.Is(err, employee.ErrEmployeeExisted),
		errors.Is(err, employee.ErrInvalidCountryOrCity),
		errors.Is(err, employee.ErrInvalidFileExtension),
		errors.Is(err, employee.ErrInvalidFileSize),
		errors.Is(err, employee.ErrInvalidAccountRole),
		errors.Is(err, employee.ErrEmailExisted),
		errors.Is(err, employee.ErrTeamEmailExisted),
		errors.Is(err, employee.ErrSalaryAdvanceExceedAmount),
		errors.Is(err, employee.ErrEmployeeNotFullTime),
		errors.Is(err, employee.ErrPersonalEmailExisted):
		status = http.StatusBadRequest

	default:
		status = http.StatusInternalServerError
	}

	c.JSON(status, view.CreateResponse[any](nil, nil, err, nil, ""))
}
