package attachment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/client"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

type AttachmentService struct {
	client client.Service
}

func NewService(client client.Service) Service {
	return &AttachmentService{
		client: client,
	}
}

func (a *AttachmentService) Create(contentType, fileName string, file []byte) (string, error) {
	url := fmt.Sprintf("%v/%v/attachments.json?name=%v", model.BasecampAPIEndpoint, model.CompanyID, url.QueryEscape(fileName))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(file))
	if err != nil {
		return "", err
	}

	req.Header.Add("content-type", contentType)
	req.Header.Add("content-length", strconv.Itoa(len(file)))
	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("create attachment failed")
	}

	attachRes := &struct {
		SgID string `json:"attachable_sgid"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(attachRes); err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return attachRes.SgID, nil
}
