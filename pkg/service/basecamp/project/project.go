package project

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/client"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

type ProjectService struct {
	client client.Service
}

func NewService(client client.Service) Service {
	return &ProjectService{
		client: client,
	}
}

func (p *ProjectService) GetAll() ([]model.Project, error) {
	url := fmt.Sprintf(`https://3.basecampapi.com/%d/projects.json`, consts.CompanyBasecampID)
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

func (p *ProjectService) Get(id int) (model.Project, error) {
	var result model.Project
	url := fmt.Sprintf(`https://3.basecampapi.com/%d/projects/%d.json`, consts.CompanyBasecampID, id)
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
