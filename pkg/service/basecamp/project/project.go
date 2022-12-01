package project

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/client"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

type service struct {
	client client.ClientService
}

func New(client client.ClientService) ProjectService {
	return &service{
		client: client,
	}
}

func (p *service) GetAll() ([]model.Project, error) {
	url := fmt.Sprintf(`https://3.basecampapi.com/%d/projects.json`, model.CompanyBasecampID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	res, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var result []model.Project
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (p *service) Get(id int) (model.Project, error) {
	var result model.Project
	url := fmt.Sprintf(`https://3.basecampapi.com/%d/projects/%d.json`, model.CompanyBasecampID, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return result, err
	}

	res, err := p.client.Do(req)
	if err != nil {
		return result, err
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return result, err
	}

	return result, nil
}
