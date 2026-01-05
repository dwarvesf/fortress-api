package notion

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/gin-gonic/gin"
)

type IHandler interface {
	CreateTechRadar(c *gin.Context)
	GetAvailableProjectsChangelog(c *gin.Context)
	ListEarns(c *gin.Context)
	ListMemos(c *gin.Context)
	ListTechRadars(c *gin.Context)
	ListAudiences(c *gin.Context)
	ListEvents(c *gin.Context)
	ListDigests(c *gin.Context)
	ListUpdates(c *gin.Context)
	ListIssues(c *gin.Context)
	ListStaffingDemands(c *gin.Context)
	ListHiringPositions(c *gin.Context)
	ListProjectMilestones(c *gin.Context)
	SendNewsLetter(c *gin.Context)
	SendProjectChangelog(c *gin.Context)
	SyncTaskOrderLogs(c *gin.Context)
	InitTaskOrderLogs(c *gin.Context)
	CreateContractorFees(c *gin.Context)
	CreateContractorPayouts(c *gin.Context)
	SendTaskOrderConfirmation(c *gin.Context)
}

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
