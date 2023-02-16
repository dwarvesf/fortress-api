package discord

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
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

func (h *handler) SyncDiscordInfo(c *gin.Context) {
	discordMembers, err := h.service.Discord.GetMembers()
	if err != nil {
		h.logger.Error(err, "failed to get members from discord")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	socialAccounts, err := h.store.SocialAccount.GetByType(h.repo.DB(), model.SocialAccountTypeDiscord.String())
	if err != nil {
		h.logger.Error(err, "failed to get discord accounts")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	discordIDMap := make(map[string]string)
	discordUsernameMap := make(map[string]string)

	for _, member := range discordMembers {
		username := fmt.Sprintf("%s#%s", member.User.Username, member.User.Discriminator)
		discordIDMap[member.User.ID] = username
		discordUsernameMap[username] = member.User.ID
	}

	tx, done := h.repo.NewTransaction()

	for _, sa := range socialAccounts {
		if sa.AccountID == "" && sa.Name == "" {
			continue
		}

		// Update discord_id from username
		if sa.AccountID == "" {
			accountID, ok := discordUsernameMap[sa.Name]
			if !ok {
				h.logger.AddField("username", sa.Name).Info("username does not exist in guild")
				continue
			}

			sa.AccountID = accountID
			_, err := h.store.SocialAccount.UpdateSelectedFieldsByID(tx.DB(), sa.ID.String(), *sa, "account_id")
			if err != nil {
				h.logger.AddField("id", sa.ID).Error(err, "failed to update account_id")
			}

			continue
		}

		// Update username from discord_id
		username, ok := discordIDMap[sa.AccountID]
		if !ok {
			h.logger.Field("account_id", sa.AccountID).Info("discord id does not exist in guild")
			continue
		}

		if sa.Name != username {
			sa.Name = username
			_, err := h.store.SocialAccount.UpdateSelectedFieldsByID(tx.DB(), sa.ID.String(), *sa, "name")
			if err != nil {
				h.logger.AddField("id", sa.ID).Error(err, "failed to update name of social account")
			}
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, done(nil), nil, "ok"))
}
