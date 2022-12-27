package errs

import (
	"errors"
	"fmt"
)

var (
	// record not found errors
	ErrEventNotFound         = errors.New("event not found")
	ErrProjectNotFound       = errors.New("project not found")
	ErrTopicNotFound         = errors.New("topic not found")
	ErrEmployeeNotFound      = errors.New("employee not found")
	ErrEventReviewerNotFound = errors.New("employee event reviewer not found")

	// invalid errors
	ErrInvalidEventID      = errors.New("invalid event id")
	ErrInvalidEventType    = errors.New("invalid event type")
	ErrInvalidReviewerID   = errors.New("invalid reviewer id")
	ErrInvalidFeedbackID   = errors.New("invalid feedback id")
	ErrInvalidTopicID      = errors.New("invalid topic id")
	ErrInvalidEventSubType = errors.New("invalid event subtype")
	ErrInvalidQuarter      = errors.New("invalid quarter")
	ErrInvalidYear         = errors.New("invalid year")
	ErrInvalidDate         = errors.New("invalid date")
	ErrInvalidDateRange    = errors.New("invalid date range")

	// other errors
	ErrEventAlreadyExisted      = errors.New("event already existed")
	ErrReviewAlreadySent        = errors.New("review already sent")
	ErrEmployeeNotReady         = errors.New("employee not ready")
	ErrCanNotUpdateParticipants = errors.New("can not update participants")
	ErrEventHasBeenDone         = errors.New("event has been done")
	ErrNoValidProjectForEvent   = errors.New("no valid project for event")
)

func ErrEventQuestionNotFound(id string) error {
	return fmt.Errorf("employee event question not found: %v", id)
}
