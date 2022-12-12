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
		return ErrInvalidReviewerStatus
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
		return ErrInvalidEventID
	}

	if i.TopicID == "" || !model.IsUUIDFromString(i.TopicID) {
		return ErrInvalidTopicID
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
