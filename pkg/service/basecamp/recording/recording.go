package recording

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/client"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

type RecordingService struct {
	client client.Service
}

func NewService(client client.Service) Service {
	return &RecordingService{
		client: client,
	}
}

func (r *RecordingService) Trash(projectID, recordingID string) error {
	url := fmt.Sprintf("%v/%v/buckets/%v/recordings/%v/status/trashed.json", model.BasecampAPIEndpoint, model.CompanyID, projectID, recordingID)

	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return err
	}

	res, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

func (r *RecordingService) TryToGetInvoiceImageURL(url string) (string, error) {
	resp, err := r.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	tmp := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&tmp)
	if err != nil {
		return "", err
	}

	strI, ok := tmp["description"]
	if !ok {
		return "", nil
	}
	str, ok := strI.(string)
	if !ok {
		return "", nil
	}

	return ensureToGetImageURLFromRawHTML(str), nil
}

func ensureToGetImageURLFromRawHTML(raw string) string {
	doc, err := htmlquery.Parse(strings.NewReader(raw))
	if err != nil {
		return ""
	}

	// try to get first img url from raw html
	node := htmlquery.FindOne(doc, "//img")
	if node == nil {
		return ""
	}
	for _, v := range node.Attr {
		if v.Key == "src" {
			return v.Val
		}
	}

	return ""
}

func (r *RecordingService) Archive(projectID, recordingID int) error {
	url := fmt.Sprintf("%v/%v/buckets/%d/recordings/%d/status/archived.json", model.BasecampAPIEndpoint, model.CompanyID, projectID, recordingID)

	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return err
	}

	res, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

func (r *RecordingService) GetFrom(from time.Time, recordingType string) ([]model.Recording, error) {
	url := fmt.Sprintf("%v/%v/projects/recordings.json?type=%v&sort=updated_at", model.BasecampAPIEndpoint, model.CompanyID, recordingType)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var recordings, result []model.Recording
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}
	for _, recording := range recordings {
		if recording.UpdatedAt.Before(from) {
			return result, nil
		}
		result = append(result, recording)
	}
	link := res.Header.Get("Link")
	page := 2
	for link != "" {
		var request *http.Request
		request, err = http.NewRequest("GET", fmt.Sprintf("%v&page=%v", url, page), nil)
		if err != nil {
			return nil, err
		}

		var response *http.Response
		response, err = r.client.Do(request)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		ss := []model.Recording{}
		if err := json.NewDecoder(response.Body).Decode(&ss); err != nil {
			return nil, err
		}
		for _, recording := range ss {
			if recording.UpdatedAt.Before(from) {
				return result, nil
			}
			result = append(result, recording)
		}

		link = response.Header.Get("Link")
		page++
	}

	return result, nil
}

func (r *RecordingService) GetEvents(from time.Time, projectID, recordingID int) ([]model.Event, error) {
	url := fmt.Sprintf("%v/%v/buckets/%v/recordings/%v/events.json", model.BasecampAPIEndpoint, model.CompanyID, projectID, recordingID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var events, result []model.Event
	if err := json.NewDecoder(res.Body).Decode(&events); err != nil {
		return nil, err
	}
	for _, event := range events {
		if event.CreatedAt.Before(from) {
			return result, nil
		}
		result = append(result, event)
	}
	link := res.Header.Get("Link")
	page := 2
	for link != "" {
		var request *http.Request
		request, err = http.NewRequest("GET", fmt.Sprintf("%v&page=%v", url, page), nil)
		if err != nil {
			return nil, err
		}

		var response *http.Response
		response, err = r.client.Do(request)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		ss := []model.Event{}
		if err := json.NewDecoder(response.Body).Decode(&ss); err != nil {
			return nil, err
		}
		for _, event := range ss {
			if event.CreatedAt.Before(from) {
				return result, nil
			}
			result = append(result, event)
		}

		link = response.Header.Get("Link")
		page++
	}

	return result, nil
}
