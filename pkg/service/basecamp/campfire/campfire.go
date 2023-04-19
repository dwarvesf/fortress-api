package campfire

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/client"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

type CampfireService struct {
	client client.Service
	logger logger.Logger
	cfg    *config.Config
}

func NewService(client client.Service, logger logger.Logger, cfg *config.Config) Service {
	return &CampfireService{
		client: client,
		logger: logger,
		cfg:    cfg,
	}
}

func (c *CampfireService) CreateLine(projectID, campfireID int, line string) error {
	url := fmt.Sprintf("%v/%v/buckets/%v/chats/%v/lines.json", model.BasecampAPIEndpoint, model.CompanyID, projectID, campfireID)

	jsonMessage, err := json.Marshal(model.CampfireLine{Content: line})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonMessage))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

func (c *CampfireService) BotCreateLine(projectID, campfireID int, line string) error {
	url := fmt.Sprintf("%v/%v/integrations/%v/buckets/%v/chats/%v/lines.json", model.BasecampAPIEndpoint, model.CompanyID, c.cfg.Basecamp.BotKey, projectID, campfireID)

	jsonMessage, err := json.Marshal(model.CampfireLine{Content: line})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonMessage))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

func (c *CampfireService) BotReply(callbackURL string, message string) error {
	campfireMessage := model.CampfireLine{Content: message}
	jsonMessage, err := json.Marshal(campfireMessage)
	if err != nil {
		c.logger.AddField("message", campfireMessage).Error(err, "failed to marshal message")
		return err
	}
	resp, err := http.Post(callbackURL, "application/json", bytes.NewBuffer(jsonMessage))
	if err != nil {
		c.logger.Fields(logger.Fields{"message": campfireMessage, "url": callbackURL}).Error(err, "failed to send request")
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(err, "read response body failed")
		return err
	}
	if resp.StatusCode/100 > 2 {
		err = fmt.Errorf(string(data))
		c.logger.Fields(logger.Fields{"code": resp.StatusCode}).Error(err, "request failed")
		return err
	}

	return nil
}
