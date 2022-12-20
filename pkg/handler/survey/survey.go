package survey

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler/survey/errs"
	"github.com/dwarvesf/fortress-api/pkg/handler/survey/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
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

// ListSurvey godoc
// @Summary Get list event
// @Description Get list event
// @Tags Survey
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param subtype query string true "Event Subtype"
// @Param projectIDs query []string false "ProjectIDs"
// @Param page query string false "Page"
// @Param size query string false "Size"
// @Success 200 {object} view.ListSurveyResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /surveys [get]
func (h *handler) ListSurvey(c *gin.Context) {
	input := request.GetListSurveyInput{}
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
		"handler": "survey",
		"method":  "ListSurvey",
		"input":   input,
	})

	events, total, err := h.store.FeedbackEvent.
		GetBySubtypeAndProjectIDs(h.repo.DB(), input.Subtype, input.ProjectIDs, input.Pagination)
	if err != nil {
		l.Error(err, "failed to get feedback events by subtype")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// count likert-scale question by event and projects
	if input.Subtype == model.EventSubtypeWork.String() {
		for i := range events {
			count, err := h.store.EmployeeEventQuestion.
				CountLikertScaleByEventIDAndProjectIDs(h.repo.DB(), events[i].ID.String(), input.ProjectIDs)
			if err != nil {
				l.AddField("eventID", events[i].ID).Error(err, "failed to count likert-scale by eventID and projectIDs")
				c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
				return
			}

			events[i].Count = count
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToListSurvey(events),
		&view.PaginationResponse{Pagination: input.Pagination, Total: total}, nil, nil, ""))
}

// GetSurveyDetail godoc
// @Summary Get survey detail
// @Description Get survey detail
// @Tags Survey
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Feedback Event ID"
// @Param page query string false "Page"
// @Param size query string false "Size"
// @Success 200 {object} view.ListSurveyDetailResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /surveys/{id} [get]
func (h *handler) GetSurveyDetail(c *gin.Context) {
	input := request.GetSurveyDetailInput{
		EventID: c.Param("id"),
	}
	if err := c.ShouldBindQuery(&input.Query); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if err := input.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	input.Query.Standardize()

	l := h.logger.Fields(logger.Fields{
		"handler": "survey",
		"method":  "GetSurveyDetail",
		"input":   input,
	})

	// check feedback event existence
	event, err := h.store.FeedbackEvent.One(h.repo.DB(), input.EventID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(err, "event not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrEventNotFound, input, ""))
		return
	}
	if err != nil {
		l.Error(err, "failed to get feedback event")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	topics, total, err := h.store.EmployeeEventTopic.All(h.repo.DB(),
		employeeeventtopic.GetByEventIDInput{
			EventID: input.EventID,
			Keyword: input.Query.Keyword,
			Preload: true,
			Paging:  true,
		},
		&input.Query.Pagination)
	if err != nil {
		l.Error(err, "failed to get employee event topic by eventID")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	event.Topics = topics

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToSurveyDetail(event),
		&view.PaginationResponse{Pagination: input.Query.Pagination, Total: total}, nil, nil, ""))
}

// CreateSurvey godoc
// @Summary Create new survey
// @Description Create new survey
// @Tags Survey
// @Accept  json
// @Produce  json
// @Param Body body request.CreateSurveyFeedbackInput true "Body"
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

	var req request.CreateSurveyFeedbackInput

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "survey",
		"method":  "CreateSurvey",
		"input":   req,
	})

	if !model.EventSubtype(req.Type).IsValidSurvey() {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidEventSubType, req, ""))
		return
	}

	tx, done := h.repo.NewTransaction()

	switch model.EventSubtype(req.Type) {
	case model.EventSubtypePeerReview:
		status, err := h.createPeerReview(tx.DB(), req, userID)
		if err != nil {
			l.Error(err, "failed to create new survey peer review")
			c.JSON(status, view.CreateResponse[any](nil, nil, done(err), nil, ""))
			return
		}
	case model.EventSubtypeEngagement:
		status, err := h.createEngagement(tx.DB(), req, userID)
		if err != nil {
			l.Error(err, "failed to create new survey engagement")
			c.JSON(status, view.CreateResponse[any](nil, nil, done(err), nil, ""))
			return
		}
	case model.EventSubtypeWork:
		status, err := h.createWorkEvent(tx.DB(), req, userID)
		if err != nil {
			l.Error(err, "failed to create new survey work")
			c.JSON(status, view.CreateResponse[any](nil, nil, done(err), nil, ""))
			return
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, done(nil), nil, "success"))

}

