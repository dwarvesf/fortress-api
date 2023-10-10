package view

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

var DwarvesAuthor = BasicEmployeeInfo{
	FullName:    "Dwarves' team",
	DisplayName: "Dwarves' team",
}

type Feedback struct {
	TopicID         string             `json:"topicID"`
	Title           string             `json:"title"`
	EventID         string             `json:"eventID"`
	EmployeeID      string             `json:"employeeID"`
	ProjectID       string             `json:"projectID"`
	EventReviewerID string             `json:"eventReviewerID"`
	Type            string             `json:"type"`
	Subtype         string             `json:"subtype"`
	Status          string             `json:"status"`
	IsRead          bool               `json:"isRead"`
	LastUpdated     *time.Time         `json:"lastUpdated"`
	Author          *BasicEmployeeInfo `json:"author"`
} // @name Feedback

func ToListFeedback(eTopics []*model.EmployeeEventTopic) []Feedback {
	var results = make([]Feedback, 0, len(eTopics))

	for _, topic := range eTopics {
		if len(topic.EmployeeEventReviewers) == 0 {
			continue
		}

		feedback := Feedback{
			Title:           topic.Title,
			TopicID:         topic.ID.String(),
			EventID:         topic.EventID.String(),
			EmployeeID:      topic.EmployeeID.String(),
			ProjectID:       topic.ProjectID.String(),
			Type:            topic.Event.Type.String(),
			Subtype:         topic.Event.Subtype.String(),
			Status:          topic.EmployeeEventReviewers[0].ReviewerStatus.String(),
			IsRead:          topic.EmployeeEventReviewers[0].IsRead,
			EventReviewerID: topic.EmployeeEventReviewers[0].ReviewerID.String(),
			LastUpdated:     topic.EmployeeEventReviewers[0].UpdatedAt,
			Author:          toBasicEmployeeInfo(topic.Event.Employee),
		}

		if topic.Event.Type == model.EventTypeSurvey {
			feedback.Author = &DwarvesAuthor
		}

		results = append(results, feedback)
	}

	return results
}

type ListFeedbackResponse struct {
	Data []Feedback `json:"data"`
} // @name ListFeedbackResponse

type QuestionAnswer struct {
	EventQuestionID string               `json:"eventQuestionID"`
	Content         string               `json:"content"`
	Answer          string               `json:"answer"`
	Note            string               `json:"note"`
	Type            string               `json:"type"`
	Order           int64                `json:"order"`
	Domain          model.QuestionDomain `json:"domain"`
} // @name QuestionAnswer

type FeedBackReviewDetail struct {
	Questions    []QuestionAnswer  `json:"questions"`
	TopicName    string            `json:"topicName"`
	Relationship string            `json:"relationship"`
	Employee     BasicEmployeeInfo `json:"employee"`
	Reviewer     BasicEmployeeInfo `json:"reviewer"`
	Project      *BasicProjectInfo `json:"project"`
}

type FeedbackDetail struct {
	Answers      []*QuestionAnswer `json:"answers"`
	Status       string            `json:"status"`
	EmployeeID   string            `json:"employeeID"`
	Reviewer     BasicEmployeeInfo `json:"reviewer"`
	TopicID      string            `json:"topicID"`
	EventID      string            `json:"eventID"`
	Title        string            `json:"title"`
	Relationship string            `json:"relationship"`
	Project      *BasicProjectInfo `json:"project"`
} // @name FeedbackDetail

type FeedbackDetailResponse struct {
	Data FeedbackDetail `json:"data"`
} // @name FeedbackDetailResponse

type FeedbackDetailInfo struct {
	Status       model.EventReviewerStatus
	EmployeeID   string
	Reviewer     *model.Employee
	TopicID      string
	EventID      string
	Title        string
	Relationship model.Relationship
	Project      *model.Project
}

