package vault

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
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

// StoreVaultTransaction godoc
// @Summary Store vault tx as icy tx from Mochi service
// @Description Store vault tx as icy tx from Mochi service
// @Tags Vault
// @Accept  json
// @Produce  json
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /cron-jobs/store-vault-transaction [post]
func (h *handler) StoreVaultTransaction(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "vault",
		"method":  "StoreVaultTransaction",
	})

	// currently support
	supportedVaults := []string{"18", "19", "20"}

	startOfTheWeek := timeutil.FormatDateForCurl(timeutil.GetStartDayOfWeek(time.Now().Local()).Format(time.RFC3339))
	endOfTheWeek := timeutil.FormatDateForCurl(timeutil.GetEndDayOfWeek(time.Now().Local()).Format(time.RFC3339))

	for _, vaultId := range supportedVaults {
		req := &model.VaultTransactionRequest{
			VaultId:   vaultId,
			StartTime: startOfTheWeek,
			EndTime:   endOfTheWeek,
		}
		res, err := h.service.Mochi.GetVaultTransaction(req)
		if err != nil {
			l.Error(err, "GetVaultTransaction failed")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}

		for _, transaction := range res.Data {
			err := h.store.IcyTransaction.Create(h.repo.DB(), &model.IcyTransaction{
				Vault:              transaction.VaultName,
				Amount:             transaction.Amount,
				Token:              transaction.Token,
				SenderDiscordId:    transaction.Sender,
				RecipientAddress:   transaction.ToAddress,
				RecipientDiscordId: transaction.Target,
			})
			if err != nil {
				l.Error(err, "Create IcyTransaction failed")
				c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
				return
			}
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}
