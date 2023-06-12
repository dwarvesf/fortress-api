package discord

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/employee"
	"github.com/dwarvesf/fortress-api/pkg/store/onleaverequest"
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
		username := member.User.Username
		if member.User.Discriminator != "0" {
			username = fmt.Sprintf("%s#%s", member.User.Username, member.User.Discriminator)
		}

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

// BirthdayDailyMessage check if today is birthday of any employee in the system
// if yes, send birthday message to employee through discord
func (h *handler) BirthdayDailyMessage(c *gin.Context) {
	// check app run mode
	projectID := consts.OperationID
	todoListID := consts.BirthdayToDoListID

	if h.config.Env != "prod" {
		projectID = consts.PlaygroundID
		todoListID = consts.PlaygroundBirthdayTodoID
	}

	//random message pool
	pool := []string{
		`Dear %s, we wish you courage and persistence in reaching all your greatest goals. Have a great birthday!`,
		`Happy Birthday to %s. No one knows your real age, except God, Human Resources and you yourself. Enjoy the blast!`,
		`Happy Birthday, %s! Thank you for being such a great team player and for giving us a perfect excuse to party on a weekday! Let's go grab a drink!`,
		`Just so you know you'd look much younger if not for working in this field :) Happy Birthday, %s`,
		`Congratulation on a great day! Here's to another year closer of retiring! Happy Birthday, %s!`,
		`%s, thank you for being a part of making this company more lively and cheerful. Wish you all the best in this special day.`,
		`Dear %s, we wish you a great birthday and a memorable year. From all the Dwarves brothers.`,
		`I can’t believe you are still single – lol. I hope you have a super day and get everything you want like a companion to share it with. Happy birthday to %s!`,
		`Here’s to another year of version controlling, bug reports, and comments about the documentation looking like code. Happy birthday, mate %s!`,
		`Hope your birthday loops run smoothly and that you don’t break out of the for loop too soon. Cheers, %s!`,
		`Happy birthday, %s. May your code works perfectly the first time you ran it.`,
		`I wish you could have a programming language that does not need compiling, installing, or debugging to run perfectly on the first run. Have a happy birthday, %s`,
	}

	// query active user
	filter := employee.EmployeeFilter{
		WorkingStatuses: []string{"full-time"},
	}

	employees, _, err := h.store.Employee.All(h.repo.DB(), filter, model.Pagination{
		Page: 0,
		Size: 1000,
	})
	if err != nil {
		h.logger.Error(err, "failed to get employees")
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// format message if there is user's birthday
	var names string
	var birthDateNames []string
	todayDate := time.Now().Format("01/02")
	for _, e := range employees {
		now := time.Now()
		if now.Day() == e.DateOfBirth.Day() && now.Month() == e.DateOfBirth.Month() {
			sa := model.SocialAccounts(e.SocialAccounts).GetDiscord()
			if sa != nil && sa.AccountID != "" {
				discordID := sa.AccountID
				names += fmt.Sprintf("<@%s>, ", discordID)
				birthDateNames = append(birthDateNames, e.FullName)
			}
		}
	}

	if len(birthDateNames) == 0 {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, nil, "no birthday today"))
		return
	}

	rand.Seed(time.Now().Unix())
	msg := fmt.Sprintf(pool[rand.Intn(len(pool))], strings.TrimSuffix(names, ", "))

	//send message to Discord channel
	var discordMsg model.DiscordMessage
	discordMsg, err = h.service.Discord.PostBirthdayMsg(msg)
	if err != nil {
		h.logger.Error(err, "failed to post Discord message")
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, discordMsg, ""))
		return
	}

	//Make Basecamp todos
	for _, birthDateName := range birthDateNames {
		birthDayTodo := bcModel.Todo{
			Title:   fmt.Sprintf("Prepare gift for %s, %s", birthDateName, todayDate),
			Content: fmt.Sprintf("Prepare gift for %s, %s", birthDateName, todayDate),
		}
		_, err := h.service.Basecamp.Todo.Create(projectID, todoListID, birthDayTodo)
		if err != nil {
			h.logger.Error(err, "failed create Basecamp todo")
			c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, birthDateName, "k"))
			return
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

// OnLeaveMessage check if today is birthday of any employee in the system
// if yes, send birthday message to employee thru discord
func (h *handler) OnLeaveMessage(c *gin.Context) {
	todayDate := time.Now().Format("2006-01-02")
	onLeaveData, err := h.store.OnLeaveRequest.All(h.repo.DB(), onleaverequest.GetOnLeaveInput{Date: todayDate})
	if err != nil {
		h.logger.Error(err, "failed to get employees")
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	if len(onLeaveData) == 0 {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, nil, "there is no one on leave today"))
		return
	}

	var names string
	for _, e := range onLeaveData {
		discordInfo := model.SocialAccounts(e.Creator.SocialAccounts).GetDiscord()
		names += fmt.Sprintf("<@%s>, ", discordInfo.AccountID)
	}

	msg := fmt.Sprintf("Please be notified that %s will be absent today", strings.TrimSuffix(names, ", "))

	discordMsg, err := h.service.Discord.SendMessage(msg, h.config.Discord.Webhooks.AuditLog)
	if err != nil {
		h.logger.Error(err, "failed to post Discord message")
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, discordMsg, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}
