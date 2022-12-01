package message_line

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/client"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

type service struct {
	client client.ClientService
	botKey string
}

func New(client client.ClientService, botKey string) MsgLineService {
	return &service{
		client: client,
		botKey: botKey,
	}
}

func (s *service) CreateMsgLine(projectID int64, campfireID int64, line string) error {
	jsonMsg, err := json.Marshal(model.CampfireLine{Content: line})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%v/%v/integrations/%v/buckets/%v/chats/%v/lines.json", model.BasecampAPIEndpoint, model.CompanyID, s.botKey, projectID, campfireID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonMsg))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	res, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}
