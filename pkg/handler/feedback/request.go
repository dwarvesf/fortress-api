package feedback

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type GetListFeedbackInput struct {
	model.Pagination

	Status string `json:"status" form:"status"`
}

func (i *GetListFeedbackInput) Validate() error {
	if i.Status != "" && !model.EventReviewerStatus(i.Status).IsValid() {
		return ErrInvalidReviewerStatus
	}

	return nil
}

type DetailInput struct {
	EventID string
	TopicID string
}

func (i *DetailInput) Validate() error {
	if i.EventID == "" || !model.IsUUIDFromString(i.EventID) {
		return ErrInvalidFeedbackID
	}

	if i.TopicID == "" || !model.IsUUIDFromString(i.TopicID) {
		return ErrInvalidTopicID
	}

	return nil
}

type GetListSurveyInput struct {
	model.Pagination

	Subtype string `json:"subtype" form:"subtype" binding:"required"`
}

func (i *GetListSurveyInput) Validate() error {
	if i.Subtype == "" || !model.EventSubtype(i.Subtype).IsSurveyValid() {
		return ErrInvalidEventType
	}

	return nil
}

type GetSurveyDetailInput struct {
	EventID    string
	Pagination model.Pagination
}

func (i *GetSurveyDetailInput) Validate() error {
	if i.EventID == "" || !model.IsUUIDFromString(i.EventID) {
		return ErrInvalidEventID
	}

	return nil
}
