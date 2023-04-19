package schedule

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/client"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

type ScheduleService struct {
	client client.Service
	logger logger.Logger
}

func NewService(client client.Service, logger logger.Logger) Service {
	return &ScheduleService{
		client: client,
		logger: logger,
	}
}

func (s *ScheduleService) CreateScheduleEntry(projectID int64, scheduleID int64, scheduleEntry model.ScheduleEntry) (*model.ScheduleEntry, error) {
	url := fmt.Sprintf("%v/%v/buckets/%v/schedules/%v/entries.json", model.BasecampAPIEndpoint, model.CompanyID, projectID, scheduleID)
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

	b, err := io.ReadAll(resp.Body)
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

func (s *ScheduleService) GetScheduleEntries(projectID, scheduleID int64) ([]*model.ScheduleEntry, error) {
	url := fmt.Sprintf("%v/%v/buckets/%v/schedules/%v/entries.json", model.BasecampAPIEndpoint, model.CompanyID, projectID, scheduleID)

	res := []*model.ScheduleEntry{}
	get := func(page int) (bool, error) {
		req, err := http.NewRequest("GET", fmt.Sprintf("%v?page=%v", url, page), nil)
		if err != nil {
			s.logger.AddField("url", url).Error(err, "failed to create req")
			return false, err
		}
		resp, err := s.client.Do(req)
		if err != nil {
			s.logger.Error(err, "failed to send req")
			return false, err
		}
		defer resp.Body.Close()

		if resp.StatusCode/100 > 2 {
			err = fmt.Errorf("failed to get schedule entries")
			s.logger.Fields(logger.Fields{
				"StatusCode": resp.StatusCode,
				"ProjectID":  projectID,
				"ScheduleID": scheduleID,
			}).Error(err, "request failed")
			return false, err
		}

		entries, err := responseToScheduleEntries(resp)
		if err != nil {
			return false, err
		}
		res = append(res, entries...)

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

func responseToScheduleEntries(res *http.Response) ([]*model.ScheduleEntry, error) {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var scheduleEntries []*model.ScheduleEntry
	err = json.Unmarshal(body, &scheduleEntries)
	if err != nil {
		logger.L.AddField("body", string(body)).Error(err, "failed to unmarshal body")
		return nil, err
	}

	return scheduleEntries, nil
}

func (s *ScheduleService) UpdateSheduleEntry(projectID int64, se *model.ScheduleEntry) error {
	url := fmt.Sprintf("%v/%v/buckets/%v/schedule_entries/%v.json", model.BasecampAPIEndpoint, model.CompanyID, projectID, se.ID)
	jsonTodo, err := json.Marshal(se)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonTodo))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return err
}
