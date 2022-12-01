package message_board

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/client"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

type service struct {
	client client.ClientService
}

func New(client client.ClientService) MsgBoardService {
	return &service{
		client: client,
	}
}

func (m *service) Create(
	projectID int64,
	msgBoardID int64,
	msg *model.Message,
) error {
	url := fmt.Sprintf(
		"%v/%v/buckets/%v/message_boards/%v/messages.json",
		model.BasecampAPIEndpoint,
		model.CompanyID,
		projectID,
		msgBoardID,
	)
	jsonMessage, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonMessage))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	res, err := m.client.Do(req)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusCreated {
		err = errors.New(string(b))
		return err
	}

	if err = json.Unmarshal(b, msg); err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}
