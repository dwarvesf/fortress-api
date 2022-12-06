package feedback

import (
	"errors"
	"net/http"

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

	detailInfo := view.FeedbackDetailInfo{
		Status:     eventReviewer.Status,
		EmployeeID: topic.EmployeeID.String(),
		ReviewerID: userID,
		TopicID:    input.TopicID,
		EventID:    input.EventID,
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
	exists, err := h.store.FeedbackEvent.IsExist(h.repo.DB(), input.EventID)
	if err != nil {
		l.Error(err, "failed to check feedback event existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if !exists {
		l.Error(ErrEventNotFound, "event not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrEventNotFound, input, ""))
		return
	}

	topics, total, err := h.store.EmployeeEventTopic.GetByEventIDWithPagination(h.repo.DB(), input.EventID, input.Pagination)
	if err != nil {
		l.Error(err, "failed to get employee event topic by eventID")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToListSurveyDetail(topics),
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
// @Router /feedbacks/{id}/topics/{topicID}/answers [put]
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

	if eventReviewer.Status == model.EventReviewerStatusDone {
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
	eventReviewer.Status = input.Body.Status
	_, err = h.store.EmployeeEventReviewer.UpdateSelectedFieldsByID(tx.DB(), eventReviewer.ID.String(), *eventReviewer, "status")
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

	detailInfo := view.FeedbackDetailInfo{
		Status:     eventReviewer.Status,
		EmployeeID: topic.EmployeeID.String(),
		ReviewerID: userID,
		TopicID:    input.TopicID,
		EventID:    input.EventID,
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToListSubmitFeedback(eventQuestions, detailInfo), nil, nil, done(nil), ""))
}
