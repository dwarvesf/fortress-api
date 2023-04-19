package comment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/client"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

type CommentService struct {
	client client.Service
}

func NewService(client client.Service) Service {
	return &CommentService{
		client: client,
	}
}

func (c *CommentService) Create(projectID int, recordingID int, comment *model.Comment) error {
	url := fmt.Sprintf("%v/%v/buckets/%v/recordings/%v/comments.json", model.BasecampAPIEndpoint, model.CompanyID, projectID, recordingID)
	jsonGroup, err := json.Marshal(comment)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonGroup))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	res.Body.Close()
	return nil
}

func (c *CommentService) Gets(projectID int, recordingID int) ([]model.Comment, error) {
	comments := []model.Comment{}
	url := fmt.Sprintf("%v/%v/buckets/%v/recordings/%v/comments.json", model.BasecampAPIEndpoint, model.CompanyID, projectID, recordingID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return comments, err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return comments, err
	}
	defer res.Body.Close()

	if err = json.NewDecoder(res.Body).Decode(&comments); err != nil {
		return comments, err
	}
	link := res.Header.Get("Link")
	page := 2
	for link != "" {
		var request *http.Request
		request, err = http.NewRequest("GET", fmt.Sprintf("%v?page=%v", url, page), nil)
		if err != nil {
			return nil, err
		}
		request.Header.Add("Content-Type", "application/json")

		var response *http.Response
		response, err = c.client.Do(request)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		ss := []model.Comment{}
		if err := json.NewDecoder(response.Body).Decode(&ss); err != nil {
			return nil, err
		}
		comments = append(comments, ss...)

		link = response.Header.Get("Link")
		page++
	}

	return comments, nil
}
