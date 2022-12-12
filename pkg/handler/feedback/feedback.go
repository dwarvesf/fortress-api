package feedback

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/employeeeventquestion"
	"github.com/dwarvesf/fortress-api/pkg/store/employeeeventtopic"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
	repo    store.DBRepo
	config  *config.Config
}

// New returns a handler
func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		store:   store,
		repo:    repo,
		service: service,
		logger:  logger,
		config:  cfg,
	}
}

// List godoc
// @Summary Get list feedbacks
// @Description Get list feedbacks
// @Tags Feedback
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param status query string false "Status"
// @Param page query string false "Page"
// @Param size query string false "Size"
// @Success 200 {object} view.ListFeedbackResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /feedbacks [get]
func (h *handler) List(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	input := GetListFeedbackInput{}
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if err := input.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	input.Standardize()

	l := h.logger.Fields(logger.Fields{
		"handler": "feedback",
		"method":  "List",
		"userID":  userID,
		"input":   input,
	})

	rs, total, err := h.store.EmployeeEventTopic.GetByEmployeeIDWithPagination(h.repo.DB(),
		userID,
		employeeeventtopic.GetByEmployeeIDInput{Status: input.Status},
		input.Pagination)
	if err != nil {
		l.Error(err, "failed to get employee event topic by employeeID")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToListFeedback(rs),
		&view.PaginationResponse{Pagination: input.Pagination, Total: total}, nil, nil, ""))
}

