package bankaccount

import (
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
	repo    store.DBRepo
	config  *config.Config
}

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
// @Summary Get all bank accounts
// @Description Get all bank accounts
// @Tags Bank
// @Accept  json
// @Produce  json
// @Success 200 {object} []view.ListBankAccountResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /bank-accounts [get]
func (h *handler) List(c *gin.Context) {
	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "bank",
		"method":  "List",
	})

	res, err := h.store.BankAccount.All(h.repo.DB())
	if err != nil {
		l.Error(err, "failed to get all bank accounts")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToListBankAccount(res), nil, nil, nil, ""))
}
