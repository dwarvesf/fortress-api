package mb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/client"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

type messageBoard struct {
	client client.Service
}

// NewService -- create new message board service
func NewService(client client.Service) Service {
	return &messageBoard{
		client: client,
	}
}

func (m *messageBoard) Create(message *model.Message, projectID int, messageBoardID int) error {
	url := fmt.Sprintf("%v/%v/buckets/%v/message_boards/%v/messages.json", model.BasecampAPIEndpoint, model.CompanyID, projectID, messageBoardID)
	jsonMessage, err := json.Marshal(message)
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

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusCreated {
		err = errors.New(string(b))
		return err
	}

	if err = json.Unmarshal(b, message); err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

func (m *messageBoard) Get(projectID int, messageID int) (message model.Message, err error) {
	res := model.Message{}
	url := fmt.Sprintf("%v/%v/buckets/%v/messages/%v.json", model.BasecampAPIEndpoint, model.CompanyID, projectID, messageID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return res, err
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return res, err
	}
	if resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return res, err
		}
		return res, errors.New(string(b))
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return res, err
	}

	return res, nil
}

func (m *messageBoard) GetList(projectID int, messageBoardID int) ([]model.Message, error) {
	url := fmt.Sprintf("%v/%v/buckets/%v/message_boards/%v/messages.json", model.BasecampAPIEndpoint, model.CompanyID, projectID, messageBoardID)

	res := []model.Message{}
	get := func(page int) (bool, error) {
		req, err := http.NewRequest("GET", fmt.Sprintf("%v?page=%v", url, page), nil)
		if err != nil {
			return false, err
		}
		resp, err := m.client.Do(req)
		if err != nil {
			return false, err
		}
		if resp.StatusCode != http.StatusOK {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return false, err
			}
			return false, errors.New(string(b))
		}
		defer resp.Body.Close()

		var msg []model.Message
		if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
			return false, err
		}
		res = append(res, msg...)

		return (resp.Header.Get("Link") != ""), nil
	}

	page := 1
	for {
		morePage, err := get(page)
		if err != nil {
			return nil, err
		}
		if !morePage {
			return res, nil
		}
		page++
	}
}
