package people

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/client"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/project"
)

type PeopleService struct {
	client client.Service
}

func NewService(client client.Service) Service {
	return &PeopleService{
		client: client,
	}
}

func (p *PeopleService) GetByID(id int) (*model.Person, error) {
	url := fmt.Sprintf(`https://3.basecampapi.com/%d/people/%d.json`, consts.CompanyBasecampID, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	res, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var person model.Person
	if err := json.NewDecoder(res.Body).Decode(&person); err != nil {
		return nil, err
	}

	return &person, nil
}

// GetUserInfo get UserInfo func
func (p *PeopleService) GetInfo() (*model.UserInfo, error) {
	req, err := http.NewRequest("GET", model.GetBasecampUserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var result model.UserInfo
	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (p *PeopleService) Create(name, email, organization string) (id int64, sgID string, err error) {
	woodlandEntry := model.PeopleEntry{
		Create: []model.PeopleCreate{
			{
				Name:         name,
				EmailAddress: email,
				CompanyName:  organization,
			},
		},
	}
	return p.UpdateInProject(consts.WoodlandID, woodlandEntry)
}

func (p *PeopleService) UpdateInProject(projectID int64, peopleEntry model.PeopleEntry) (int64, string, error) {
	url := fmt.Sprintf("%v/%v/projects/%v/people/users.json", model.BasecampAPIEndpoint, model.CompanyID, projectID)
	jsonAdd, err := json.Marshal(peopleEntry)
	if err != nil {
		return 0, "", err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonAdd))
	if err != nil {
		return 0, "", err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := p.client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer res.Body.Close()

	response := struct {
		Granted []struct {
			ID   int64  `json:"id"`
			SgID string `json:"attachable_sgid"`
		} `json:"granted"`
		Revoked []struct {
			ID   int64  `json:"id"`
			SgID string `json:"attachable_sgid"`
		} `json:"revoked"`
	}{}
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return 0, "", err
	}
	if len(response.Granted) == 0 && len(response.Revoked) == 0 {
		return 0, "", ErrNotInProject
	}
	if len(response.Granted) == 0 {
		return response.Revoked[0].ID, response.Revoked[0].SgID, nil
	}

	return response.Granted[0].ID, response.Granted[0].SgID, nil
}

func (p *PeopleService) Remove(userID int64) error {
	projects, err := project.NewService(p.client).GetAll()
	if err != nil {
		return err
	}
	peopleEntry := model.PeopleEntry{
		Revoke: []int64{userID},
	}
	for i := range projects {
		_, _, err = p.UpdateInProject(projects[i].ID, peopleEntry)
		if err != nil && err != ErrNotInProject {
			return err
		}
	}
	return nil
}

func (p *PeopleService) GetAllOnProject(projectID int) ([]model.Person, error) {
	url := fmt.Sprintf(`https://3.basecampapi.com/%d/projects/%d/people.json`, consts.CompanyBasecampID, projectID)

	var people []model.Person
	get := func(page int) (bool, error) {
		req, err := http.NewRequest("GET", fmt.Sprintf("%v?page=%v", url, page), nil)
		if err != nil {
			return false, err
		}
		res, err := p.client.Do(req)
		if err != nil {
			return false, err
		}
		defer res.Body.Close()

		var p []model.Person
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return false, err
		}
		if err := json.Unmarshal(b, &p); err != nil {
			return false, err
		}
		people = append(people, p...)

		return (res.Header.Get("Link") != ""), nil
	}

	page := 1
	for {
		morePage, err := get(page)
		if err != nil {
			return nil, err
		}
		if !morePage {
			return people, nil
		}
		page++
	}
}
