package vault

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/mochi"
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

	icyTxs := make([]model.IcyTransaction, 0)
	for _, vaultID := range supportedVaults {
		req := &mochi.VaultTransactionRequest{
			VaultID:   vaultID,
			StartTime: startOfTheWeek,
			EndTime:   endOfTheWeek,
		}
		res, err := h.service.Mochi.GetVaultTransaction(req)
		if err != nil {
			l.Error(err, "failed to get GetVaultTransaction")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}

		for _, transaction := range res.Data {
			// skip case transfer through wallet address
			if transaction.Target == "" {
				continue
			}

			txnTime, err := time.Parse("2006-01-02T15:04:05Z", transaction.CreatedAt)
			if err != nil {
				continue
			}

			srcDiscordAccount, err := h.store.DiscordAccount.OneByDiscordID(h.repo.DB(), transaction.Sender)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				l.Error(err, "failed to get src discord account")
				c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
				return
			}

			srcEmployee := make([]*model.Employee, 0)
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				srcEmployee, err = h.store.Employee.GetByDiscordAccountID(h.repo.DB(), srcDiscordAccount.ID.String())
				if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					l.Error(err, "failed to get src employee by discord account ID")
					c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
					return
				}
			}

			destDiscordAccount, err := h.store.DiscordAccount.OneByDiscordID(h.repo.DB(), transaction.Target)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				l.Error(err, " failed to get dest discord account")
				c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
				return
			}

			destEmployee := make([]*model.Employee, 0)
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				destEmployee, err = h.store.Employee.GetByDiscordAccountID(h.repo.DB(), destDiscordAccount.ID.String())
				if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					l.Error(err, "failed to get dest employee by discord account ID")
					c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
					return
				}
			}

			icyTx := model.IcyTransaction{
				Category: strings.ToLower(transaction.VaultName),
				TxnTime:  txnTime,
				Amount:   transaction.Amount,
				Sender:   transaction.Sender,
				Target:   transaction.Target,
			}

			if len(srcEmployee) > 0 {
				icyTx.SrcEmployeeID = srcEmployee[0].ID
			}

			if len(destEmployee) > 0 {
				icyTx.DestEmployeeID = destEmployee[0].ID
			}

			icyTxs = append(icyTxs, icyTx)
		}
	}

	// case no tx from mochi
	if len(icyTxs) == 0 {
		l.Info("There is no transaction in this week")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, nil, nil, "there is no transaction in this week"))
		return
	}

	tx, done := h.repo.NewTransaction()
	if err := h.store.IcyTransaction.Create(tx.DB(), icyTxs); err != nil {
		l.Error(done(err), "failed to Create IcyTransaction")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, done(nil), nil, "ok"))
}
