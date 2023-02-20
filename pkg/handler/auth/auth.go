package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/controller/auth"
	"github.com/dwarvesf/fortress-api/pkg/handler/auth/errs"
	"github.com/dwarvesf/fortress-api/pkg/handler/auth/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	controller *controller.Controller
	logger     logger.Logger
	config     *config.Config
}

func New(controller *controller.Controller, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		controller: controller,
		logger:     logger,
		config:     cfg,
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

	e, jwt, err := h.controller.Auth.Auth(auth.AuthenticationInput{
		Code:        req.Code,
		RedirectURL: req.RedirectURL,
	})
	if err != nil {
		l.Info("failed to called controller")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	// 3. return auth data
	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToAuthData(jwt, e), nil, nil, nil, ""))
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
	userID, err := utils.GetUserIDFromContext(c, h.config)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "auth",
		"method":  "Me",
	})

	rs, perms, err := h.controller.Auth.Me(userID)
	if err != nil {
		l.Error(err, "error query employee from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToAuthorizedUserData(rs, perms), nil, nil, nil, ""))
}

// CreateAPIKey godoc
// @Summary Create API key
// @Description Create API key
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.APIKeyResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /auth/api-key [post]
func (h *handler) CreateAPIKey(c *gin.Context) {
	var body request.CreateAPIKeyInput

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "auth",
		"method":  "CreateAPIKey",
	})

	key, err := h.controller.Auth.CreateAPIKey(body.RoleID)
	if err != nil {
		l.Error(err, "error create api key")
		errs.ConvertControllerErr(c, err)
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](&view.APIKeyData{
		Key: key,
	}, nil, nil, nil, ""))
}