func (h *handler) createPeerReview(db *gorm.DB, req request.CreateSurveyFeedbackInput, userID string) (int, error) {
	//1. convert data
	var startTime, endTime time.Time
	var title string

	if req.Year < time.Now().Year()-1 {
		return http.StatusBadRequest, errs.ErrInvalidYear
	}

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
		return http.StatusBadRequest, errs.ErrInvalidQuarter
	}

	//1.2 check event existed
	_, err := h.store.FeedbackEvent.GetByTypeInTimeRange(db, model.EventTypeSurvey, model.EventSubtypePeerReview, &startTime, &endTime)
	if err == nil {
		return http.StatusBadRequest, errs.ErrEventAlreadyExisted
	} else {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusInternalServerError, err
		}
	}

	//1.3 check employee existed
	createdBy, err := h.store.Employee.One(db, userID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusNotFound, errs.ErrEmployeeNotFound
		}
		return http.StatusInternalServerError, err
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
		return http.StatusInternalServerError, err
	}

	//3. create EmployeeEventTopic
	employees, err := h.store.Employee.GetByWorkingStatus(db, model.WorkingStatusFullTime)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	eets := make([]model.EmployeeEventTopic, 0)
	for _, e := range employees {
		topicTitle := fmt.Sprintf("Peer Performance Review: %s - %s", e.DisplayName, title)
		eets = append(eets, model.EmployeeEventTopic{
			BaseModel: model.BaseModel{
				ID: model.NewUUID(),
			},
			Title:      topicTitle,
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
			return http.StatusInternalServerError, err
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
		return http.StatusInternalServerError, err
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
			return http.StatusInternalServerError, err
		}
		i = to
	}

	//4. create EmployeeEventQuestion
	questions, err := h.store.Question.AllByCategory(db, model.EventTypeSurvey, model.EventSubtypePeerReview)
	if err != nil {
		return http.StatusInternalServerError, err
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
			return http.StatusInternalServerError, err
		}
		i = to
	}

	return http.StatusOK, nil
}

