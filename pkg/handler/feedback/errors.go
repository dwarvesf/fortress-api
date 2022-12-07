package feedback

import "errors"

var (
	ErrEventNotFound                 = errors.New("event not found")
	ErrEmployeeEventReviewerNotFound = errors.New("employee event reviewer not found")
	ErrTopicNotFound                 = errors.New("topic not found")
	ErrInvalidReviewerStatus         = errors.New("invalid reviewer status")
	ErrInvalidEventType              = errors.New("invalid event type")
	ErrInvalidEventID                = errors.New("invalid eventID")
	ErrInvalidReviewerID             = errors.New("invalid reviewer id")
	ErrInvalidFeedbackID             = errors.New("invalid feedback id")
	ErrInvalidTopicID                = errors.New("invalid topic id")
)
