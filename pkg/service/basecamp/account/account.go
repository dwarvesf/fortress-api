package account

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/client"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/project"
)

var (
	errNotInProject = errors.New("This account does not belong to this project")
)

type service struct {
	client client.ClientService
}

func New(client client.ClientService) AccountService {
	return &service{
		client: client,
	}
}

// Create creates Basecamp account and add to Woodland
func (s *service) Create(
	name string,
	email string,
	org string,
) (int64, string, error) {
	woodlandEntry := model.PeopleEntry{
		Create: []model.PeopleCreate{
			{
				Name:         name,
				EmailAddress: email,
				CompanyName:  org,
			},
		},
	}

	return s.updateInProject(model.BasecampWoodlandID, woodlandEntry)
}

func (s *service) Get(bcID int) (*model.Person, error) {
	url := fmt.Sprintf("https://3.basecampapi.com/%d/people/%d.json", model.CompanyBasecampID, bcID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		err = fmt.Errorf("can not get Basecamp account: %s", err.Error())
		return nil, err
	}
	res, err := s.client.Do(req)
	if err != nil {
		err = fmt.Errorf("can not create new request: %s", err.Error())
		return nil, err
	}
	defer res.Body.Close()

	var person model.Person
	if err := json.NewDecoder(res.Body).Decode(&person); err != nil {
		err := fmt.Errorf("can not parse Basecamp account: %s", err.Error())
		return nil, err
	}
	return &person, nil
}

func (s *service) updateInProject(
	projectID int64,
	peopleEntry model.PeopleEntry,
) (int64, string, error) {
	// update user
	res, err := s.updateAccount(projectID, peopleEntry)
	if err != nil {
		return 0, "", err
	}

	// parse information
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
		return 0, "", errNotInProject
	}

	if len(response.Granted) == 0 {
		return response.Revoked[0].ID, response.Revoked[0].SgID, nil
	}

	return response.Granted[0].ID, response.Granted[0].SgID, nil
}

func (s *service) updateAccount(projectID int64, peopleEntry model.PeopleEntry) (*http.Response, error) {
	url := fmt.Sprintf(
		"%v/%v/projects/%v/people/users.json",
		model.BasecampAPIEndpoint,
		model.CompanyID,
		projectID,
	)

	bData, err := json.Marshal(peopleEntry)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(bData))
	if err != nil {
		return nil, err
	}

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return res, nil
}

func (s *service) UpdateInProject(projectID int64, peopleEntry model.PeopleEntry) (int64, string, error) {
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

	res, err := s.client.Do(req)
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
		return 0, "", errNotInProject
	}
	if len(response.Granted) == 0 {
		return response.Revoked[0].ID, response.Revoked[0].SgID, nil
	}

	return response.Granted[0].ID, response.Granted[0].SgID, nil
}

func (s *service) Remove(userID int64) error {
	projects, err := project.New(s.client).GetAll()
	if err != nil {
		return err
	}
	peopleEntry := model.PeopleEntry{
		Revoke: []int64{userID},
	}
	for i := range projects {
		_, _, err = s.UpdateInProject(projects[i].ID, peopleEntry)
		if err != nil && err != errNotInProject {
			return err
		}
	}
	return nil
}