func (h *handler) createEngagement(db *gorm.DB, req request.CreateSurveyFeedbackInput, userID string) (int, error) {
	//1. convert data
	var startTime, endTime time.Time
	var title string

	if req.Year < time.Now().Year()-1 {
		return http.StatusBadRequest, errs.ErrInvalidYear
	}

	switch strings.ToLower(strings.ReplaceAll(req.Quarter, " ", "")) {
	case "q1":
		startTime = time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC)
		endTime = time.Date(req.Year, 3, 31, 23, 59, 59, 59, time.UTC)
		title = fmt.Sprintf("Q1, %d", req.Year)
	case "q2":
		startTime = time.Date(req.Year, 4, 1, 0, 0, 0, 0, time.UTC)
		endTime = time.Date(req.Year, 6, 30, 23, 59, 59, 59, time.UTC)
		title = fmt.Sprintf("Q2, %d", req.Year)
	case "q3":
		startTime = time.Date(req.Year, 7, 1, 0, 0, 0, 0, time.UTC)
		endTime = time.Date(req.Year, 9, 30, 23, 59, 59, 59, time.UTC)
		title = fmt.Sprintf("Q3, %d", req.Year)
	case "q4":
		startTime = time.Date(req.Year, 10, 1, 0, 0, 0, 0, time.UTC)
		endTime = time.Date(req.Year, 12, 31, 23, 59, 59, 59, time.UTC)
		title = fmt.Sprintf("Q4, %d", req.Year)
	default:
		return http.StatusBadRequest, errs.ErrInvalidQuarter
	}

	//1.2 check event existed
	_, err := h.store.FeedbackEvent.GetByTypeInTimeRange(db, model.EventTypeSurvey, model.EventSubtypeEngagement, &startTime, &endTime)
	if err == nil {
		return http.StatusBadRequest, errs.ErrEventAlreadyExisted
	} else {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusInternalServerError, err
		}
	}

	//1.3 check employee existed
	createdBy, err := h.store.Employee.One(db, userID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusNotFound, errs.ErrEmployeeNotFound
		}
		return http.StatusInternalServerError, err
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
		return http.StatusInternalServerError, err
	}

	//3. create EmployeeEventTopic
	employees, err := h.store.Employee.GetByWorkingStatus(db, model.WorkingStatusFullTime)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	eets := make([]model.EmployeeEventTopic, 0)
	for _, e := range employees {
		topicTitle := fmt.Sprintf("Engagement Survey: %s - %s", e.DisplayName, title)
		eets = append(eets, model.EmployeeEventTopic{
			BaseModel: model.BaseModel{
				ID: model.NewUUID(),
			},
			Title:      topicTitle,
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
			return http.StatusInternalServerError, err
		}
		i = to
	}

	//4. create EmployeeEventReviewer
	employeeEventMapper := make(map[model.UUID]model.UUID)
	for _, e := range eets {
		employeeEventMapper[e.EmployeeID] = e.ID
	}

	reviewers := make([]model.EmployeeEventReviewer, 0)

	for _, e := range eets {
		reviewers = append(reviewers, model.EmployeeEventReviewer{
			BaseModel: model.BaseModel{
				ID: model.NewUUID(),
			},
			EventID:              event.ID,
			EmployeeEventTopicID: e.ID,
			ReviewerID:           e.EmployeeID,
			Relationship:         model.RelationshipSelf,
			AuthorStatus:         model.EventAuthorStatusDraft,
			ReviewerStatus:       model.EventReviewerStatusNone,
			IsShared:             false,
			IsRead:               false,
		})
	}

	i = 0
	for i < len(reviewers) {
		to := i + 100
		if to > len(reviewers) {
			to = len(reviewers)
		}
		_, err = h.store.EmployeeEventReviewer.BatchCreate(db, reviewers[i:to])
		if err != nil {
			return http.StatusInternalServerError, err
		}
		i = to
	}

	//5. create EmployeeEventQuestion
	questions, err := h.store.Question.AllByCategory(db, model.EventTypeSurvey, model.EventSubtypeEngagement)
	if err != nil {
		return http.StatusInternalServerError, err
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
			return http.StatusInternalServerError, err
		}
		i = to
	}

	return 200, nil
}