func ToListFeedbackDetails(questions []*model.EmployeeEventQuestion, detailInfo FeedbackDetailInfo) FeedbackDetail {
	var rs FeedbackDetail

	for _, q := range questions {
		if q.Type == model.QuestionTypeScale.String() {
			q.Answer = model.AgreementLevelValueMap[q.Answer].String()
		}

		rs.Answers = append(rs.Answers, &QuestionAnswer{
			EventQuestionID: q.ID.String(),
			Content:         q.Content,
			Answer:          q.Answer,
			Note:            q.Note,
			Type:            q.Type,
			Order:           q.Order,
			Domain:          q.Domain,
		})
	}

	rs.Reviewer = *toBasicEmployeeInfo(*detailInfo.Reviewer)

	if detailInfo.Project != nil {
		rs.Project = toBasicProjectInfo(*detailInfo.Project)
	}

	rs.Status = detailInfo.Status.String()
	rs.EmployeeID = detailInfo.EmployeeID
	rs.TopicID = detailInfo.TopicID
	rs.EventID = detailInfo.EventID
	rs.Title = detailInfo.Title
	rs.Relationship = detailInfo.Relationship.String()

	return rs
}

type SubmitFeedback struct {
	EventID    string            `json:"eventID"`
	TopicID    string            `json:"topicID"`
	EmployeeID string            `json:"employeeID"`
	Title      string            `json:"title"`
	Status     string            `json:"status"`
	Reviewer   BasicEmployeeInfo `json:"reviewer"`
	Answers    []*QuestionAnswer `json:"answers"`
} // @name SubmitFeedback

type SubmitFeedbackResponse struct {
	Data SubmitFeedback `json:"data"`
} // @name SubmitFeedbackResponse

func ToListSubmitFeedback(questions []*model.EmployeeEventQuestion, detailInfo FeedbackDetailInfo) SubmitFeedback {
	var rs SubmitFeedback

	for _, q := range questions {
		if q.Type == model.QuestionTypeScale.String() {
			q.Answer = model.AgreementLevelValueMap[q.Answer].String()
		}

		rs.Answers = append(rs.Answers, &QuestionAnswer{
			EventQuestionID: q.ID.String(),
			Content:         q.Content,
			Answer:          q.Answer,
			Note:            q.Note,
			Type:            q.Type,
			Order:           q.Order,
			Domain:          q.Domain,
		})
	}

	if detailInfo.Reviewer != nil {
		rs.Reviewer = *toBasicEmployeeInfo(*detailInfo.Reviewer)
	}

	rs.Status = detailInfo.Status.String()
	rs.EmployeeID = detailInfo.EmployeeID
	rs.TopicID = detailInfo.TopicID
	rs.EventID = detailInfo.EventID
	rs.Title = detailInfo.Title

	return rs
}

type FeedbackReviewDetailResponse struct {
	Data *FeedBackReviewDetail `json:"data"`
}

func ToFeedbackReviewDetail(questions []*model.EmployeeEventQuestion, topic *model.EmployeeEventTopic, reviewer *model.EmployeeEventReviewer, project *model.Project) FeedBackReviewDetail {
	var qs []QuestionAnswer

	for _, q := range questions {
		if q.Type == model.QuestionTypeScale.String() {
			q.Answer = model.AgreementLevelValueMap[q.Answer].String()
		}

		qs = append(qs, QuestionAnswer{
			EventQuestionID: q.ID.String(),
			Content:         q.Content,
			Answer:          q.Answer,
			Note:            q.Note,
			Type:            q.Type,
			Order:           q.Order,
			Domain:          q.Domain,
		})
	}

	rs := FeedBackReviewDetail{
		Questions:    qs,
		TopicName:    topic.Title,
		Relationship: reviewer.Relationship.String(),
	}

	if topic.Employee != nil {
		rs.Employee = *toBasicEmployeeInfo(*topic.Employee)
	}

	if reviewer.Reviewer != nil {
		rs.Reviewer = *toBasicEmployeeInfo(*reviewer.Reviewer)
	}

	if project != nil {
		rs.Project = toBasicProjectInfo(*project)
	}

	return rs
}

type UnreadFeedbackCountResponse struct {
	Data UnreadFeedbackCountData `json:"data"`
} // @name UnreadFeedbackCountResponse

type UnreadFeedbackCountData struct {
	Count      int64  `json:"count"`
	ReviewerID string `json:"reviewerID"`
} // @name UnreadFeedbackCountData

func ToUnreadFeedbackCountData(reviewerID string, count int64) UnreadFeedbackCountData {
	return UnreadFeedbackCountData{
		ReviewerID: reviewerID,
		Count:      count,
	}
}
