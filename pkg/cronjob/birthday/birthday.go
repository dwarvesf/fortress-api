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
	"github.com/dwarvesf/fortress-api/pkg/store"

	"github.com/gin-gonic/gin"
)

type birthday struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
	repo    store.DBRepo
	config  *config.Config
}

// func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) ICronjob {
// 	return &birthday{store: store, repo: repo, service: service, logger: logger, config: cfg}
// }

func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) ICronjob {
	return &birthday{store: store, repo: repo, service: service, logger: logger, config: cfg}
}

var check = false

func (b *birthday) BirthdayDailyMessage(c *gin.Context) {
	funcName := "BirthdayDailyMessage"
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

	employees, err := b.store.Employee.GetAllActive(b.repo.DB())
	if err != nil {
		b.logger.Error(err, fmt.Sprintf("%s get users failed", funcName))
		return
	}

	// format message if there is user's birthday
	var names string
	for _, employee := range employees {
		now := time.Now()
		if now.Day() == employee.DateOfBirth.Day() && now.Month() == employee.DateOfBirth.Month() {
			names += fmt.Sprintf("<@%s>, ", employee.DiscordID)
		}
	}

	rand.Seed(time.Now().Unix())
	msg := fmt.Sprintf(pool[rand.Intn(len(pool))], strings.TrimSuffix(names, ", "))

	// send message to Discord channel
	var discordMsg model.DiscordMessage
	discordMsg, err = b.service.Discord.PostBirthdayMsg(msg)
	if err != nil {
		b.logger.Error(err, fmt.Sprintf("%s can not post Discord message", funcName))
		return
	}

	c.JSON(http.StatusOK, discordMsg)
}