func (h *handler) createWorkEvent(db *gorm.DB, req request.CreateSurveyFeedbackInput, userID string) (int, error) {
	//1. convert data
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return http.StatusBadRequest, errs.ErrInvalidDate
	}
	title := date.Format("Jan 11, 2006")

	//1.2 check event existed
	_, err = h.store.FeedbackEvent.GetByTypeInTimeRange(db, model.EventTypeSurvey, model.EventSubtypeWork, &date, &date)
	if err == nil {
		return http.StatusBadRequest, errs.ErrEventAlreadyExisted
	} else {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusInternalServerError, err
		}
	}

	//1.3 check employee existed
	createdBy, err := h.store.Employee.One(db, userID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusNotFound, errs.ErrEmployeeNotFound
		}
		return http.StatusInternalServerError, err
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
		StartDate: &date,
		EndDate:   &date,
	})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	//3. create EmployeeEventTopic
	employees, err := h.store.ProjectMember.GetByProjectIDs(db, req.ProjectIDs)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	eets := make([]model.EmployeeEventTopic, 0)
	for _, e := range employees {
		eets = append(eets, model.EmployeeEventTopic{
			BaseModel: model.BaseModel{
				ID: model.NewUUID(),
			},
			Title:      fmt.Sprintf("Work and Team Survey: %s - %s", title, e.Employee.DisplayName),
			EventID:    event.ID,
			EmployeeID: e.EmployeeID,
			ProjectID:  e.ProjectID,
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
			return http.StatusInternalServerError, err
		}
		i = to
	}

	//4. create EmployeeEventReviewer
	employeeEventMapper := make(map[model.UUID]model.UUID)
	for _, e := range eets {
		employeeEventMapper[e.EmployeeID] = e.ID
	}

	reviewers := make([]model.EmployeeEventReviewer, 0)

	for _, e := range eets {
		reviewers = append(reviewers, model.EmployeeEventReviewer{
			BaseModel: model.BaseModel{
				ID: model.NewUUID(),
			},
			EventID:              event.ID,
			EmployeeEventTopicID: e.ID,
			ReviewerID:           e.EmployeeID,
			Relationship:         model.RelationshipSelf,
			AuthorStatus:         model.EventAuthorStatusSent,
			ReviewerStatus:       model.EventReviewerStatusNew,
			IsShared:             false,
			IsRead:               false,
		})
	}

	i = 0
	for i < len(reviewers) {
		to := i + 100
		if to > len(reviewers) {
			to = len(reviewers)
		}
		_, err = h.store.EmployeeEventReviewer.BatchCreate(db, reviewers[i:to])
		if err != nil {
			return http.StatusInternalServerError, err
		}
		i = to
	}

	//5. create EmployeeEventQuestion
	questions, err := h.store.Question.AllByCategory(db, model.EventTypeSurvey, model.EventSubtypeWork)
	if err != nil {
		return http.StatusInternalServerError, err
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
			return http.StatusInternalServerError, err
		}
		i = to
	}

	return http.StatusOK, nil
}

// SendSurvey godoc
// @Summary Send the survey
// @Description Send the survey
// @Tags Survey
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Feedback Event ID"
// @Param Body body request.SendSurveyInput true "Body"
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /surveys/{id}/send [post]
func (h *handler) SendSurvey(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" || !model.IsUUIDFromString(eventID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidEventID, nil, ""))
		return
	}

	var input request.SendSurveyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "survey",
		"method":  "SendSurvey",
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

func (h *handler) updateEventReviewer(db *gorm.DB, l logger.Logger, data request.Survey, eventID string) (int, error) {
	// Validate EventID and TopicID
	_, err := h.store.EmployeeEventTopic.One(db, data.TopicID.String(), eventID, false)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(errs.ErrTopicNotFound, "topic not found")
		return http.StatusNotFound, errs.ErrTopicNotFound
	}
	if err != nil {
		l.Error(err, "failed to get employee event topic")
		return http.StatusInternalServerError, err
	}

	// Update Event status to inprogress
	event, err := h.store.FeedbackEvent.One(db, eventID)
	if err != nil {
		l.Error(err, "failed to get feedback event")
		return http.StatusInternalServerError, err
	}

	if event.Status == model.EventStatusDraft {
		event.Status = model.EventStatusInProgress

		_, err = h.store.FeedbackEvent.UpdateSelectedFieldsByID(db, eventID, *event, "status")
		if err != nil {
			l.Error(err, "failed to update status of feedback event")
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
			l.Errorf(err, "failed to get employee reviewer for reviewer ", participant.String())
			return http.StatusInternalServerError, err
		}

		if eventReviewer.ReviewerStatus == model.EventReviewerStatusNone {
			eventReviewer.ReviewerStatus = model.EventReviewerStatusNew
			eventReviewer.AuthorStatus = model.EventAuthorStatusSent

			_, err = h.store.EmployeeEventReviewer.UpdateSelectedFieldsByID(db, eventReviewer.ID.String(), *eventReviewer, "reviewer_status", "author_status")
			if err != nil {
				l.Errorf(err, "failed to update employee reviewer for reviewer ", participant.String())
				return http.StatusInternalServerError, err
			}
		}
	}

	return http.StatusOK, nil
}

