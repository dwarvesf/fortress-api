package event

import (
	"context"
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/mochipay"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/discordevent"
	"github.com/ethereum/go-ethereum/log"
)

type IController interface {
	SweepOgifEvent(c context.Context) error
}

type controller struct {
	service *service.Service
	logger  logger.Logger
	config  *config.Config
	store   *store.Store
	repo    store.DBRepo
}

func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IController {
	return &controller{
		store:   store,
		service: service,
		logger:  logger,
		config:  cfg,
		repo:    repo,
	}
}
func (c *controller) SweepOgifEvent(ctx context.Context) error {
	isSender := true
	// Fetch latest 50 transactions
	txReq := mochipay.ListTransactionsRequest{
		ActionList:   []mochipay.TransactionAction{mochipay.TransactionActionVaultTransfer},
		IsSender:     &isSender,
		ProfileID:    "1707664412564787200",
		Size:         100,
		SortBy:       "created_at-",
		Status:       mochipay.TransactionStatusSuccess,
		TokenAddress: mochipay.ICYAddress,
	}
	txResp, err := c.service.MochiPay.GetListTransactions(txReq)
	if err != nil {
		return err
	}

	if len(txResp.Data) == 0 {
		return nil // No transactions to process
	}

	// Get the time of the last transaction
	lastTxTime := txResp.Data[len(txResp.Data)-1].CreatedAt

	// Fetch all events after the last transaction time
	events, err := c.store.DiscordEvent.All(c.repo.DB(), &discordevent.Query{
		After: &lastTxTime,
		Limit: 100,
	}, true)
	if err != nil {
		return err
	}

	txMap := make(map[string]bool)
	for _, tx := range txResp.Data {
		txMap[tx.Id] = false
	}

	for _, event := range events {
		if len(event.EventSpeakers) != 0 {
			continue
		}
		// Find transactions for this event
		for _, tx := range txResp.Data {
			if !strings.HasPrefix(strings.ToLower(tx.Metadata["message"].(string)), "ogif") {
				log.Debug("tx is not ogif")
				continue
			}

			txCreatedDate := time.Date(tx.CreatedAt.Year(), tx.CreatedAt.Month(), tx.CreatedAt.Day(), 0, 0, 0, 0, tx.CreatedAt.Location())
			eventDate := time.Date(event.Date.Year(), event.Date.Month(), event.Date.Day(), 0, 0, 0, 0, event.Date.Location())
			if txCreatedDate.Before(eventDate) {
				log.Debug("tx created before event date")
				continue
			}

			if txMap[tx.Id] {
				log.Debug("tx already processed")
				continue
			}

			profile, err := c.service.MochiProfile.GetProfile(tx.FromProfileId)
			if err != nil {
				txMap[tx.Id] = true
				c.logger.Error(err, "failed to get MochiProfile")
				continue
			}

			var discordID string
			for _, account := range profile.AssociatedAccounts {
				if account.Platform == "discord" {
					discordID = account.PlatformIdentifier
					break
				}
			}

			if discordID == "" {
				log.Debug("tx is not from discord")
				continue
			}

			discordAccount, err := c.store.DiscordAccount.OneByDiscordID(c.repo.DB(), discordID)
			if err != nil {
				c.logger.Error(err, "failed to get discordAccount")
				continue
			}

			speaker := &model.EventSpeaker{
				EventID:          event.ID,
				DiscordAccountID: discordAccount.ID,
				Topic:            tx.Metadata["message"].(string),
			}

			_, err = c.store.EventSpeaker.Create(c.repo.DB(), speaker)
			if err != nil {
				c.logger.Error(err, "failed to create event speaker")
				continue
			}

			txMap[tx.Id] = true
		}
	}

	return nil
}
