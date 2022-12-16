package errs

import (
	"errors"
	"fmt"
)

var (
	// record not found errors
	ErrEventNotFound         = errors.New("event not found")
	ErrReviewerNotFound      = errors.New("reviewer not found")
	ErrTopicNotFound         = errors.New("topic not found")
	ErrEmployeeNotFound      = errors.New("employee not found")
	ErrEventReviewerNotFound = errors.New("employee event reviewer not found")

	// invalid errors
	ErrInvalidAnswers        = errors.New("invalid answers data")
	ErrInvalidEventID        = errors.New("invalid event id")
	ErrInvalidReviewerStatus = errors.New("invalid reviewer status")
	ErrInvalidEventType      = errors.New("invalid event type")
	ErrInvalidReviewerID     = errors.New("invalid reviewer id")
	ErrInvalidFeedbackID     = errors.New("invalid feedback id")
	ErrInvalidTopicID        = errors.New("invalid topic id")
	ErrInvalidEventSubType   = errors.New("invalid event subtype")
	ErrInvalidQuarter        = errors.New("invalid quarter")

	// other errors
	ErrEventAlreadyExisted      = errors.New("event already existed")
	ErrReviewAlreadySent        = errors.New("review already sent")
	ErrUnansweredquestions      = errors.New("must answer all questions")
	ErrCouldNotEditDoneFeedback = errors.New("could not edit the feedback marked as done")
	ErrEmployeeNotReady         = errors.New("employee not ready")
	ErrAlreadySent              = errors.New("surveys already sent to all participants")
	ErrUnfinishedReviewer       = errors.New("all reviewers have to finish before marked done")
	ErrCanNotUpdateParticipants = errors.New("can not update participants")
)

func ErrEventQuestionNotFound(id string) error {
	return fmt.Errorf("employee event question not found: %v", id)
}
