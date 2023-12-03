package view

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type Survey struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Type      string        `json:"type"`
	Subtype   string        `json:"subtype"`
	Status    string        `json:"status"`
	StartDate *time.Time    `json:"startDate"`
	EndDate   *time.Time    `json:"endDate"`
	Count     FeedbackCount `json:"count"`
	Domains   []Domain      `json:"domains"`
} // @name Survey

type FeedbackCount struct {
	Total int `json:"total"`
	Sent  int `json:"sent"`
	Done  int `json:"done"`
} // @name FeedbackCount

type Domain struct {
	Name    string           `json:"name"`
	Average float32          `json:"average"`
	Count   LikertScaleCount `json:"count"`
} // @name Domain

type LikertScaleCount struct {
	StronglyDisagree int `json:"stronglyDisagree" gorm:"column:strongly_disagree"`
	Disagree         int `json:"disagree" gorm:"column:disagree"`
	Mixed            int `json:"mixed" gorm:"column:mixed"`
	Agree            int `json:"agree" gorm:"column:agree"`
	StronglyAgree    int `json:"stronglyAgree" gorm:"column:strongly_agree"`
} // @name LikertScaleCount

func ToListSurvey(events []*model.FeedbackEvent) []Survey {
	var results = make([]Survey, 0, len(events))

	for _, e := range events {
		var sent, done int

		// calculate feedback count
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

		// calculate domain value
		// average = (count1 * weight1 + count2 * weight2 + ...) / (count1 + count2 + ...)
		var domains []Domain
		if e.Subtype == model.EventSubtypeWork && len(e.QuestionDomainCounts) > 0 {
			domains = make([]Domain, 0)
			for _, count := range e.QuestionDomainCounts {
				var average float32
				total := count.StronglyDisagree +
					count.Disagree +
					count.Mixed +
					count.Agree +
					count.StronglyAgree

				if total > 0 {
					average = float32(count.StronglyDisagree+
						count.Disagree*2+
						count.Mixed*3+
						count.Agree*4+
						count.StronglyAgree*5) / float32(total)
				}

				domains = append(domains, Domain{
					Name:    count.Domain.String(),
					Average: average,
					Count: LikertScaleCount{
						StronglyDisagree: count.StronglyDisagree,
						Disagree:         count.Disagree,
						Mixed:            count.Mixed,
						Agree:            count.Agree,
						StronglyAgree:    count.StronglyAgree,
					},
				})
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
			Domains: domains,
		})
	}

	return results
}

type ListSurveyResponse struct {
	Data []Survey `json:"data"`
} // @name ListSurveyResponse

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
} // @name SurveyDetail

type Topic struct {
	ID           string              `json:"id"`
	EventID      string              `json:"eventID"`
	ReviewID     string              `json:"reviewID,omitempty"`
	Title        string              `json:"title"`
	Type         string              `json:"type"`
	Subtype      string              `json:"subtype"`
	Status       string              `json:"status,omitempty"`
	IsForcedDone bool                `json:"isForcedDone"`
	Employee     BasicEmployeeInfo   `json:"employee"`
	Participants []BasicEmployeeInfo `json:"participants"`
	Count        *FeedbackCount      `json:"count"`
	Domains      []Domain            `json:"domains"`
	Comments     int                 `json:"comments"`
	Project      *BasicProjectInfo   `json:"project"`
} // @name Topic

