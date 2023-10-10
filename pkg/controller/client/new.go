package client

import (
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler/client/request"
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
	config  *config.Config
}

func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IController {
	return &controller{
		store:   store,
		repo:    repo,
		service: service,
		logger:  logger,
		config:  cfg,
	}
}

type IController interface {
	Create(c *gin.Context, input request.CreateClientRequest) (client *model.Client, err error)
	List(c *gin.Context) (client []*model.Client, err error)
	Detail(c *gin.Context, clientID string) (client *model.Client, err error)
	Update(c *gin.Context, clientID string, input request.UpdateClientInput) (errCode int, err error)
	Delete(c *gin.Context, clientID string) (errCode int, err error)
	PublicList(c *gin.Context) (client []*model.Client, err error)
}
