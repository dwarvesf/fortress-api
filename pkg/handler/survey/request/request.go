package request

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/handler/survey/errs"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

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

type GetSurveyDetailQuery struct {
	model.Pagination
	Keyword  string   `json:"keyword" form:"keyword"`
	Projects []string `json:"projects" form:"projects"`
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

type SendSurveyInput struct {
	Type     string       `json:"type" form:"type" binding:"required"`
	TopicIDs []model.UUID `json:"topicIDs" form:"topicIDs"`
}

// CreateSurveyFeedbackInput view for create survey feedback
type CreateSurveyFeedbackInput struct {
	Quarter  string `json:"quarter"`
	Year     int    `json:"year"`
	Type     string `json:"type" binding:"required"`
	FromDate string `json:"fromDate"`
	ToDate   string `json:"toDate"`
}

// Validate input for create survey feedback
func (i *CreateSurveyFeedbackInput) Validate() error {
	if !model.EventSubtype(i.Type).IsValidSurvey() {
		return errs.ErrInvalidEventSubType
	}

	if i.Type == model.EventSubtypeWork.String() {
		fromDate, err := time.Parse("2006-01-02", i.FromDate)
		if err != nil {
			return errs.ErrInvalidDate
		}

		toDate, err := time.Parse("2006-01-02", i.ToDate)
		if err != nil {
			return errs.ErrInvalidDate
		}

		if fromDate.After(toDate) {
			return errs.ErrInvalidDateRange
		}
	}

	return nil
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
