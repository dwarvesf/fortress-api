package model

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"
)

// BasecampWebhookMessage is a structure display basecamp webhook message
type BasecampWebhookMessage struct {
	Kind      string            `json:"kind,omitempty"`
	Recording BasecampRecording `json:"recording,omitempty"`
	Creator   BasecampUser      `json:"creator,omitempty"`
}

func (msg *BasecampWebhookMessage) Decode(body []byte) error {
	return json.Unmarshal(body, &msg)
}

func (msg *BasecampWebhookMessage) Read(rc io.ReadCloser) []byte {
	defer rc.Close()
	body, _ := ioutil.ReadAll(rc)
	return body
}

// isOperationComplete return true when parent (Todolist) title contain "Operations" example title ("Operations | July 2019")
func (req *BasecampWebhookMessage) IsOperationComplete() bool {
	split := strings.Split(
		strings.Replace(req.Recording.Parent.Title, " ", "", -1),
		"|",
	)
	if len(split) < 2 {
		return false
	}
	if strings.ToLower(split[0]) != "operations" {
		return false
	}

	return true
}

func (req *BasecampWebhookMessage) IsExpenseComplete() bool {
	pt := req.Recording.Parent.Title
	if len(pt) < 8 || strings.ToLower(pt[:8]) != "expenses" {
		return false
	}

	return true
}