// DeleteSurvey godoc
// @Summary Delete survey by id
// @Description Delete survey by id
// @Tags Survey
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
		"handler": "survey",
		"method":  "DeleteSurvey",
		"eventID": eventID,
	})

	tx, done := h.repo.NewTransaction()

	if statusCode, err := h.doSurveyDelete(tx.DB(), eventID); err != nil {
		l.Error(err, "failed to delete survey")
		c.JSON(statusCode, view.CreateResponse[any](nil, nil, done(err), eventID, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, done(nil), nil, "ok"))
}

func (h *handler) doSurveyDelete(db *gorm.DB, eventID string) (int, error) {
	l := h.logger.Fields(logger.Fields{
		"handler": "survey",
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
// @Tags Survey
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.FeedbackReviewDetailResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /surveys/{id}/topics/{topicID}/reviews/{reviewID} [get]
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
		"handler": "survey",
		"method":  "GetSurveyReviewDetail",
	})

	topic, err := h.store.EmployeeEventTopic.One(h.repo.DB(), topicID, eventID, true)
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

	var project *model.Project
	if !topic.ProjectID.IsZero() {
		project, err = h.store.Project.One(h.repo.DB(), topic.ProjectID.String(), false)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(errs.ErrProjectNotFound, "project not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, nil, ""))
			return
		}

		if err != nil {
			l.Error(err, "failed to get project")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToFeedbackReviewDetail(questions, topic, review, project), nil, nil, nil, ""))
}

// DeleteSurveyTopic godoc
// @Summary delete survey topic
// @Description delete survey topic
// @Tags Survey
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
		"handler": "survey",
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

	_, err = h.store.EmployeeEventTopic.One(tx.DB(), topicID, eventID, false)
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

	if err := h.store.EmployeeEventTopic.DeleteByID(tx.DB(), topicID); err != nil {
		l.Error(err, "failed to delete event topics")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, done(nil), nil, "ok"))
}

