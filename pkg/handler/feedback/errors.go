package feedback

import "errors"

var (
	ErrEmployeeEventReviewerNotFound = errors.New("employee event reviewer not found")
	ErrTopicNotFound                 = errors.New("topic not found")
	ErrInvalidReviewerStatus         = errors.New("invalid reviewer status")
	ErrInvalidReviewerID             = errors.New("invalid reviewer id")
	ErrInvalidFeedbackID             = errors.New("invalid feedback id")
	ErrInvalidTopicID                = errors.New("invalid topic id")
)
