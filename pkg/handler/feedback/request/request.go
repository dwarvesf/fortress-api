package request

import (
	"github.com/dwarvesf/fortress-api/pkg/handler/feedback/errs"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type GetListFeedbackInput struct {
	view.Pagination

	Status string `json:"status" form:"status"`
}

func (i *GetListFeedbackInput) Validate() error {
	if i.Status != "" && !model.EventReviewerStatus(i.Status).IsValid() {
		return errs.ErrInvalidReviewerStatus
	}

	return nil
}

type DetailInput struct {
	EventID string
	TopicID string
}

func (i *DetailInput) Validate() error {
	if i.EventID == "" || !model.IsUUIDFromString(i.EventID) {
		return errs.ErrInvalidFeedbackID
	}

	if i.TopicID == "" || !model.IsUUIDFromString(i.TopicID) {
		return errs.ErrInvalidTopicID
	}

	return nil
}

type BasicEventQuestionRequest struct {
	EventQuestionID model.UUID `json:"eventQuestionID" form:"eventQuestionID" binding:"required"`
	Answer          string     `json:"answer" form:"answer"`
	Note            string     `json:"note" form:"note"`
} // @name BasicEventQuestionRequest

type SubmitFeedbackRequest struct {
	Answers []BasicEventQuestionRequest `json:"answers" form:"answers" binding:"required"`
	Status  model.EventReviewerStatus   `json:"status" form:"status" binding:"required"`
} // @name SubmitFeedbackRequest

func (i *SubmitFeedbackRequest) Validate() error {
	if !i.Status.IsValid() {
		return errs.ErrInvalidReviewerStatus
	}

	return nil
}

type SubmitInput struct {
	Body    SubmitFeedbackRequest
	EventID string
	TopicID string
}

func (i *SubmitInput) Validate() error {
	if i.EventID == "" || !model.IsUUIDFromString(i.EventID) {
		return errs.ErrInvalidEventID
	}

	if i.TopicID == "" || !model.IsUUIDFromString(i.TopicID) {
		return errs.ErrInvalidTopicID
	}

	return i.Body.Validate()
}
