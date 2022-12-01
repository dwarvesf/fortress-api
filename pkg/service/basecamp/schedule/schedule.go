package schedule

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/client"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

type service struct {
	client client.ClientService
}

func New(c client.ClientService) ScheduleService {
	return &service{
		client: c,
	}
}

func (s *service) CreateScheduleEntry(projectID int64, scheduleID int64, scheduleEntry model.ScheduleEntry) (*model.ScheduleEntry, error) {
	url := fmt.Sprintf("%v/%v/buckets/%v/schedules/%v/entries.json?notify=true", model.BasecampAPIEndpoint, model.CompanyID, projectID, scheduleID)
	jsonTodo, err := json.Marshal(scheduleEntry)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonTodo))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	res := &model.ScheduleEntry{}
	err = json.Unmarshal(b, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}
