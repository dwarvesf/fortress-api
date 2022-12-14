package feedback

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler/feedback/errs"
	"github.com/dwarvesf/fortress-api/pkg/handler/feedback/request"
	surveyRequest "github.com/dwarvesf/fortress-api/pkg/handler/survey/request"
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

	input := request.GetListFeedbackInput{}
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
// @Router /feedbacks/{id}/topics/{topicID} [get]
func (h *handler) Detail(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	input := request.DetailInput{
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
		l.Error(errs.ErrTopicNotFound, "topic not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrTopicNotFound, input, ""))
		return
	}

	if err != nil {
		l.Error(err, "failed when getting topic")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	eventReviewer, err := h.store.EmployeeEventReviewer.GetByReviewerID(h.repo.DB(), userID, input.TopicID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(errs.ErrEmployeeEventReviewerNotFound, "employee event reviewer not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrEmployeeEventReviewerNotFound, nil, ""))
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
		l.Error(errs.ErrReviewerNotFound, "reviewer not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrReviewerNotFound, nil, ""))
		return
	}

	if err != nil {
		l.Error(err, "failed to get reviewer")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	detailInfo := view.FeedbackDetailInfo{
		Status:       eventReviewer.ReviewerStatus,
		EmployeeID:   topic.EmployeeID.String(),
		Reviewer:     reviewer,
		TopicID:      input.TopicID,
		EventID:      input.EventID,
		Title:        topic.Title,
		Relationship: eventReviewer.Relationship,
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToListFeedbackDetails(questions, detailInfo), nil, nil, nil, ""))
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
// @Param Body body request.SubmitBody true "Body"
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

	input := request.SubmitInput{
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
		l.Error(errs.ErrTopicNotFound, "topic not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, done(errs.ErrTopicNotFound), input, ""))
		return
	}

	if err != nil {
		l.Error(err, "failed when getting topic")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	eventReviewer, err := h.store.EmployeeEventReviewer.GetByReviewerID(tx.DB(), userID, input.TopicID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(errs.ErrEventReviewerNotFound, "employee event reviewer not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, done(errs.ErrEventReviewerNotFound), input, ""))
		return
	}

	if err != nil {
		l.Error(err, "failed to get employee event reviewer record")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	if eventReviewer.ReviewerStatus == model.EventReviewerStatusDone {
		l.Error(errs.ErrCouldNotEditDoneFeedback, "could not edit the feedback marked as done")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(errs.ErrCouldNotEditDoneFeedback), nil, ""))
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
			l.Error(errs.ErrEventQuestionNotFound(e.EventQuestionID.String()), "employee event question not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, done(errs.ErrEventQuestionNotFound(e.EventQuestionID.String())), input, ""))
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
				l.Error(errs.ErrUnansweredquestions, "there are some unanswered questions")
				c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, done(errs.ErrUnansweredquestions), input, ""))
				return
			}
		}
	}

	reviewer, err := h.store.Employee.One(h.repo.DB(), userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(errs.ErrReviewerNotFound, "reviewer not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrReviewerNotFound, nil, ""))
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

// SendPerformanceReview godoc
// @Summary Send the performance review
// @Description Send the performance review
// @Tags Feedback
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Feedback Event ID"
// @Param Body body request.SendPerformanceReviewInput true "Body"
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /surveys/{id}/send [post]
func (h *handler) SendPerformanceReview(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" || !model.IsUUIDFromString(eventID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidEventID, nil, ""))
		return
	}

	var input surveyRequest.SendPerformanceReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "feedback",
		"method":  "SendPerformanceReview",
		"eventID": eventID,
		"input":   input,
	})

	// Begin transaction
	tx, done := h.repo.NewTransaction()

	for _, data := range input.Topics {
		errCode, err := h.updateEventReviewer(tx.DB(), l, data, eventID)
		if err != nil {
			l.Error(err, "error when running function updateEventReviewer")
			c.JSON(errCode, view.CreateResponse[any](nil, nil, done(err), input, ""))
			return
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, done(nil), "ok"))
}

func (h *handler) updateEventReviewer(db *gorm.DB, l logger.Logger, data surveyRequest.PerformanceReviewTopic, eventID string) (int, error) {
	// Validate EventID and TopicID
	_, err := h.store.EmployeeEventTopic.One(db, data.TopicID.String(), eventID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(errs.ErrTopicNotFound, "topic not found")
		return http.StatusNotFound, errs.ErrTopicNotFound
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
		eventReviewer, err := h.store.EmployeeEventReviewer.GetByReviewerID(db, participant.String(), data.TopicID.String())
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

// DeleteSurvey godoc
// @Summary Delete survey by id
// @Description Delete survey by id
// @Tags Feedback
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Feedback Event ID"
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /surveys/{id} [delete]
func (h *handler) DeleteSurvey(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" || !model.IsUUIDFromString(eventID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidEventID, eventID, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "feedback",
		"method":  "DeleteSurvey",
		"eventID": eventID,
	})

	tx, done := h.repo.NewTransaction()

	if code, err := h.deleteSurvey(tx.DB(), eventID); err != nil {
		l.Error(err, "failed to delete survey")
		c.JSON(code, view.CreateResponse[any](nil, nil, done(err), eventID, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, done(nil), nil, "ok"))
}

func (h *handler) deleteSurvey(db *gorm.DB, eventID string) (int, error) {
	l := h.logger.Fields(logger.Fields{
		"handler": "feedback",
		"method":  "deleteSurvey",
		"eventID": eventID,
	})

	// check feedback event existence
	exists, err := h.store.FeedbackEvent.IsExist(db, eventID)
	if err != nil {
		l.Error(err, "failed to check feedback event existence")
		return http.StatusInternalServerError, err
	}
	if !exists {
		l.Error(err, "feedback event not found")
		return http.StatusNotFound, errs.ErrEventNotFound
	}

	if err := h.store.EmployeeEventQuestion.DeleteByEventID(db, eventID); err != nil {
		l.Error(err, "failed to delete feedback events")
		return http.StatusInternalServerError, err
	}

	if err := h.store.EmployeeEventReviewer.DeleteByEventID(db, eventID); err != nil {
		l.Error(err, "failed to delete event reviewers")
		return http.StatusInternalServerError, err
	}

	if err := h.store.EmployeeEventTopic.DeleteByEventID(db, eventID); err != nil {
		l.Error(err, "failed to delete event topics")
		return http.StatusInternalServerError, err
	}

	if err := h.store.FeedbackEvent.DeleteByID(db, eventID); err != nil {
		l.Error(err, "failed to delete event")
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

// GetSurveyReviewDetail godoc
// @Summary Get survey review detail
// @Description Get survey review detail
// @Tags Feedback
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.FeedbackReviewDetailResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /surveys/{id}/topics/{topicID}/reviews/{reviewID} [post]
func (h *handler) GetSurveyReviewDetail(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" || !model.IsUUIDFromString(eventID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidEventID, nil, ""))
		return
	}

	topicID := c.Param("topicID")
	if topicID == "" || !model.IsUUIDFromString(eventID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidTopicID, nil, ""))
		return
	}

	reviewID := c.Param("reviewID")
	if reviewID == "" || !model.IsUUIDFromString(eventID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidReviewerID, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "feedback",
		"method":  "GetSurveyReviewDetail",
	})

	topic, err := h.store.EmployeeEventTopic.One(h.repo.DB(), topicID, eventID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("topic not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrTopicNotFound, nil, ""))
			return
		}
		l.Error(err, "failed when getting topic")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	review, err := h.store.EmployeeEventReviewer.One(h.repo.DB(), reviewID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("review not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrEventReviewerNotFound, nil, ""))
			return
		}
		l.Error(err, "failed when getting review")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	if review.EmployeeEventTopicID != topic.ID {
		l.Info("review not belong topic")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrEventReviewerNotFound, nil, ""))
		return
	}

	questions, err := h.store.EmployeeEventQuestion.GetByEventReviewerID(h.repo.DB(), review.ID.String())
	if err != nil {
		l.Error(err, "failed when getting questions")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToFeedbackReviewDetail(questions, topic, review), nil, nil, nil, ""))
}

// DeleteSurveyTopic godoc
// @Summary delete survey topic
// @Description delete survey topic
// @Tags Feedback
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /surveys/{id}/topics/{topicID} [delete]
func (h *handler) DeleteSurveyTopic(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" || !model.IsUUIDFromString(eventID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidEventID, nil, ""))
		return
	}

	topicID := c.Param("topicID")
	if topicID == "" || !model.IsUUIDFromString(eventID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidTopicID, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "feedback",
		"method":  "DeleteSurveyTopic",
	})

	tx, done := h.repo.NewTransaction()

	// check feedback event existence
	exists, err := h.store.FeedbackEvent.IsExist(tx.DB(), eventID)
	if err != nil {
		l.Error(err, "failed to check feedback event existence")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}
	if !exists {
		l.Error(err, "feedback event not found")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, done(errs.ErrEventNotFound), nil, ""))
		return
	}

	_, err = h.store.EmployeeEventTopic.One(tx.DB(), topicID, eventID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("topic not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, done(errs.ErrTopicNotFound), nil, ""))
			return
		}
		l.Error(err, "failed when getting topic")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	reviews, err := h.store.EmployeeEventReviewer.GetByTopicID(tx.DB(), topicID)
	if err != nil {
		l.Error(err, "failed when getting reviews")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	reviewIDList := make([]string, 0)
	for _, r := range reviews {
		if r.AuthorStatus != model.EventAuthorStatusDraft {
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(errs.ErrReviewAlreadySent), nil, ""))
			return
		}
		reviewIDList = append(reviewIDList, r.ID.String())
	}

	if err := h.store.EmployeeEventQuestion.DeleteByEventReviewerIDList(tx.DB(), reviewIDList); err != nil {
		l.Error(err, "failed to delete feedback events")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	if err := h.store.EmployeeEventReviewer.DeleteByTopicID(tx.DB(), topicID); err != nil {
		l.Error(err, "failed to delete event reviewers")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	if err := h.store.EmployeeEventTopic.DeleteByID(tx.DB(), eventID); err != nil {
		l.Error(err, "failed to delete event topics")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, done(nil), nil, "ok"))
}