// Detail godoc
// @Summary Get feedback detail for logged-in users
// @Description Get feedback detail for logged-in users
// @Tags Feedback
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Feedback Event ID"
// @Param topicID path string true "Employee Event Topic ID"
// @Success 200 {object} view.FeedbackDetailResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /feedbacks/:id/topics/:topicID [get]
func (h *handler) Detail(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	input := DetailInput{
		EventID: c.Param("id"),
		TopicID: c.Param("topicID"),
	}

	if err := input.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "feedback",
		"method":  "Detail",
		"input":   input,
	})

	// Check topic and feedback existence
	topic, err := h.store.EmployeeEventTopic.One(h.repo.DB(), input.TopicID, input.EventID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(ErrTopicNotFound, "topic not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrTopicNotFound, input, ""))
		return
	}

	if err != nil {
		l.Error(err, "failed when getting topic")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	eventReviewer, err := h.store.EmployeeEventReviewer.One(h.repo.DB(), userID, input.TopicID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(ErrEmployeeEventReviewerNotFound, "employee event reviewer not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrEmployeeEventReviewerNotFound, nil, ""))
		return
	}

	if err != nil {
		l.Error(err, "failed to get employee event reviewer")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	questions, err := h.store.EmployeeEventQuestion.GetByEventReviewerID(h.repo.DB(), eventReviewer.ID.String())
	if err != nil {
		l.Error(err, "failed to get employee event question by reviewer")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	reviewer, err := h.store.Employee.One(h.repo.DB(), userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(ErrReviewerNotFound, "reviewer not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrReviewerNotFound, nil, ""))
		return
	}

	if err != nil {
		l.Error(err, "failed to get reviewer")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	detailInfo := view.FeedbackDetailInfo{
		Status:     eventReviewer.ReviewerStatus,
		EmployeeID: topic.EmployeeID.String(),
		Reviewer:   reviewer,
		TopicID:    input.TopicID,
		EventID:    input.EventID,
		Title:      topic.Title,
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToListFeedbackDetails(questions, detailInfo), nil, nil, nil, ""))
}

// ListSurvey godoc
// @Summary Get list event
// @Description Get list event
// @Tags Feedback
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param subtype query string true "Event Subtype"
// @Param page query string false "Page"
// @Param size query string false "Size"
// @Success 200 {object} view.ListSurveyResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /surveys [get]
func (h *handler) ListSurvey(c *gin.Context) {
	input := GetListSurveyInput{}
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if err := input.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	input.Standardize()

	l := h.logger.Fields(logger.Fields{
		"handler": "feedback",
		"method":  "ListSurvey",
		"input":   input,
	})

	events, total, err := h.store.FeedbackEvent.
		GetBySubtypeWithPagination(h.repo.DB(), input.Subtype, input.Pagination)
	if err != nil {
		l.Error(err, "failed to get feedback events by subtype")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToListSurvey(events),
		&view.PaginationResponse{Pagination: input.Pagination, Total: total}, nil, nil, ""))
}

// GetSurveyDetail godoc
// @Summary Get survey detail
// @Description Get survey detail
// @Tags Feedback
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param page query string false "Page"
// @Param size query string false "Size"
// @Success 200 {object} view.ListSurveyDetailResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /surveys/{id} [get]
func (h *handler) GetSurveyDetail(c *gin.Context) {
	input := GetSurveyDetailInput{
		EventID: c.Param("id"),
	}
	if err := c.ShouldBindQuery(&input.Pagination); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if err := input.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	input.Pagination.Standardize()

	l := h.logger.Fields(logger.Fields{
		"handler": "feedback",
		"method":  "GetSurveyDetail",
		"input":   input,
	})

	// check feedback event existence
	event, err := h.store.FeedbackEvent.One(h.repo.DB(), input.EventID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(err, "event not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrEventNotFound, input, ""))
		return
	}
	if err != nil {
		l.Error(err, "failed to get feedback event")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	topics, total, err := h.store.EmployeeEventTopic.GetByEventIDWithPagination(h.repo.DB(), input.EventID, input.Pagination)
	if err != nil {
		l.Error(err, "failed to get employee event topic by eventID")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	event.Topics = topics

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToListSurveyDetail(event),
		&view.PaginationResponse{Pagination: input.Pagination, Total: total}, nil, nil, ""))
}

// Submit godoc
// @Summary Submit the draft or done answers
// @Description Submit the draft or done answers
// @Tags Feedback
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Feedback Event ID"
// @Param topicID path string true "Employee Event Topic ID"
// @Param Body body SubmitBody true "Body"
// @Success 200 {object} view.SubmitFeedbackResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /feedbacks/{id}/topics/{topicID}/submit [post]
func (h *handler) Submit(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	input := SubmitInput{
		EventID: c.Param("id"),
		TopicID: c.Param("topicID"),
	}

	if err := c.ShouldBindJSON(&input.Body); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if err := input.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "feedback",
		"method":  "Submit",
		"userID":  userID,
		"input":   input,
	})

	// Begin transaction
	tx, done := h.repo.NewTransaction()

	// Check topic existence and validate eventID
	topic, err := h.store.EmployeeEventTopic.One(tx.DB(), input.TopicID, input.EventID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(ErrTopicNotFound, "topic not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, done(ErrTopicNotFound), input, ""))
		return
	}

	if err != nil {
		l.Error(err, "failed when getting topic")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	eventReviewer, err := h.store.EmployeeEventReviewer.One(tx.DB(), userID, input.TopicID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(ErrEventReviewerNotFound, "employee event reviewer not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, done(ErrEventReviewerNotFound), input, ""))
		return
	}

	if err != nil {
		l.Error(err, "failed to get employee event reviewer record")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	if eventReviewer.ReviewerStatus == model.EventReviewerStatusDone {
		l.Error(ErrCouldNotEditDoneFeedback, "could not edit the feedback marked as done")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(ErrCouldNotEditDoneFeedback), nil, ""))
		return
	}

	// check questionID existence
	eventQuestions, err := h.store.EmployeeEventQuestion.GetByEventReviewerID(tx.DB(), eventReviewer.ID.String())
	if err != nil {
		l.Error(err, "failed to validate questionID")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	questionMap := model.ToQuestionMap(eventQuestions)
	for _, e := range input.Body.Answers {
		_, ok := questionMap[e.EventQuestionID]
		if !ok {
			l.Error(errEventQuestionNotFound(e.EventQuestionID.String()), "employee event question not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, done(errEventQuestionNotFound(e.EventQuestionID.String())), input, ""))
			return
		}
	}

	// Update answers in employee_event_questions table
	for _, e := range input.Body.Answers {
		data := employeeeventquestion.BasicEventQuestion{
			EventQuestionID: e.EventQuestionID.String(),
			Answer:          e.Answer,
			Note:            e.Note,
		}

		if err := h.store.EmployeeEventQuestion.UpdateAnswers(tx.DB(), data); err != nil {
			l.Error(err, "failed to update employee event question")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
			return
		}
	}

	// Update status in employee_event_reviewers table
	eventReviewer.ReviewerStatus = input.Body.Status
	if input.Body.Status == model.EventReviewerStatusDone {
		eventReviewer.AuthorStatus = model.EventAuthorStatusDone
	}
	_, err = h.store.EmployeeEventReviewer.UpdateSelectedFieldsByID(tx.DB(), eventReviewer.ID.String(), *eventReviewer, "reviewer_status", "author_status")
	if err != nil {
		l.Error(err, "failed to update employee event question")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	eventQuestions, err = h.store.EmployeeEventQuestion.GetByEventReviewerID(tx.DB(), eventReviewer.ID.String())
	if err != nil {
		l.Error(err, "failed to get all empoyee event questions by event reviewer")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	if input.Body.Status == model.EventReviewerStatusDone {
		for _, e := range eventQuestions {
			if e.Answer == "" {
				l.Error(ErrUnansweredquestions, "there are some unanswered questions")
				c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, done(ErrUnansweredquestions), input, ""))
				return
			}
		}
	}

	reviewer, err := h.store.Employee.One(h.repo.DB(), userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(ErrReviewerNotFound, "reviewer not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrReviewerNotFound, nil, ""))
		return
	}

	if err != nil {
		l.Error(err, "failed to get reviewer")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	detailInfo := view.FeedbackDetailInfo{
		Status:     eventReviewer.ReviewerStatus,
		EmployeeID: topic.EmployeeID.String(),
		Reviewer:   reviewer,
		TopicID:    input.TopicID,
		EventID:    input.EventID,
		Title:      topic.Title,
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToListSubmitFeedback(eventQuestions, detailInfo), nil, nil, done(nil), ""))
}

// CreateSurvey godoc
// @Summary Create new survey
// @Description Create new survey
// @Tags Feedback
// @Accept  json
// @Produce  json
// @Param Body body CreateSurveyFeedbackInput true "Body"
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /surveys [post]
func (h *handler) CreateSurvey(c *gin.Context) {
	// 1. parse request
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	var req CreateSurveyFeedbackInput

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "feedback",
		"method":  "CreateSurvey",
		"input":   req,
	})

	if !model.EventSubtype(req.Type).IsValidSurvey() {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrInvalidEventSubType, req, ""))
		return
	}

	tx, done := h.repo.NewTransaction()

	if model.EventSubtype(req.Type) == model.EventSubtypePeerReview {
		status, err := h.createPeerReview(tx.DB(), req, userID)
		if err != nil {
			l.Error(err, "failed to create new survet")
			c.JSON(status, view.CreateResponse[any](nil, nil, done(err), nil, ""))
			return
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, done(nil), nil, "success"))

}

func (h *handler) createPeerReview(db *gorm.DB, req CreateSurveyFeedbackInput, userID string) (int, error) {
	//1. convert data
	var startTime, endTime time.Time
	var title string

	switch strings.ToLower(strings.ReplaceAll(req.Quarter, " ", "")) {
	case "q1,q2":
		startTime = time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC)
		endTime = time.Date(req.Year, 6, 30, 23, 59, 59, 59, time.UTC)
		title = fmt.Sprintf("Q1/Q2, %d", req.Year)
	case "q3,q4":
		startTime = time.Date(req.Year, 7, 1, 0, 0, 0, 0, time.UTC)
		endTime = time.Date(req.Year, 12, 31, 23, 59, 59, 59, time.UTC)
		title = fmt.Sprintf("Q3/Q4, %d", req.Year)
	default:
		return 400, ErrInvalidQuarter
	}

	//1.2 check employee existed
	createdBy, err := h.store.Employee.One(db, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 404, ErrEmployeeNotFound
		}
		return 500, err
	}

	//2. Create FeedbackEvent
	event, err := h.store.FeedbackEvent.Create(db, &model.FeedbackEvent{
		BaseModel: model.BaseModel{
			ID: model.NewUUID(),
		},
		Title:     title,
		Type:      model.EventTypeSurvey,
		Subtype:   model.EventSubtype(req.Type),
		Status:    model.EventStatusDraft,
		CreatedBy: createdBy.ID,
		StartDate: &startTime,
		EndDate:   &endTime,
	})
	if err != nil {
		return 500, err
	}

	//3. create EmployeeEventTopic
	employees, err := h.store.Employee.GetByWorkingStatus(db, model.WorkingStatusFullTime)
	if err != nil {
		return 500, err
	}

	eets := make([]model.EmployeeEventTopic, 0)
	for _, e := range employees {
		eets = append(eets, model.EmployeeEventTopic{
			BaseModel: model.BaseModel{
				ID: model.NewUUID(),
			},
			Title:      title,
			EventID:    event.ID,
			EmployeeID: e.ID,
		})
	}

	i := 0
	for i < len(eets) {
		to := i + 100
		if to > len(eets) {
			to = len(eets)
		}
		_, err = h.store.EmployeeEventTopic.BatchCreate(db, eets[i:to])
		if err != nil {
			return 500, err
		}
		i = to
	}

	//4. create EmployeeEventReviewer
	employeeEventMapper := make(map[model.UUID]model.UUID)
	for _, e := range eets {
		employeeEventMapper[e.EmployeeID] = e.ID
	}

	reviewers := make([]model.EmployeeEventReviewer, 0)

	for i, e := range eets {
		if !employees[i].LineManagerID.IsZero() {
			reviewers = append(reviewers, model.EmployeeEventReviewer{
				BaseModel: model.BaseModel{
					ID: model.NewUUID(),
				},
				EventID:              event.ID,
				EmployeeEventTopicID: e.ID,
				ReviewerID:           employees[i].LineManagerID,
				Relationship:         model.RelationshipLineManager,
				AuthorStatus:         model.EventAuthorStatusDraft,
				ReviewerStatus:       model.EventReviewerStatusNone,
				IsShared:             false,
				IsRead:               false,
			})
		}
	}

	peers, err := h.store.WorkUnitMember.GetPeerReviewerInTimeRange(db, &startTime, &endTime)
	if err != nil {
		return 500, err
	}

	for _, p := range peers {
		if !p.ReviewerID.IsZero() {
			reviewers = append(reviewers, model.EmployeeEventReviewer{
				BaseModel: model.BaseModel{
					ID: model.NewUUID(),
				},
				EventID:              event.ID,
				EmployeeEventTopicID: employeeEventMapper[p.EmployeeID],
				ReviewerID:           p.ReviewerID,
				Relationship:         model.RelationshipPeer,
				AuthorStatus:         model.EventAuthorStatusDraft,
				ReviewerStatus:       model.EventReviewerStatusNone,
				IsShared:             false,
				IsRead:               false,
			})
		}
	}

	i = 0
	for i < len(reviewers) {
		to := i + 100
		if to > len(reviewers) {
			to = len(reviewers)
		}
		_, err = h.store.EmployeeEventReviewer.BatchCreate(db, reviewers[i:to])
		if err != nil {
			return 500, err
		}
		i = to
	}

	//4. create EmployeeEventQuestion
	questions, err := h.store.Question.AllByCategory(db, model.EventTypeSurvey, model.EventSubtypePeerReview)
	if err != nil {
		return 500, err
	}

	eventQuestions := make([]model.EmployeeEventQuestion, 0)

	for _, r := range reviewers {
		for _, q := range questions {
			eventQuestions = append(eventQuestions, model.EmployeeEventQuestion{
				BaseModel: model.BaseModel{
					ID: model.NewUUID(),
				},
				EmployeeEventReviewerID: r.ID,
				QuestionID:              q.ID,
				EventID:                 r.EventID,
				Content:                 q.Content,
				Type:                    q.Type.String(),
				Order:                   q.Order,
			})
		}
	}

	i = 0
	for i < len(eventQuestions) {
		to := i + 100
		if to > len(eventQuestions) {
			to = len(eventQuestions)
		}
		_, err = h.store.EmployeeEventQuestion.BatchCreate(db, eventQuestions[i:to])
		if err != nil {
			return 500, err
		}
		i = to
	}

	return 200, nil
}

// SendPerformmentReview godoc
// @Summary Send the performance review
// @Description Send the performance review
// @Tags Feedback
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Feedback Event ID"
// @Param Body body PerformanceReviewListInput true "Body"
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /surveys/{id} [post]
func (h *handler) SendPerformmentReview(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" || !model.IsUUIDFromString(eventID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrInvalidEventID, nil, ""))
		return
	}

	var input PerformanceReviewListInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "feedback",
		"method":  "SendPerformmentReview",
		"eventID": eventID,
		"input":   input,
	})

	// Begin transaction
	tx, done := h.repo.NewTransaction()

	for _, data := range input.ReviewList {
		errCode, err := h.updateEventReviewer(tx.DB(), l, data, eventID)
		if err != nil {
			c.JSON(errCode, view.CreateResponse[any](nil, nil, done(err), input, ""))
			return
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, done(nil), "ok"))
}

