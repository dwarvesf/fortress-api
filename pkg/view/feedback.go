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
	Answers      []*QuestionAnswer `json:"answers"`
	Status       string            `json:"status"`
	EmployeeID   string            `json:"employeeID"`
	Reviewer     BasicEmployeeInfo `json:"reviewer"`
	TopicID      string            `json:"topicID"`
	EventID      string            `json:"eventID"`
	Title        string            `json:"title"`
	Relationship string            `json:"relationship"`
}

type FeedbackDetailResponse struct {
	Data FeedbackDetail `json:"data"`
}

type FeedbackDetailInfo struct {
	Status       model.EventReviewerStatus
	EmployeeID   string
	Reviewer     *model.Employee
	TopicID      string
	EventID      string
	Title        string
	Relationship model.Relationship
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
	rs.Relationship = detailInfo.Relationship.String()

	return rs
}

type PeerReviewer struct {
	EventReviewerID string                    `json:"eventReviewerID"`
	Reviewer        *BasicEmployeeInfo        `json:"reviewer"`
	Status          model.EventReviewerStatus `json:"status"`
	Relationship    model.Relationship        `json:"relationship"`
}

type SurveyTopicDetail struct {
	TopicID      string             `json:"topicID"`
	Title        string             `json:"title"`
	Employee     *BasicEmployeeInfo `json:"employee"`
	Participants []PeerReviewer     `json:"participants"`
}

type SurveyTopicDetailResponse struct {
	Data SurveyTopicDetail `json:"data"`
}

func ToPeerReviewDetail(topic *model.EmployeeEventTopic) SurveyTopicDetail {
	rs := SurveyTopicDetail{
		TopicID:  topic.ID.String(),
		Employee: ToBasicEmployeeInfo(*topic.Employee),
		Title:    topic.Title,
	}

	for _, eventReviewer := range topic.EmployeeEventReviewers {
		participant := PeerReviewer{
			EventReviewerID: eventReviewer.ID.String(),
			Status:          model.EventReviewerStatus(eventReviewer.AuthorStatus),
			Relationship:    eventReviewer.Relationship,
		}

		if eventReviewer.Reviewer != nil {
			participant.Reviewer = ToBasicEmployeeInfo(*eventReviewer.Reviewer)
		}

		rs.Participants = append(rs.Participants, participant)

	}

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
	Total int `json:"total"`
	Sent  int `json:"sent"`
	Done  int `json:"done"`
}

func ToListSurvey(events []*model.FeedbackEvent) []Survey {
	var results = make([]Survey, 0, len(events))

	for _, e := range events {
		var sent, done int

		for _, topic := range e.Topics {
			var topicSent, topicDone int

			for _, reviewer := range topic.EmployeeEventReviewers {
				if reviewer.AuthorStatus != model.EventAuthorStatusDraft {
					topicSent++
				}
				if reviewer.AuthorStatus == model.EventAuthorStatusDone {
					topicDone++
				}
			}

			if topicSent == len(topic.EmployeeEventReviewers) {
				sent++
			}

			if topicDone == len(topic.EmployeeEventReviewers) {
				done++
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
				Total: len(e.Topics),
				Sent:  sent,
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
	ReviewID     string              `json:"reviewID,omitempty"`
	Title        string              `json:"title"`
	Type         string              `json:"type"`
	Subtype      string              `json:"subtype"`
	Status       string              `json:"status,omitempty"`
	Employee     BasicEmployeeInfo   `json:"employee"`
	Participants []BasicEmployeeInfo `json:"participants"`
	Count        *FeedbackCount      `json:"count"`
}

func ToSurveyDetail(event *model.FeedbackEvent) SurveyDetail {
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
		newTopic := Topic{
			ID:       topic.ID.String(),
			EventID:  topic.EventID.String(),
			Title:    topic.Title,
			Type:     topic.Event.Type.String(),
			Subtype:  topic.Event.Subtype.String(),
			Employee: *ToBasicEmployeeInfo(*topic.Employee),
		}

		// just use for peer-review survey
		if topic.Event.Subtype == model.EventSubtypePeerReview {
			participants := make([]BasicEmployeeInfo, 0, len(topic.EmployeeEventReviewers))
			for _, reviewer := range topic.EmployeeEventReviewers {
				employee := ToBasicEmployeeInfo(*reviewer.Reviewer)
				participants = append(participants, *employee)
			}

			var sent, done int
			for _, reviewer := range topic.EmployeeEventReviewers {
				if reviewer.AuthorStatus != model.EventAuthorStatusDraft {
					sent++
				}
				if reviewer.AuthorStatus == model.EventAuthorStatusDone {
					done++
				}
			}

			newTopic.Participants = participants
			newTopic.Count = &FeedbackCount{
				Total: len(topic.EmployeeEventReviewers),
				Sent:  sent,
				Done:  done,
			}
		}

		// just use for engagement survey
		if event.Subtype == model.EventSubtypeEngagement && len(topic.EmployeeEventReviewers) > 0 {
			newTopic.ReviewID = topic.EmployeeEventReviewers[0].ID.String()
			newTopic.Status = topic.EmployeeEventReviewers[0].AuthorStatus.String()
		}

		topics = append(topics, newTopic)
	}

	result.Topics = topics
	return result
}

type ListSurveyDetailResponse struct {
	Data SurveyDetail `json:"data"`
}
