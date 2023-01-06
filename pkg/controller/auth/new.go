package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type controller struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
	repo    store.DBRepo
}

func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger) IController {
	return &controller{
		store:   store,
		repo:    repo,
		service: service,
		logger:  logger,
	}
}

type IController interface {
	Auth(c *gin.Context, in AuthenticationInput) (employee *model.Employee, jwt string, err error)
	Me(c *gin.Context, userID string) (employee *model.Employee, perms []*model.Permission, err error)
}
