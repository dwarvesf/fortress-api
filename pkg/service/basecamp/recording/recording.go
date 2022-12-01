package recording

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/client"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

type service struct {
	client client.ClientService
}

func New(client client.ClientService) RecordingService {
	return &service{
		client: client,
	}
}

func (s *service) GetFromTime(recordingType string, from time.Time) ([]model.Recording, error) {
	url := fmt.Sprintf("%v/%v/projects/recordings.json?type=%v&sort=updated_at", model.BasecampAPIEndpoint, model.CompanyID, recordingType)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var records, result []model.Recording
	if err = json.NewDecoder(res.Body).Decode(&records); err != nil {
		return nil, err
	}
	for _, record := range records {
		if record.UpdatedAt.Before(from) {
			return result, nil
		}
		result = append(result, record)
	}

	// next page
	link := res.Header.Get("Link")
	page := 2
	for link != "" {

		req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%v&page=%v", url, page), nil)
		if err != nil {
			return nil, err
		}

		res, err = s.client.Do(req)
		if err != nil {
			return nil, err
		}

		if err = json.NewDecoder(res.Body).Decode(&records); err != nil {
			return nil, err
		}
		for _, record := range records {
			if record.UpdatedAt.Before(from) {
				return result, nil
			}
			result = append(result, record)
		}

		link = res.Header.Get("Link")
		page++
	}
	defer res.Body.Close()

	return result, nil
}

func (s *service) GetEventsFromTime(
	bucketID int64,
	recordingID int64,
	from time.Time,
) ([]model.Event, error) {

	url := fmt.Sprintf("%v/%v/buckets/%v/recordings/%v/events.json", model.BasecampAPIEndpoint, model.CompanyID, bucketID, recordingID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var events, result []model.Event
	if err = json.NewDecoder(res.Body).Decode(&events); err != nil {
		return nil, err
	}
	for _, event := range events {
		if event.CreatedAt.Before(from) {
			return result, nil
		}
		result = append(result, event)
	}

	// next page
	link := res.Header.Get("Link")
	page := 2
	for link != "" {
		req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%v&page=%v", url, page), nil)
		if err != nil {
			return nil, err
		}

		res, err = s.client.Do(req)
		if err != nil {
			return nil, err
		}

		if err = json.NewDecoder(res.Body).Decode(&events); err != nil {
			return nil, err
		}
		for _, event := range events {
			if event.CreatedAt.Before(from) {
				return result, nil
			}
			result = append(result, event)
		}

		link = res.Header.Get("Link")
		page++
	}
	defer res.Body.Close()

	return result, nil
}
