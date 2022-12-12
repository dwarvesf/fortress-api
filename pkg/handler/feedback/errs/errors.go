package errs

import (
	"errors"
	"fmt"
)

var (
	ErrEventNotFound                 = errors.New("event not found")
	ErrReviewerNotFound              = errors.New("reviewer not found")
	ErrEmployeeEventReviewerNotFound = errors.New("employee event reviewer not found")
	ErrInvalidEventType              = errors.New("invalid event type")
	ErrInvalidReviewerID             = errors.New("invalid reviewer id")
	ErrInvalidFeedbackID             = errors.New("invalid feedback id")
	ErrTopicNotFound                 = errors.New("topic not found")
	ErrEventReviewerNotFound         = errors.New("employee event reviewer not found")

	ErrInvalidAnswers        = errors.New("invalid answers data")
	ErrInvalidEventID        = errors.New("invalid event id")
	ErrInvalidReviewerStatus = errors.New("invalid reviewer status")
	ErrReviewAlreadySent     = errors.New("review already sent")

	ErrUnansweredquestions      = errors.New("must answer all questions")
	ErrCouldNotEditDoneFeedback = errors.New("could not edit the feedback marked as done")
	ErrInvalidTopicID           = errors.New("invalid topic id")
	ErrInvalidEventSubType      = errors.New("invalid event subtype")
	ErrInvalidQuarter           = errors.New("invalid quarter")
	ErrEmployeeNotFound         = errors.New("employee not found")
	ErrEmployeeNotReady         = errors.New("employee not ready")
	ErrAlreadySent              = errors.New("surveys already sent to all participants")
)

func ErrEventQuestionNotFound(id string) error {
	return fmt.Errorf("employee event question not found: %v", id)
}
