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
}

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
			Author:          ToBasicEmployeeInfo(topic.Event.Employee),
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
}

type QuestionAnswer struct {
	EventQuestionID string `json:"eventQuestionID"`
	Content         string `json:"content"`
	Answer          string `json:"answer"`
	Note            string `json:"note"`
	Type            string `json:"type"`
	Order           int64  `json:"order"`
}

type FeedBackReviewDetail struct {
	Questions    []QuestionAnswer  `json:"questions"`
	TopicName    string            `json:"topicName"`
	Relationship string            `json:"relationship"`
	Employee     BasicEmployeeInfo `json:"employee"`
	Reviewer     BasicEmployeeInfo `json:"reviewer"`
}

type FeedbackDetail struct {
	Answers    []*QuestionAnswer `json:"answers"`
	Status     string            `json:"status"`
	EmployeeID string            `json:"employeeID"`
	Reviewer   BasicEmployeeInfo `json:"reviewer"`
	TopicID    string            `json:"topicID"`
	EventID    string            `json:"eventID"`
	Title      string            `json:"title"`
}

type FeedbackDetailResponse struct {
	Data FeedbackDetail `json:"data"`
}

type FeedbackDetailInfo struct {
	Status     model.EventReviewerStatus
	EmployeeID string
	Reviewer   *model.Employee
	TopicID    string
	EventID    string
	Title      string
}

func ToListFeedbackDetails(questions []*model.EmployeeEventQuestion, detailInfo FeedbackDetailInfo) FeedbackDetail {
	var rs FeedbackDetail

	for _, q := range questions {
		rs.Answers = append(rs.Answers, &QuestionAnswer{
			EventQuestionID: q.ID.String(),
			Content:         q.Content,
			Answer:          q.Answer,
			Note:            q.Note,
			Type:            q.Type,
			Order:           q.Order,
		})
	}

	rs.Reviewer = BasicEmployeeInfo{
		ID:          detailInfo.Reviewer.ID.String(),
		FullName:    detailInfo.Reviewer.FullName,
		DisplayName: detailInfo.Reviewer.DisplayName,
		Avatar:      detailInfo.Reviewer.Avatar,
	}

	rs.Status = detailInfo.Status.String()
	rs.EmployeeID = detailInfo.EmployeeID
	rs.TopicID = detailInfo.TopicID
	rs.EventID = detailInfo.EventID
	rs.Title = detailInfo.Title

	return rs
}

type SubmitFeedback struct {
	Answers    []*QuestionAnswer `json:"answers"`
	Status     string            `json:"status"`
	EmployeeID string            `json:"employeeID"`
	Reviewer   BasicEmployeeInfo `json:"reviewer"`
	TopicID    string            `json:"topicID"`
	EventID    string            `json:"eventID"`
	Title      string            `json:"title"`
}

type SubmitFeedbackResponse struct {
	Data SubmitFeedback `json:"data"`
}

