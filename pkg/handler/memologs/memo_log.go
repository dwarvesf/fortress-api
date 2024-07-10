package memologs

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/handler/memologs/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/mochiprofile"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/memolog"
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/gin-gonic/gin"
)

type handler struct {
	controller *controller.Controller
	store      *store.Store
	service    *service.Service
	logger     logger.Logger
	repo       store.DBRepo
	config     *config.Config
}

// New returns a handler
func New(controller *controller.Controller, store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		controller: controller,
		store:      store,
		repo:       repo,
		service:    service,
		logger:     logger,
		config:     cfg,
	}
}

func (h *handler) Create(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "memologs",
			"method":  "Create",
		},
	)

	body := request.CreateMemoLogsRequest{}
	if err := c.ShouldBindJSON(&body); err != nil {
		l.Error(err, "[memologs.Create] failed to decode body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	memologs := make([]model.MemoLog, 0)
	for _, b := range body {
		publishedAt, _ := time.Parse(time.RFC3339Nano, b.PublishedAt)

		existedAuthors, err := h.store.DiscordAccount.ListByMemoUsername(h.repo.DB(), b.Authors)
		if err != nil {
			l.Errorf(err, "[memologs.Create] failed to get authors", "authors", b.Authors)
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, b, ""))
			return
		}

		authors := make([]model.DiscordAccount, 0)

		// If author does not exist, we will create new author, use memo username as discord username to query discord member
		mapExistedAuthors := make(map[string]model.DiscordAccount)
		for _, author := range existedAuthors {
			mapExistedAuthors[author.MemoUsername] = author
		}

		for _, authorMemoUsername := range b.Authors {
			if a, ok := mapExistedAuthors[authorMemoUsername]; ok {
				authors = append(authors, a)
				continue
			}

			// Build new author aka new community member
			var newAuthor model.DiscordAccount

			// Search discord user by memo username
			discordMembers, err := h.service.Discord.SearchMember(authorMemoUsername)
			if err != nil {
				l.Errorf(err, "[memologs.Create] failed to get discord user", "discord username", authorMemoUsername)
			}

			var discordMember discordgo.Member
			if len(discordMembers) == 1 && discordMembers[0] != nil {
				discordMember = *discordMembers[0]
			}

			newAuthor.MemoUsername = authorMemoUsername
			newAuthor.DiscordUsername = discordMember.User.Username
			newAuthor.DiscordID = discordMember.User.ID
			newAuthor.PersonalEmail = discordMember.User.Email
			newAuthor.Roles = discordMember.Roles

			// Get profile by discord ID
			profile, err := h.service.MochiProfile.GetProfileByDiscordID(discordMember.User.ID)
			if err != nil {
				l.Errorf(err, "[memologs.Create] failed to get profile", "discord id", discordMember.User.ID)
			}

			if profile != nil {
				for _, assocAccount := range profile.AssociatedAccounts {
					if assocAccount.Platform == mochiprofile.ProfilePlatformGithub {
						githubIDInt, err := strconv.ParseInt(assocAccount.PlatformIdentifier, 10, 64)
						if err != nil {
							l.Errorf(err, "[memologs.Create] failed to parse github ID %d", assocAccount.PlatformIdentifier)
							continue
						}

						newAuthor.GithubUsername, err = h.service.Github.RetrieveUsernameByID(context.Background(), githubIDInt)
						if err != nil {
							l.Errorf(err, "[memologs.Create] failed to get github username with ID %d", githubIDInt)
							continue
						}

						break
					}
				}
			}

			authors = append(authors, newAuthor)
		}

		b := model.MemoLog{
			Title:       b.Title,
			URL:         b.URL,
			Authors:     authors,
			Tags:        b.Tags,
			PublishedAt: &publishedAt,
			Description: b.Description,
			Reward:      b.Reward,
		}
		memologs = append(memologs, b)
	}

	logs, err := h.store.MemoLog.Create(h.repo.DB(), memologs)
	if err != nil {
		l.Errorf(err, "[memologs.Create] failed to create new memo logs", "memologs", memologs)
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, memologs, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToMemoLog(logs), nil, nil, body, ""))
}

func (h *handler) List(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "memologs",
			"method":  "List",
		},
	)

	var fromPtr, toPtr *time.Time

	fromStr := c.Query("from")
	if fromStr != "" {
		from, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			l.Error(err, "failed to parse from time")
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}

		fromPtr = &from
	}

	toStr := c.Query("to")
	if toStr != "" {
		to, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			l.Error(err, "failed to parse to time")
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}

		toPtr = &to
	}

	memoLogs, err := h.store.MemoLog.List(h.repo.DB(), memolog.ListFilter{
		From: fromPtr,
		To:   toPtr,
	})
	if err != nil {
		l.Error(err, "failed to get memologs")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToMemoLog(memoLogs), nil, nil, nil, ""))
}

func (h *handler) Sync(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "memologs",
			"method":  "Sync",
		},
	)

	results, err := h.controller.MemoLog.Sync()
	if err != nil {
		l.Error(err, "failed to sync memologs")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToMemoLog(results), nil, nil, nil, "ok"))
}

func (h *handler) ListOpenPullRequest(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "memologs",
			"method":  "ListOpenPullRequest",
		},
	)

	memoprs, err := h.controller.MemoLog.ListOpenPullRequest()
	if err != nil {
		l.Error(err, "failed to list open pull request")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](memoprs, nil, nil, nil, "ok"))
}
