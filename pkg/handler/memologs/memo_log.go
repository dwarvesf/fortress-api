package memologs

import (
	"net/http"
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

		existedAuthors, err := h.store.CommunityMember.ListByUsernames(h.repo.DB(), b.Authors)
		if err != nil {
			l.Errorf(err, "[memologs.Create] failed to get authors", "authors", b.Authors)
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, b, ""))
			return
		}

		authors := make([]model.CommunityMember, 0)

		mapExistedAuthors := make(map[string]model.CommunityMember)
		for _, author := range existedAuthors {
			mapExistedAuthors[author.DiscordUsername] = author
		}

		for _, authorDiscordUsername := range b.Authors {
			if a, ok := mapExistedAuthors[authorDiscordUsername]; ok {
				authors = append(authors, a)
				continue
			}

			// Build new author aka new community member
			var newAuthor model.CommunityMember

			// Get discord user by discord username
			discordMembers, err := h.service.Discord.SearchMember(authorDiscordUsername)
			if err != nil {
				l.Errorf(err, "[memologs.Create] failed to get discord user", "discord username", authorDiscordUsername)
			}

			var discordMember discordgo.Member
			if len(discordMembers) == 1 && discordMembers[0] != nil {
				discordMember = *discordMembers[0]
			}

			newAuthor.DiscordUsername = authorDiscordUsername
			newAuthor.DiscordID = discordMember.User.ID

			// Get employee by discord ID
			employee, err := h.store.Employee.GetByDiscordUsername(h.repo.DB(), authorDiscordUsername)
			if err != nil {
				l.Errorf(err, "[memologs.Create] failed to get employee", "discord username", authorDiscordUsername)
			}

			if employee != nil {
				newAuthor.EmployeeID = &employee.ID
				newAuthor.PersonalEmail = employee.PersonalEmail
			}

			// Get profile by discord ID
			profile, err := h.service.MochiProfile.GetProfileByDiscordID(discordMember.User.ID)
			if err != nil {
				l.Errorf(err, "[memologs.Create] failed to get profile", "discord username", authorDiscordUsername)
			}

			if profile != nil {
				for _, assocAccount := range profile.AssociatedAccounts {
					if assocAccount.Platform == mochiprofile.ProfilePlatformGithub {
						newAuthor.GithubUsername = assocAccount.PlatformIdentifier
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

	memoLogs, err := h.store.MemoLog.List(h.repo.DB())
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