func ToListSubmitFeedback(questions []*model.EmployeeEventQuestion, detailInfo FeedbackDetailInfo) SubmitFeedback {
	var rs SubmitFeedback

	for _, q := range questions {
		rs.Answers = append(rs.Answers, &QuestionAnswer{
			EventQuestionID: q.ID.String(),
			Content:         q.Content,
			Answer:          q.Answer,
			Note:            q.Note,
			Type:            q.Type,
			Order:           q.Order,
		})
	}

	rs.Reviewer = BasicEmployeeInfo{
		ID:          detailInfo.Reviewer.ID.String(),
		FullName:    detailInfo.Reviewer.FullName,
		DisplayName: detailInfo.Reviewer.DisplayName,
		Avatar:      detailInfo.Reviewer.Avatar,
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

func ToFeedbackReviewDetail(questions []*model.EmployeeEventQuestion, topic *model.EmployeeEventTopic, reviewer *model.EmployeeEventReviewer) FeedBackReviewDetail {
	var qs []QuestionAnswer

	for _, q := range questions {
		qs = append(qs, QuestionAnswer{
			EventQuestionID: q.ID.String(),
			Content:         q.Content,
			Answer:          q.Answer,
			Note:            q.Note,
			Type:            q.Type,
			Order:           q.Order,
		})
	}

	return FeedBackReviewDetail{
		Questions:    qs,
		TopicName:    topic.Title,
		Relationship: reviewer.Relationship.String(),
		Employee: BasicEmployeeInfo{
			ID:          topic.EmployeeID.String(),
			FullName:    topic.Employee.FullName,
			DisplayName: topic.Employee.DisplayName,
			Avatar:      topic.Employee.Avatar,
		},
		Reviewer: BasicEmployeeInfo{
			ID:          reviewer.ReviewerID.String(),
			FullName:    reviewer.Reviewer.FullName,
			DisplayName: reviewer.Reviewer.DisplayName,
			Avatar:      reviewer.Reviewer.Avatar,
		},
	}
}

type Survey struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Type      string        `json:"type"`
	Subtype   string        `json:"subtype"`
	Status    string        `json:"status"`
	StartDate *time.Time    `json:"startDate"`
	EndDate   *time.Time    `json:"endDate"`
	Count     FeedbackCount `json:"count"`
}

type FeedbackCount struct {
	Total int64 `json:"total"`
	Sent  int64 `json:"sent"`
	Done  int64 `json:"done"`
}

func ToListSurvey(events []*model.FeedbackEvent) []Survey {
	var results = make([]Survey, 0, len(events))

	for _, e := range events {
		var done, total int64

		for _, topic := range e.Topics {
			total += int64(len(topic.EmployeeEventReviewers))
			for _, reviewer := range topic.EmployeeEventReviewers {
				if reviewer.AuthorStatus == model.EventAuthorStatusDone {
					done++
				}
			}
		}

		results = append(results, Survey{
			ID:        e.ID.String(),
			Title:     e.Title,
			Type:      e.Type.String(),
			Subtype:   e.Subtype.String(),
			Status:    e.Status.String(),
			StartDate: e.StartDate,
			EndDate:   e.EndDate,
			Count: FeedbackCount{
				Total: total,
				Sent:  total,
				Done:  done,
			},
		})
	}

	return results
}

type ListSurveyResponse struct {
	Data []Survey `json:"data"`
}

type SurveyDetail struct {
	EventID   string             `json:"eventID"`
	Title     string             `json:"title"`
	Type      string             `json:"type"`
	Subtype   string             `json:"subtype"`
	Status    string             `json:"status"`
	StartDate *time.Time         `json:"startDate"`
	EndDate   *time.Time         `json:"endDate"`
	Author    *BasicEmployeeInfo `json:"author"`
	Topics    []Topic            `json:"topics"`
}

type Topic struct {
	ID           string              `json:"id"`
	EventID      string              `json:"eventID"`
	Title        string              `json:"title"`
	Type         string              `json:"type"`
	Subtype      string              `json:"subtype"`
	Employee     BasicEmployeeInfo   `json:"employee"`
	Participants []BasicEmployeeInfo `json:"participants"`
	Count        FeedbackCount       `json:"count"`
}

func ToListSurveyDetail(event *model.FeedbackEvent) SurveyDetail {
	result := SurveyDetail{
		EventID:   event.ID.String(),
		Title:     event.Title,
		Type:      event.Type.String(),
		Subtype:   event.Subtype.String(),
		Status:    event.Status.String(),
		StartDate: event.StartDate,
		EndDate:   event.EndDate,
		Author:    ToBasicEmployeeInfo(event.Employee),
	}

	var topics = make([]Topic, 0, len(event.Topics))

	for _, topic := range event.Topics {
		total := int64(len(topic.EmployeeEventReviewers))
		var sent, done int64

		for _, reviewer := range topic.EmployeeEventReviewers {
			switch reviewer.AuthorStatus {
			case model.EventAuthorStatusSent:
				sent++
			case model.EventAuthorStatusDone:
				done++
			}
		}

		participants := make([]BasicEmployeeInfo, 0, len(topic.EmployeeEventReviewers))
		for _, reviewer := range topic.EmployeeEventReviewers {
			employee := ToBasicEmployeeInfo(*reviewer.Reviewer)
			participants = append(participants, *employee)
		}

		topics = append(topics, Topic{
			ID:       topic.ID.String(),
			Title:    topic.Title,
			Type:     topic.Event.Type.String(),
			Subtype:  topic.Event.Subtype.String(),
			Employee: *ToBasicEmployeeInfo(*topic.Employee),
			Count: FeedbackCount{
				Total: total,
				Sent:  sent,
				Done:  done,
			},
			Participants: participants,
		})
	}

	result.Topics = topics
	return result
}

type ListSurveyDetailResponse struct {
	Data SurveyDetail `json:"data"`
}
