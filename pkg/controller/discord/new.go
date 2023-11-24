package discord

import (
	"fmt"
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type IController interface {
	Log(in model.LogDiscordInput) error
	PublicAdvanceSalaryLog(in model.LogDiscordInput) error
}

type controller struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
	config  *config.Config
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

func (c *controller) Log(in model.LogDiscordInput) error {
	// Get discord template
	template, err := c.store.DiscordLogTemplate.GetTemplateByType(c.repo.DB(), in.Type)
	if err != nil {
		c.logger.Field("err", err.Error()).Warn("Get Discord Template failed")
		return err
	}

	data := in.Data.(map[string]interface{})

	// get employee_id in discord format if any
	if employeeID, ok := data["employee_id"]; ok {
		employee, err := c.store.Employee.One(c.repo.DB(), employeeID.(string), false)
		if err != nil {
			c.logger.Field("err", err.Error()).Warn("Get Employee failed")
			return err
		}

		accountID := employee.DisplayName
		if employee.DiscordAccount != nil && employee.DiscordAccount.DiscordID != "" {
			accountID = fmt.Sprintf("<@%s>", employee.DiscordAccount.DiscordID)
		}

		data["employee_id"] = accountID
	}

	if updatedEmployeeID, ok := data["updated_employee_id"]; ok {
		updatedEmployee, err := c.store.Employee.One(c.repo.DB(), updatedEmployeeID.(string), false)
		if err != nil {
			c.logger.Field("err", err.Error()).Warn("Get Employee failed")
			return err
		}

		accountID := updatedEmployee.DisplayName
		if updatedEmployee.DiscordAccount != nil && updatedEmployee.DiscordAccount.DiscordID != "" {
			accountID = fmt.Sprintf("<@%s>", updatedEmployee.DiscordAccount.DiscordID)
		}

		data["updated_employee_id"] = accountID
	}

	// Replace template
	content := template.Content
	for k, v := range data {
		content = strings.ReplaceAll(content, fmt.Sprintf("{{ %s }}", k), fmt.Sprintf("%v", v))
	}

	// log discord
	_, err = c.service.Discord.SendMessage(model.DiscordMessage{
		Content: content,
	}, c.config.Discord.Webhooks.AuditLog)
	if err != nil {
		c.logger.Field("err", err.Error()).Warn("Log failed")
		return err
	}

	return nil
}

func (c *controller) PublicAdvanceSalaryLog(in model.LogDiscordInput) error {
	data := in.Data.(map[string]interface{})

	icyAmount := data["icy_amount"]
	usdAmount := data["usd_amount"]

	desc := fmt.Sprintf("🧊 %v ICY (%v) has been sent to an anonymous peep as a salary advance.\n", icyAmount, usdAmount)
	desc += "\nFull-time peeps can use `?salary advance` to take a short-term credit benefit."

	embedMessage := model.DiscordMessageEmbed{
		Author:      model.DiscordMessageAuthor{},
		Title:       "💸 New ICY Payment 💸",
		URL:         "",
		Description: desc,
		Color:       1752220,
		Fields:      nil,
		Thumbnail:   model.DiscordMessageImage{},
		Image:       model.DiscordMessageImage{},
		Footer: model.DiscordMessageFooter{
			IconURL: "https://cdn.discordapp.com/avatars/564764617545482251/9c9bd4aaba164fc0b92f13f052405b4d.webp?size=160",
			Text:    "?help to see all commands",
		},
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000+07:00"),
	}

	// log discord
	_, err := c.service.Discord.SendMessage(model.DiscordMessage{
		Embeds: []model.DiscordMessageEmbed{embedMessage},
	}, c.config.Discord.Webhooks.ICYPublicLog)
	if err != nil {
		c.logger.Field("err", err.Error()).Warn("Log failed")
		return err
	}

	return nil
}