// GetPeerReviewDetail godoc
// @Summary Get detail for peer review
// @Description Get detail for peer review
// @Tags Feedback
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Feedback Event ID"
// @Param topicID path string true "Employee Event Topic ID"
// @Success 200 {object} view.PeerReviewDetailResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /surveys/{id}/topics/{topicID} [get]
func (h *handler) GetPeerReviewDetail(c *gin.Context) {
	input := surveyRequest.PeerReviewDetailInput{
		EventID: c.Param("id"),
		TopicID: c.Param("topicID"),
	}

	if err := input.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "feedback",
		"method":  "GetPeerReviewDetail",
		"input":   input,
	})

	// Check topic and feedback existence
	topic, err := h.store.EmployeeEventTopic.One(h.repo.DB(), input.TopicID, input.EventID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(errs.ErrTopicNotFound, "topic not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrTopicNotFound, input, ""))
		return
	}

	if err != nil {
		l.Error(err, "failed when getting topic")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToPeerReviewDetail(topic), nil, nil, nil, ""))
}

// UpdateTopicReviewers godoc
// @Summary Update reviewers in a topic
// @Description Update reviewers in a topic
// @Tags Feedback
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param Body body request.UpdateTopicReviewersBody true "Body"
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /surveys/{id}/topics/{topicID}/employees [put]
func (h *handler) UpdateTopicReviewers(c *gin.Context) {
	input := surveyRequest.UpdateTopicReviewersInput{
		EventID: c.Param("id"),
		TopicID: c.Param("topicID"),
	}
	if err := c.ShouldBindJSON(&input.Body); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "feedback",
		"method":  "UpdateTopicReviewers",
		"input":   input,
	})

	if err := input.Validate(); err != nil {
		l.Error(err, "validate failed")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	// check feedback event existence
	exists, err := h.store.FeedbackEvent.IsExist(h.repo.DB(), input.EventID)
	if err != nil {
		l.Error(err, "failed to check feedback event existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}
	if !exists {
		l.Error(errs.ErrEventNotFound, "feedback event not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrEventNotFound, input, ""))
		return
	}

	tx, done := h.repo.NewTransaction()

	if code, err := h.updateTopicReviewer(tx.DB(), input.EventID, input.TopicID, input.Body); err != nil {
		l.Error(err, "failed to update topic reviewers")
		c.JSON(code, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, done(nil), nil, "ok"))
}

func (h *handler) updateTopicReviewer(db *gorm.DB, eventID string, topicID string, body surveyRequest.UpdateTopicReviewersBody) (int, error) {
	l := h.logger.Fields(logger.Fields{
		"handler": "feedback",
		"method":  "updateTopicReviewer",
		"eventID": eventID,
		"topicID": topicID,
		"body":    body,
	})

	employees, err := h.store.Employee.GetByWorkingStatus(db, model.WorkingStatusFullTime)
	if err != nil {
		l.Error(err, "failed to get employees by working status")
		return http.StatusInternalServerError, err
	}

	eventReviewers, err := h.store.EmployeeEventReviewer.GetByTopicID(db, topicID)
	if err != nil {
		l.Error(err, "failed to get event reviewers")
		return http.StatusInternalServerError, err
	}

	// check reviewer existence
	employeeMap := model.ToEmployeeMap(employees)
	mustCreateReviewerIDMap := map[model.UUID]bool{}
	for _, reviewerID := range body.ReviewerIDs {
		if _, ok := employeeMap[reviewerID]; !ok {
			l.Errorf(errs.ErrEmployeeNotReady, "employee %v not ready", reviewerID)
			return http.StatusBadRequest, errs.ErrEmployeeNotReady
		}

		mustCreateReviewerIDMap[reviewerID] = true
	}

	// delete event question and event topic if reviewerID is not exists in request
	for _, eventReviewer := range eventReviewers {
		if isExists, _ := mustCreateReviewerIDMap[eventReviewer.ReviewerID]; isExists {
			mustCreateReviewerIDMap[eventReviewer.ReviewerID] = false
			continue
		}

		if err := h.store.EmployeeEventQuestion.DeleteByEventReviewerID(db, eventReviewer.ID.String()); err != nil {
			l.Error(err, "failed to delete event questions")
			return http.StatusInternalServerError, err
		}

		if err := h.store.EmployeeEventReviewer.DeleteByID(db, eventReviewer.ID.String()); err != nil {
			l.Error(err, "failed to delete event reviewer")
			return http.StatusInternalServerError, err
		}
	}

	eventTopic, err := h.store.EmployeeEventTopic.One(db, topicID, eventID)
	if err != nil {
		l.Error(err, "failed to get event topic")
		return http.StatusInternalServerError, err
	}

	// create event reviewer and event question if reviewerID not exist in database
	newEventReviewers := []model.EmployeeEventReviewer{}
	for reviewerID, mustCreate := range mustCreateReviewerIDMap {
		if mustCreate {
			relationship := model.RelationshipPeer
			if employeeMap[eventTopic.EmployeeID].LineManagerID == reviewerID ||
				employeeMap[reviewerID].LineManagerID == eventTopic.EmployeeID {
				relationship = model.RelationshipLineManager
			}

			newEventReviewers = append(newEventReviewers, model.EmployeeEventReviewer{
				EmployeeEventTopicID: model.MustGetUUIDFromString(topicID),
				ReviewerID:           reviewerID,
				AuthorStatus:         model.EventAuthorStatusDraft,
				ReviewerStatus:       model.EventReviewerStatusNone,
				Relationship:         relationship,
				EventID:              model.MustGetUUIDFromString(eventID),
			})
		}
	}

	if len(newEventReviewers) > 0 {
		newEventReviewers, err = h.store.EmployeeEventReviewer.BatchCreate(db, newEventReviewers)
		if err != nil {
			l.Error(err, "failed to batch create event reviews")
			return http.StatusInternalServerError, err
		}
	}

	if err := h.createEventQuestions(db, model.EventTypeSurvey, model.EventSubtypePeerReview, newEventReviewers); err != nil {
		l.Error(err, "failed to create event questions")
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func (h *handler) createEventQuestions(db *gorm.DB, eventType model.EventType, eventSubtype model.EventSubtype, reviewers []model.EmployeeEventReviewer) error {
	l := h.logger.Fields(logger.Fields{
		"handler":      "feedback",
		"method":       "createEventQuestions",
		"eventType":    eventType,
		"eventSubtype": eventSubtype,
	})

	questions, err := h.store.Question.AllByCategory(db, eventType, eventSubtype)
	if err != nil {
		l.Error(err, "failed to get all questions by category")
		return err
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

	i := 0
	for i < len(eventQuestions) {
		to := i + 100
		if to > len(eventQuestions) {
			to = len(eventQuestions)
		}
		_, err = h.store.EmployeeEventQuestion.BatchCreate(db, eventQuestions[i:to])
		if err != nil {
			l.Error(err, "failed to batch create event questions")
			return err
		}
		i = to
	}

	return nil
}

// MarkDone godoc
// @Summary Mark done feedback event
// @Description Mark done feedback event
// @Tags Feedback
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Feedback Event ID"
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /surveys/{id}/done [put]
func (h *handler) MarkDone(c *gin.Context) {
	eventID := c.Param("id")

	if eventID == "" || !model.IsUUIDFromString(eventID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidEventID, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "feedback",
		"method":  "MarkDone",
		"eventID": eventID,
	})

	// check feedback event existence
	exists, err := h.store.FeedbackEvent.IsExist(h.repo.DB(), eventID)
	if err != nil {
		l.Error(err, "failed to check feedback event existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, eventID, ""))
		return
	}
	if !exists {
		l.Error(errs.ErrEventNotFound, "feedback event not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrEventNotFound, eventID, ""))
		return
	}

	tx, done := h.repo.NewTransaction()

	// Update event status
	event, err := h.store.FeedbackEvent.One(tx.DB(), eventID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(errs.ErrEventNotFound, "feedback event not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrEventNotFound, eventID, ""))
		return
	}

	if err != nil {
		l.Error(err, "fail to get feedback event")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, eventID, ""))
	}

	event.Status = model.EventStatusDone

	event, err = h.store.FeedbackEvent.UpdateSelectedFieldsByID(tx.DB(), eventID, *event, "status")
	if err != nil {
		l.Error(err, "fail to update status of feedback event")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, eventID, ""))
	}

	// Get all topics
	topics, err := h.store.EmployeeEventTopic.All(tx.DB(), eventID)
	if err != nil {
		l.Error(err, "failed to get all topics")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), eventID, ""))
		return
	}

	// Check done for each topic: all user have to done
	for _, topic := range topics {
		if code, err := h.checkDoneTopic(tx.DB(), l, eventID, topic.ID.String()); err != nil {
			l.Errorf(err, "failed to check done topic with ID %s", topic.ID.String())
			c.JSON(code, view.CreateResponse[any](nil, nil, done(err), eventID, ""))
			return
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, done(nil), nil, "ok"))
}

func (h *handler) checkDoneTopic(db *gorm.DB, l logger.Logger, eventID string, topicID string) (int, error) {
	// Get all reviewers
	reviewers, err := h.store.EmployeeEventReviewer.GetByTopicID(db, topicID)
	if err != nil {
		l.Errorf(err, "failed to get all reviewers with topic ID %s", topicID)
		return http.StatusInternalServerError, err
	}

	// Check done of employee event reviewer
	for _, reviewer := range reviewers {
		if reviewer.ReviewerStatus != model.EventReviewerStatusDone {
			l.Errorf(errs.ErrUnfinishedReviewer, "the evaluation performed by the reviewer with ID %s has not been finished", topicID, reviewer.ID.String())
			return http.StatusBadRequest, errs.ErrUnfinishedReviewer
		}
	}

	return http.StatusOK, nil
}

// DeleteTopicReviewers godoc
// @Summary Delete reviewers in a topic
// @Description Delete reviewers in a topic
// @Tags Feedback
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Feedback Event ID"
// @Param topicID path string true "Employee Event Topic ID"
// @Param Body body request.DeleteTopicReviewersBody true "Body"
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /surveys/{id}/topics/{topicID}/employees [delete]
func (h *handler) DeleteTopicReviewers(c *gin.Context) {
	input := surveyRequest.DeleteTopicReviewersInput{
		EventID: c.Param("id"),
		TopicID: c.Param("topicID"),
	}
	if err := c.ShouldBindJSON(&input.Body); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "feedback",
		"method":  "DeleteTopicReviewers",
		"input":   input,
	})

	if err := input.Validate(); err != nil {
		l.Error(err, "validate failed")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	// Check topic and feedback existence
	_, err := h.store.EmployeeEventTopic.One(h.repo.DB(), input.TopicID, input.EventID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(errs.ErrTopicNotFound, "topic not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrTopicNotFound, input, ""))
		return
	}

	if err != nil {
		l.Error(err, "failed when getting topic")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	tx, done := h.repo.NewTransaction()

	if code, err := h.deleteTopicReviewer(tx.DB(), input.EventID, input.TopicID, input.Body.ReviewerIDs); err != nil {
		l.Error(err, "failed to delete topic reviewers")
		c.JSON(code, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, done(nil), nil, "ok"))
}

func (h *handler) deleteTopicReviewer(db *gorm.DB, eventID string, topicID string, reviewerIDs []model.UUID) (int, error) {
	l := h.logger.Fields(logger.Fields{
		"handler":     "feedback",
		"method":      "deleteTopicReviewer",
		"eventID":     eventID,
		"topicID":     topicID,
		"reviewerIDs": reviewerIDs,
	})

	for _, reviewer := range reviewerIDs {
		eventReviewer, err := h.store.EmployeeEventReviewer.GetByReviewerID(db, reviewer.String(), topicID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Errorf(errs.ErrEventReviewerNotFound, "reviewer not found with reviewerID ", reviewer)
			return http.StatusNotFound, errs.ErrEventReviewerNotFound
		}
		if err != nil {
			l.Errorf(err, "failed when get employee event reviewer with reviewerID ", reviewer)
			return http.StatusNotFound, errs.ErrEventReviewerNotFound
		}

		// Delete employee question
		if err := h.store.EmployeeEventQuestion.DeleteByEventReviewerID(db, eventReviewer.ID.String()); err != nil {
			l.Error(err, "failed to delete employee event question")
			return http.StatusInternalServerError, err
		}

		// Delete employee reviewer
		if err := h.store.EmployeeEventReviewer.DeleteByID(db, eventReviewer.ID.String()); err != nil {
			l.Error(err, "failed to delete employee event reviewer")
			return http.StatusInternalServerError, err
		}
	}

	return http.StatusOK, nil
}