type SurveyResult struct {
	StronglyDisagree int `json:"stronglyDisagree"`
	Disagree         int `json:"disagree"`
	Mixed            int `json:"mixed"`
	Agree            int `json:"agree"`
	StronglyAgree    int `json:"stronglyAgree"`
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
		Author:    toBasicEmployeeInfo(event.Employee),
	}

	var topics = make([]Topic, 0, len(event.Topics))

	for _, topic := range event.Topics {
		newTopic := Topic{
			ID:       topic.ID.String(),
			EventID:  topic.EventID.String(),
			Title:    topic.Title,
			Type:     topic.Event.Type.String(),
			Subtype:  topic.Event.Subtype.String(),
			Employee: *toBasicEmployeeInfo(*topic.Employee),
		}

		switch topic.Event.Subtype {
		case model.EventSubtypePeerReview:
			{
				participants := make([]BasicEmployeeInfo, 0, len(topic.EmployeeEventReviewers))
				for _, reviewer := range topic.EmployeeEventReviewers {
					employee := toBasicEmployeeInfo(*reviewer.Reviewer)
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
		case model.EventSubtypeWork:
			{
				// only 1 reviewer for each topic
				if len(topic.EmployeeEventReviewers) > 0 {
					newTopic.ReviewID = topic.EmployeeEventReviewers[0].ID.String()
					newTopic.Status = topic.EmployeeEventReviewers[0].AuthorStatus.String()
					newTopic.IsForcedDone = topic.EmployeeEventReviewers[0].IsForcedDone

					if topic.Project != nil {
						newTopic.Project = toBasicProjectInfo(*topic.Project)
					}

					totalComment := 0
					answerMap := make(map[string]int)

					for _, question := range topic.EmployeeEventReviewers[0].EmployeeEventQuestions {
						if question.Note != "" {
							totalComment++
						}

						answerMap[question.Answer]++
					}

					newTopic.Comments = totalComment
					newTopic.Domains = toDomains(topic.EmployeeEventReviewers[0].EmployeeEventQuestions)
				}
			}
		case model.EventSubtypeEngagement:
			{
				// only 1 reviewer for each topic
				if len(topic.EmployeeEventReviewers) > 0 {
					newTopic.ReviewID = topic.EmployeeEventReviewers[0].ID.String()
					newTopic.Status = topic.EmployeeEventReviewers[0].AuthorStatus.String()
					newTopic.IsForcedDone = topic.EmployeeEventReviewers[0].IsForcedDone
				}
			}
		}

		topics = append(topics, newTopic)
	}

	result.Topics = topics
	return result
}

type PeerReviewer struct {
	EventReviewerID string              `json:"eventReviewerID"`
	Reviewer        *BasicEmployeeInfo  `json:"reviewer"`
	Status          EventReviewerStatus `json:"status"`
	Relationship    Relationship        `json:"relationship"`
	IsForcedDone    bool                `json:"isForcedDone"`
} // @name PeerReviewer

// EventReviewerStatus event_reviewer_status for table employee event reviewer
type EventReviewerStatus string // @name EventReviewerStatus

// EventReviewerStatus values
const (
	EventReviewerStatusNone  EventReviewerStatus = "none"
	EventReviewerStatusNew   EventReviewerStatus = "new"
	EventReviewerStatusDraft EventReviewerStatus = "draft"
	EventReviewerStatusDone  EventReviewerStatus = "done"
)

// IsValid validation for EventReviewerStatus
func (e EventReviewerStatus) IsValid() bool {
	switch e {
	case
		EventReviewerStatusNone,
		EventReviewerStatusDraft,
		EventReviewerStatusDone,
		EventReviewerStatusNew:
		return true
	}
	return false
}

// String returns the string type from the EventReviewerStatus type
func (e EventReviewerStatus) String() string {
	return string(e)
}

// Relationship relationships for table employee_event_topics
type Relationship string // @name Relationship

// values for Relationship
const (
	RelationshipPeer        Relationship = "peer"
	RelationshipLineManager Relationship = "line-manager"
	RelationshipChapterLead Relationship = "chapter-lead"
	RelationshipSelf        Relationship = "self"
)

// IsValid validation for Relationship
func (e Relationship) IsValid() bool {
	switch e {
	case
		RelationshipPeer,
		RelationshipLineManager,
		RelationshipChapterLead,
		RelationshipSelf:
		return true
	}
	return false
}

// IsValid validation for Relationship
func (e Relationship) String() string {
	return string(e)
}

type SurveyTopicDetail struct {
	TopicID      string             `json:"topicID"`
	Title        string             `json:"title"`
	Employee     *BasicEmployeeInfo `json:"employee"`
	Participants []PeerReviewer     `json:"participants"`
} // @name SurveyTopicDetail

type SurveyTopicDetailResponse struct {
	Data SurveyTopicDetail `json:"data"`
} // @name SurveyTopicDetailResponse

func ToPeerReviewDetail(topic *model.EmployeeEventTopic) SurveyTopicDetail {
	rs := SurveyTopicDetail{
		TopicID:  topic.ID.String(),
		Employee: toBasicEmployeeInfo(*topic.Employee),
		Title:    topic.Title,
	}

	for _, eventReviewer := range topic.EmployeeEventReviewers {
		participant := PeerReviewer{
			EventReviewerID: eventReviewer.ID.String(),
			Status:          EventReviewerStatus(eventReviewer.AuthorStatus),
			Relationship:    Relationship(eventReviewer.Relationship),
			IsForcedDone:    eventReviewer.IsForcedDone,
		}

		if eventReviewer.Reviewer != nil {
			participant.Reviewer = toBasicEmployeeInfo(*eventReviewer.Reviewer)
		}

		rs.Participants = append(rs.Participants, participant)
	}

	return rs
}

type ListSurveyDetailResponse struct {
	Data SurveyDetail `json:"data"`
} // @name ListSurveyDetailResponse

func toDomains(questions []model.EmployeeEventQuestion) []Domain {
	domainMap := make(map[string]map[string]int)
	domainMap[""] = make(map[string]int)
	domainMap[model.QuestionDomainWorkload.String()] = make(map[string]int)
	domainMap[model.QuestionDomainDeadline.String()] = make(map[string]int)
	domainMap[model.QuestionDomainLearning.String()] = make(map[string]int)

	for _, q := range questions {
		domainMap[q.Domain.String()][model.AgreementLevelValueMap[q.Answer].String()]++
	}

	wlCount := model.LikertScaleCount{
		StronglyDisagree: domainMap[model.QuestionDomainWorkload.String()][model.AgreementLevelStronglyDisagree.String()],
		Disagree:         domainMap[model.QuestionDomainWorkload.String()][model.AgreementLevelDisagree.String()],
		Mixed:            domainMap[model.QuestionDomainWorkload.String()][model.AgreementLevelMixed.String()],
		Agree:            domainMap[model.QuestionDomainWorkload.String()][model.AgreementLevelAgree.String()],
		StronglyAgree:    domainMap[model.QuestionDomainWorkload.String()][model.AgreementLevelStronglyAgree.String()],
	}
	dlCount := model.LikertScaleCount{
		StronglyDisagree: domainMap[model.QuestionDomainDeadline.String()][model.AgreementLevelStronglyDisagree.String()],
		Disagree:         domainMap[model.QuestionDomainDeadline.String()][model.AgreementLevelDisagree.String()],
		Mixed:            domainMap[model.QuestionDomainDeadline.String()][model.AgreementLevelMixed.String()],
		Agree:            domainMap[model.QuestionDomainDeadline.String()][model.AgreementLevelAgree.String()],
		StronglyAgree:    domainMap[model.QuestionDomainDeadline.String()][model.AgreementLevelStronglyAgree.String()],
	}
	lnCount := model.LikertScaleCount{
		StronglyDisagree: domainMap[model.QuestionDomainLearning.String()][model.AgreementLevelStronglyDisagree.String()],
		Disagree:         domainMap[model.QuestionDomainLearning.String()][model.AgreementLevelDisagree.String()],
		Mixed:            domainMap[model.QuestionDomainLearning.String()][model.AgreementLevelMixed.String()],
		Agree:            domainMap[model.QuestionDomainLearning.String()][model.AgreementLevelAgree.String()],
		StronglyAgree:    domainMap[model.QuestionDomainLearning.String()][model.AgreementLevelStronglyAgree.String()],
	}

	wlDomain := Domain{
		Name:    model.QuestionDomainWorkload.String(),
		Average: countLikertScaleAverage(wlCount),
		Count: LikertScaleCount{
			StronglyDisagree: wlCount.StronglyDisagree,
			Disagree:         wlCount.Disagree,
			Mixed:            wlCount.Mixed,
			Agree:            wlCount.Agree,
			StronglyAgree:    wlCount.StronglyAgree,
		},
	}
	dlDomain := Domain{
		Name:    model.QuestionDomainDeadline.String(),
		Average: countLikertScaleAverage(dlCount),
		Count: LikertScaleCount{
			StronglyDisagree: dlCount.StronglyDisagree,
			Disagree:         dlCount.Disagree,
			Mixed:            dlCount.Mixed,
			Agree:            dlCount.Agree,
			StronglyAgree:    dlCount.StronglyAgree,
		},
	}
	lnDomain := Domain{
		Name:    model.QuestionDomainLearning.String(),
		Average: countLikertScaleAverage(lnCount),
		Count: LikertScaleCount{
			StronglyDisagree: lnCount.StronglyDisagree,
			Disagree:         lnCount.Disagree,
			Mixed:            lnCount.Mixed,
			Agree:            lnCount.Agree,
			StronglyAgree:    lnCount.StronglyAgree,
		},
	}

	return []Domain{wlDomain, dlDomain, lnDomain}
}

func countLikertScaleAverage(count model.LikertScaleCount) float32 {
	var average float32
	total := count.StronglyDisagree +
		count.Disagree +
		count.Mixed +
		count.Agree +
		count.StronglyAgree

	if total > 0 {
		average = float32(count.StronglyDisagree+
			count.Disagree*2+
			count.Mixed*3+
			count.Agree*4+
			count.StronglyAgree*5) / float32(total)
	}
	return average
}