func (h *handler) updateEventReviewer(db *gorm.DB, l logger.Logger, data PerformanceReviewInput, eventID string) (int, error) {
	// Validate EventID and TopicID
	_, err := h.store.EmployeeEventTopic.One(db, data.TopicID.String(), eventID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(ErrTopicNotFound, "topic not found")
		return http.StatusNotFound, ErrTopicNotFound
	}
	if err != nil {
		l.Error(err, "fail to get employee event topic")
		return http.StatusInternalServerError, err
	}

	// Update Event status to inprogress
	event, err := h.store.FeedbackEvent.One(db, eventID)
	if err != nil {
		l.Error(err, "fail to get feedback event")
		return http.StatusInternalServerError, err
	}

	if event.Status == model.EventStatusDraft {
		event.Status = model.EventStatusInProgress

		event, err = h.store.FeedbackEvent.UpdateSelectedFieldsByID(db, eventID, *event, "status")
		if err != nil {
			l.Error(err, "fail to update status of feedback event")
			return http.StatusInternalServerError, err
		}
	}

	// Update status for employee reviewers
	for _, participant := range data.Participants {
		eventReviewer, err := h.store.EmployeeEventReviewer.One(db, participant.String(), data.TopicID.String())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Errorf(err, "not found employee reviewer with reviewer id = ", participant.String())
			return http.StatusNotFound, err
		}

		if err != nil {
			l.Errorf(err, "fail to get employee reviewer for reviewer ", participant.String())
			return http.StatusInternalServerError, err
		}

		if eventReviewer.ReviewerStatus == model.EventReviewerStatusNone {
			eventReviewer.ReviewerStatus = model.EventReviewerStatusNew
			eventReviewer.AuthorStatus = model.EventAuthorStatusSent

			eventReviewer, err = h.store.EmployeeEventReviewer.UpdateSelectedFieldsByID(db, eventReviewer.ID.String(), *eventReviewer, "reviewer_status", "author_status")
			if err != nil {
				l.Errorf(err, "fail to update employee reviewer for reviewer ", participant.String())
				return http.StatusInternalServerError, err
			}
		}
	}

	return http.StatusOK, nil
}
