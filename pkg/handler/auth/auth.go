package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/util"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
}

func New(store *store.Store, service *service.Service, logger logger.Logger) IHandler {
	return &handler{
		store:   store,
		service: service,
		logger:  logger,
	}
}

// One godoc
// @Summary Authorise user when login
// @Description Authorise user when login
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param code body string true "Google login code"
// @Param redirect_url body string true "Google redirect url"
// @Success 200 {object} view.AuthData
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /auth [post]
func (h *handler) Auth(c *gin.Context) {
	// 1. parse code, redirectURL from body
	var req struct {
		Code        string `json:"code" binding:"required"`
		RedirectURL string `json:"redirect_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req))
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
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req))
		return
	}

	// 2.2 get login user email from access token
	primaryEmail, err := h.service.Google.GetGoogleEmail(accessToken)
	if err != nil {
		l.Error(err, "error getting email from google")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req))
		return
	}

	// 2.3 double check empty primary email
	if primaryEmail == "" {
		l.Error(err, "error nil email from google")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req))
		return
	}

	// 2.4 check user is actived
	employee, err := h.store.Employee.OneByTeamEmail(primaryEmail)
	if err != nil {
		l.Error(err, "error query employee from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req))
		return
	}
	if employee == nil {
		l.Error(err, "error employee is not activated")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req))
		return
	}

	// 2.5 generate jwt bearer token
	authenticationInfo := model.AuthenticationInfo{
		UserID: employee.ID.String(),
		Avatar: employee.Avatar,
		Email:  primaryEmail,
	}
	jwt, err := util.GenerateJWTToken(&authenticationInfo, time.Now().Add(24*365*time.Hour).Unix(), "JWTSecretKey")
	if err != nil {
		l.Error(err, "error query employee from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req))
		return
	}

	// 3. return auth data
	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToAuthData(jwt, employee), nil, nil, nil))
}
