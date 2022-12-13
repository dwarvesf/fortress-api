package request

import (
	"github.com/dwarvesf/fortress-api/pkg/handler/feedback/errs"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type GetListFeedbackInput struct {
	model.Pagination

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

type GetListSurveyInput struct {
	model.Pagination

	Subtype string `json:"subtype" form:"subtype" binding:"required"`
}

func (i *GetListSurveyInput) Validate() error {
	if i.Subtype == "" || !model.EventSubtype(i.Subtype).IsSurveyValid() {
		return errs.ErrInvalidEventType
	}

	return nil
}

type GetSurveyDetailInput struct {
	EventID    string
	Pagination model.Pagination
}

func (i *GetSurveyDetailInput) Validate() error {
	if i.EventID == "" || !model.IsUUIDFromString(i.EventID) {
		return errs.ErrInvalidEventID
	}

	return nil
}

type BasicEventQuestionInput struct {
	EventQuestionID model.UUID `json:"eventQuestionID" form:"eventQuestionID" binding:"required"`
	Answer          string     `json:"answer" form:"answer"`
	Note            string     `json:"note" form:"note"`
}

type SubmitBody struct {
	Answers []BasicEventQuestionInput `json:"answers" form:"answers" binding:"required"`
	Status  model.EventReviewerStatus `json:"status" form:"status" binding:"required"`
}

func (i *SubmitBody) Validate() error {
	if !i.Status.IsValid() {
		return errs.ErrInvalidReviewerStatus
	}

	return nil
}

type SubmitInput struct {
	Body    SubmitBody
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

type PerformanceReviewTopic struct {
	TopicID      model.UUID   `json:"topicID" form:"topicID" binding:"required"`
	Participants []model.UUID `json:"participants" form:"participants" binding:"required"`
}

type SendPerformanceReviewInput struct {
	Topics []PerformanceReviewTopic `json:"topics" form:"topics" binding:"required"`
}

// CreateSurveyFeedbackInput view for create survey feedback
type CreateSurveyFeedbackInput struct {
	Quarter string `json:"quarter" binding:"required"`
	Year    int    `json:"year" binding:"required"`
	Type    string `json:"type" binding:"required"`
}

type PeerReviewDetailInput struct {
	EventID string
	TopicID string
}

func (i *PeerReviewDetailInput) Validate() error {
	if i.EventID == "" || !model.IsUUIDFromString(i.EventID) {
		return errs.ErrInvalidFeedbackID
	}

	if i.TopicID == "" || !model.IsUUIDFromString(i.TopicID) {
		return errs.ErrInvalidTopicID
	}

	return nil
}

// UpdateTopicReviewersBody view for update topic reviewers
type UpdateTopicReviewersBody struct {
	ReviewerIDs []model.UUID `json:"reviewerIDs"`
}

// UpdateTopicReviewersInput input of update topic reviewers request
type UpdateTopicReviewersInput struct {
	EventID string
	TopicID string
	Body    UpdateTopicReviewersBody
}

func (i *UpdateTopicReviewersInput) Validate() error {
	if i.EventID == "" || !model.IsUUIDFromString(i.EventID) {
		return errs.ErrInvalidEventID
	}

	if i.TopicID == "" || !model.IsUUIDFromString(i.TopicID) {
		return errs.ErrInvalidTopicID
	}

	return nil
}

// DeleteTopicReviewersBody view for update topic reviewers
type DeleteTopicReviewersBody struct {
	ReviewerIDs []model.UUID `json:"reviewerIDs"`
}

// DeleteTopicReviewersInput input of update topic reviewers request
type DeleteTopicReviewersInput struct {
	EventID string
	TopicID string
	Body    DeleteTopicReviewersBody
}

func (i *DeleteTopicReviewersInput) Validate() error {
	if i.EventID == "" || !model.IsUUIDFromString(i.EventID) {
		return errs.ErrInvalidEventID
	}

	if i.TopicID == "" || !model.IsUUIDFromString(i.TopicID) {
		return errs.ErrInvalidTopicID
	}

	return nil
}
