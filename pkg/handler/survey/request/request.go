package request

import (
	"github.com/dwarvesf/fortress-api/pkg/handler/survey/errs"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type GetListSurveyInput struct {
	model.Pagination

	Subtype    string   `json:"subtype" form:"subtype" binding:"required"`
	ProjectIDs []string `json:"projectIDs" form:"projectIDs"`
}

func (i *GetListSurveyInput) Validate() error {
	if i.Subtype == "" || !model.EventSubtype(i.Subtype).IsSurveyValid() {
		return errs.ErrInvalidEventType
	}

	for _, projectID := range i.ProjectIDs {
		if projectID != "" && !model.IsUUIDFromString(projectID) {
			return errs.ErrInvalidProjectID
		}
	}

	return nil
}

type GetSurveyDetailQuery struct {
	model.Pagination
	Keyword string `json:"keyword" form:"keyword"`
}

type GetSurveyDetailInput struct {
	EventID string
	Query   GetSurveyDetailQuery
}

func (i *GetSurveyDetailInput) Validate() error {
	if i.EventID == "" || !model.IsUUIDFromString(i.EventID) {
		return errs.ErrInvalidEventID
	}

	return nil
}

type Survey struct {
	TopicID      model.UUID   `json:"topicID" form:"topicID" binding:"required"`
	Participants []model.UUID `json:"participants" form:"participants" binding:"required"`
}

type SendSurveyInput struct {
	Topics []Survey `json:"topics" form:"topics" binding:"required"`
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
