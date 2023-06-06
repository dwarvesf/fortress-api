package birthday

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	models "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/employee"
	"github.com/gin-gonic/gin"
)

type birthday struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
	repo    store.DBRepo
	config  *config.Config
}

func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IHandler {
	return &birthday{store: store, repo: repo, service: service, logger: logger, config: cfg}
}

// BirthdayDailyMessage check if today is birthday of any employee in the system
// if yes, send birthday message to employee thru discord
func (b *birthday) BirthdayDailyMessage(c *gin.Context) {
	// check app run mode
	projectID := consts.OperationID
	todoListID := consts.BirthdayToDoListID

	if b.config.Env != "prod" {
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

	employees, _, err := b.store.Employee.All(b.repo.DB(), filter, model.Pagination{
		Page: 0,
		Size: 1000,
	})
	if err != nil {
		b.logger.Error(err, "get users failed")
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
			if sa != nil {
				if sa.AccountID != "" {
					discordID := sa.AccountID
					names += fmt.Sprintf("<@%s>, ", discordID)
					birthDateNames = append(birthDateNames, e.FullName)
				}

			}
		}
	}

	if len(birthDateNames) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "no birthday today"})
		return
	}

	rand.Seed(time.Now().Unix())
	msg := fmt.Sprintf(pool[rand.Intn(len(pool))], strings.TrimSuffix(names, ", "))

	//send message to Discord channel
	var discordMsg model.DiscordMessage
	discordMsg, err = b.service.Discord.PostBirthdayMsg(msg)
	if err != nil {
		b.logger.Error(err, "can not post Discord message")
		return
	}

	c.JSON(http.StatusOK, discordMsg)

	//Make Basecamp todos
	for _, birthDateNames := range birthDateNames {
		birthDayTodo := models.Todo{
			Title:   fmt.Sprintf("Prepare gift for %s, %s", birthDateNames, todayDate),
			Content: fmt.Sprintf("Prepare gift for %s, %s", birthDateNames, todayDate),
		}
		_, err := b.service.Basecamp.Todo.Create(projectID, todoListID, birthDayTodo)
		if err != nil {
			b.logger.Error(err, "Error create Basecamp todo")
			return
		}
	}
}
