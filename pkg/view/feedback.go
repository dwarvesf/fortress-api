package view

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

var DwarvesAuthor = BasisEmployeeInfo{
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
	Author          *BasisEmployeeInfo `json:"author"`
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
			Status:          topic.EmployeeEventReviewers[0].Status.String(),
			IsRead:          topic.EmployeeEventReviewers[0].IsRead,
			EventReviewerID: topic.EmployeeEventReviewers[0].ReviewerID.String(),
			LastUpdated:     topic.EmployeeEventReviewers[0].UpdatedAt,
			Author:          ToBasisEmployeeInfo(topic.Event.Employee),
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

type FeedbackDetail struct {
	Answers    []*QuestionAnswer `json:"answers"`
	Status     string            `json:"status"`
	EmployeeID string            `json:"employeeID"`
	ReviewerID string            `json:"reviewerID"`
	TopicID    string            `json:"topicID"`
	EventID    string            `json:"eventID"`
}

type FeedbackDetailResponse struct {
	Data FeedbackDetail `json:"data"`
}

type FeedbackDetailInfo struct {
	Status     model.EventReviewerStatus
	EmployeeID string
	ReviewerID string
	TopicID    string
	EventID    string
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

	rs.Status = detailInfo.Status.String()
	rs.EmployeeID = detailInfo.EmployeeID
	rs.ReviewerID = detailInfo.ReviewerID
	rs.TopicID = detailInfo.TopicID
	rs.EventID = detailInfo.EventID

	return rs
}

type SubmitFeedback struct {
	Answers    []*QuestionAnswer `json:"answers"`
	Status     string            `json:"status"`
	EmployeeID string            `json:"employeeID"`
	ReviewerID string            `json:"reviewerID"`
	TopicID    string            `json:"topicID"`
	EventID    string            `json:"eventID"`
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

	rs.Status = detailInfo.Status.String()
	rs.EmployeeID = detailInfo.EmployeeID
	rs.ReviewerID = detailInfo.ReviewerID
	rs.TopicID = detailInfo.TopicID
	rs.EventID = detailInfo.EventID

	return rs
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
				if reviewer.Status == model.EventReviewerStatusDone {
					done++
				}
			}
		}

		results = append(results, Survey{
			ID:        e.ID.String(),
			Title:     e.Title,
			Type:      e.Type.String(),
			Subtype:   e.Subtype.String(),
			Status:    e.Status,
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
	TopicID      string              `json:"topicID"`
	EventID      string              `json:"eventID"`
	Title        string              `json:"title"`
	Type         string              `json:"type"`
	Subtype      string              `json:"subtype"`
	Employee     BasisEmployeeInfo   `json:"employee"`
	Participants []BasisEmployeeInfo `json:"participants"`
	Count        FeedbackCount       `json:"count"`
}

func ToListSurveyDetail(topics []*model.EmployeeEventTopic) []SurveyDetail {
	var results = make([]SurveyDetail, 0, len(topics))

	for _, topic := range topics {
		total := int64(len(topic.EmployeeEventReviewers))
		var done int64

		for _, reviewer := range topic.EmployeeEventReviewers {
			if reviewer.Status == model.EventReviewerStatusDone {
				done++
			}
		}

		participants := make([]BasisEmployeeInfo, 0, len(topic.EmployeeEventReviewers))
		for _, reviewer := range topic.EmployeeEventReviewers {
			employee := ToBasisEmployeeInfo(*reviewer.Reviewer)
			participants = append(participants, *employee)
		}

		results = append(results, SurveyDetail{
			TopicID:  topic.ID.String(),
			EventID:  topic.EventID.String(),
			Title:    topic.Title,
			Type:     topic.Event.Type.String(),
			Subtype:  topic.Event.Subtype.String(),
			Employee: *ToBasisEmployeeInfo(*topic.Employee),
			Count: FeedbackCount{
				Total: total,
				Sent:  total,
				Done:  done,
			},
			Participants: participants,
		})
	}

	return results
}

type ListSurveyDetailResponse struct {
	Data []SurveyDetail `json:"data"`
}