// GetSurveyTopicDetail godoc
// @Summary Get detail for peer review
// @Description Get detail for peer review
// @Tags Survey
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Feedback Event ID"
// @Param topicID path string true "Employee Event Topic ID"
// @Success 200 {object} view.SurveyTopicDetailResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /surveys/{id}/topics/{topicID} [get]
func (h *handler) GetSurveyTopicDetail(c *gin.Context) {
	input := request.PeerReviewDetailInput{
		EventID: c.Param("id"),
		TopicID: c.Param("topicID"),
	}

	if err := input.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "survey",
		"method":  "GetSurveyTopicDetail",
		"input":   input,
	})

	// Check topic and feedback existence
	topic, err := h.store.EmployeeEventTopic.One(h.repo.DB(), input.TopicID, input.EventID, true)
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
// @Tags Survey
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
	input := request.UpdateTopicReviewersInput{
		EventID: c.Param("id"),
		TopicID: c.Param("topicID"),
	}
	if err := c.ShouldBindJSON(&input.Body); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "survey",
		"method":  "UpdateTopicReviewers",
		"input":   input,
	})

	if err := input.Validate(); err != nil {
		l.Error(err, "validate failed")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	// check feedback event existence
	event, err := h.store.FeedbackEvent.One(h.repo.DB(), input.EventID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(errs.ErrEventNotFound, "feedback event not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrEventNotFound, input, ""))
			return
		}
		l.Error(err, "failed to check feedback event existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if event.Subtype != model.EventSubtypePeerReview {
		l.Error(errs.ErrCanNotUpdateParticipants, "event does not allow updating participants")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrCanNotUpdateParticipants, input, ""))
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

func (h *handler) updateTopicReviewer(db *gorm.DB, eventID string, topicID string, body request.UpdateTopicReviewersBody) (int, error) {
	l := h.logger.Fields(logger.Fields{
		"handler": "survey",
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
		if isExists := mustCreateReviewerIDMap[eventReviewer.ReviewerID]; isExists {
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

	eventTopic, err := h.store.EmployeeEventTopic.One(db, topicID, eventID, false)
	if err != nil {
		l.Error(err, "failed to get event topic")
		return http.StatusInternalServerError, err
	}

	// create event reviewer and event question if reviewerID not exist in database
	newEventReviewers := make([]model.EmployeeEventReviewer, 0)
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
		"handler":      "survey",
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
// @Tags Survey
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
		"handler": "survey",
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
		l.Error(err, "failed to get feedback event")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, eventID, ""))
		return
	}

	if event.Status == model.EventStatusDone {
		l.Error(errs.ErrEventHasBeenDone, "event has been done")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrEventHasBeenDone, eventID, ""))
		return
	}

	event.Status = model.EventStatusDone

	_, err = h.store.FeedbackEvent.UpdateSelectedFieldsByID(tx.DB(), eventID, *event, "status")
	if err != nil {
		l.Error(err, "failed to update feedback event")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, eventID, ""))
		return
	}

	// Get all topics
	topics, _, err := h.store.EmployeeEventTopic.All(tx.DB(), employeeeventtopic.GetByEventIDInput{EventID: eventID}, nil)
	if err != nil {
		l.Error(err, "failed to get all topics")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), eventID, ""))
		return
	}

	// Check done for each topic: all user have to done
	for _, topic := range topics {
		if code, err := h.forceEventReviewersToDone(tx.DB(), l, topic.ID.String()); err != nil {
			l.Errorf(err, "failed to force event reviewers of topic %s to done", topic.ID.String())
			c.JSON(code, view.CreateResponse[any](nil, nil, done(err), eventID, ""))
			return
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, done(nil), nil, "ok"))
}

func (h *handler) forceEventReviewersToDone(db *gorm.DB, l logger.Logger, topicID string) (int, error) {
	// Get all reviewers
	reviewers, err := h.store.EmployeeEventReviewer.GetByTopicID(db, topicID)
	if err != nil {
		l.Errorf(err, "failed to get all reviewers with topic ID %s", topicID)
		return http.StatusInternalServerError, err
	}

	for _, reviewer := range reviewers {
		if reviewer.AuthorStatus != model.EventAuthorStatusDone ||
			reviewer.ReviewerStatus != model.EventReviewerStatusDone {
			reviewer.AuthorStatus = model.EventAuthorStatusDone
			reviewer.ReviewerStatus = model.EventReviewerStatusDone
			reviewer.IsForcedDone = true
		}

		_, err := h.store.EmployeeEventReviewer.UpdateSelectedFieldsByID(db, reviewer.ID.String(), *reviewer,
			"author_status",
			"reviewer_status",
			"is_forced_done")
		if err != nil {
			l.AddField("eventReviewerID", reviewer.ID).Error(err, "failed to update event reviewer status")
			return http.StatusInternalServerError, err
		}
	}

	return http.StatusOK, nil
}

// DeleteTopicReviewers godoc
// @Summary Delete reviewers in a topic
// @Description Delete reviewers in a topic
// @Tags Survey
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
	input := request.DeleteTopicReviewersInput{
		EventID: c.Param("id"),
		TopicID: c.Param("topicID"),
	}
	if err := c.ShouldBindJSON(&input.Body); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "survey",
		"method":  "DeleteTopicReviewers",
		"input":   input,
	})

	if err := input.Validate(); err != nil {
		l.Error(err, "validate failed")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	// Check topic and feedback existence
	_, err := h.store.EmployeeEventTopic.One(h.repo.DB(), input.TopicID, input.EventID, false)
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
		"handler":     "survey",
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
