package discord

import (
	"fmt"
	"strings"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type IController interface {
	LogDiscord(in model.LogDiscordInput) error
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

func (c *controller) LogDiscord(in model.LogDiscordInput) error {
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
		data["employee_id"] = &employee.DiscordName
	}

	if updatedEmployeeID, ok := data["updated_employee_id"]; ok {
		updatedEmployee, err := c.store.Employee.One(c.repo.DB(), updatedEmployeeID.(string), false)
		if err != nil {
			c.logger.Field("err", err.Error()).Warn("Get Employee failed")
			return err
		}
		data["employee_id"] = &updatedEmployee.DiscordName
	}

	// Replace template
	content := template.Content
	for k, v := range data {
		content = strings.ReplaceAll(content, fmt.Sprintf("{{ %s }}", k), fmt.Sprintf("%v", v))
	}

	// log discord
	_, err = c.service.Discord.SendMessage(c.config.Discord.Webhooks.AuditLog, content)
	if err != nil {
		c.logger.Field("err", err.Error()).Warn("LogDiscord failed")
		return err
	}

	return nil
}
