package auth

import (
	"errors"
	"gorm.io/gorm"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
	repo    store.DBRepo
}

func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger) IHandler {
	return &handler{
		store:   store,
		repo:    repo,
		service: service,
		logger:  logger,
	}
}

// Auth godoc
// @Summary Authorise user when login
// @Description Authorise user when login
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param code body string true "Google login code"
// @Param redirectUrl body string true "Google redirect url"
// @Success 200 {object} view.AuthData
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /auth [post]
func (h *handler) Auth(c *gin.Context) {
	// 1. parse code, redirectUrl from body
	var req struct {
		Code        string `json:"code" binding:"required"`
		RedirectURL string `json:"redirectUrl" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	// 1.1 prepare the logger
	l := h.logger.Fields(logger.Fields{
		"handler": "auth",
		"method":  "Auth",
		"body":    req,
	})

	// 2.1 get access token from req code and redirect url
	accessToken, err := h.service.Google.GetAccessToken(req.Code, req.RedirectURL)
	if err != nil {
		l.Error(err, "error getting access token from google")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	// 2.2 get login user email from access token
	primaryEmail, err := h.service.Google.GetGoogleEmail(accessToken)
	if err != nil {
		l.Error(err, "error getting email from google")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	// 2.3 double check empty primary email
	if primaryEmail == "" {
		l.Error(err, "error nil email from google")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	// 2.4 check user is active
	employee, err := h.store.Employee.OneByEmail(h.repo.DB(), primaryEmail)
	if err != nil {
		l.Error(err, "error query employee from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}
	if employee == nil {
		l.Error(err, "error employee is not activated")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	// 2.5 generate jwt bearer token
	authenticationInfo := model.AuthenticationInfo{
		UserID: employee.ID.String(),
		Avatar: employee.Avatar,
		Email:  primaryEmail,
	}
	jwt, err := utils.GenerateJWTToken(&authenticationInfo, time.Now().Add(24*365*time.Hour).Unix(), "JWTSecretKey")
	if err != nil {
		l.Error(err, "error query employee from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	// 3. return auth data
	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToAuthData(jwt, employee), nil, nil, nil, ""))
}

// Me godoc
// @Summary Get logged-in user data
// @Description Get logged-in user data
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.AuthUserResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /auth/me [get]
func (h *handler) Me(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "auth",
		"method":  "Me",
	})

	rs, err := h.store.Employee.One(h.repo.DB(), userID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("user not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
		l.Error(err, "error query employee from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	perms, err := h.store.Permission.GetByEmployeeID(h.repo.DB(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToAuthorizedUserData(rs, perms), nil, nil, nil, ""))
}
