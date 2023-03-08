package subscription

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/client"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

type SubscriptionService struct {
	client client.Service
}

func NewService(client client.Service) Service {
	return &SubscriptionService{
		client: client,
	}
}

func (s *SubscriptionService) Subscribe(url string, list *model.SubscriptionList) error {
	json, err := json.Marshal(list)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(json))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
