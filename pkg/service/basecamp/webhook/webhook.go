package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/client"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

type WebhookService struct {
	client client.Service
}

func NewService(client client.Service) Service {
	return &WebhookService{
		client: client,
	}
}

func (w *WebhookService) FindWebHook(projectID int, hookID int) (*model.Hook, error) {
	url := fmt.Sprintf("%v/%v/buckets/%v/webhooks/%v.json", model.BasecampAPIEndpoint, model.CompanyID, projectID, hookID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	rs := model.Hook{}
	if err := json.NewDecoder(res.Body).Decode(&rs); err != nil {
		return nil, err
	}
	return &rs, nil
}

func (w *WebhookService) UpdateWebHook(projectID, hookID int, hookBody model.Hook) error {
	url := fmt.Sprintf("%v/%v/buckets/%v/webhooks/%v.json", model.BasecampAPIEndpoint, model.CompanyID, projectID, hookID)

	jsonHook, err := json.Marshal(hookBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonHook))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	res, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}
