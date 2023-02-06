package discord

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

var (
	client = http.DefaultClient
)

type discordClient struct {
	// campfire
	webhook string
}

func New(webhook string) DiscordService {
	return &discordClient{
		webhook: webhook,
	}
}

func (d *discordClient) PostBirthdayMsg(msg string) (model.DiscordMessage, error) {
	discordMsg := model.DiscordMessage{Content: msg}
	reqByte, err := json.Marshal(discordMsg)
	if err != nil {
		return discordMsg, err
	}

	payload := bytes.NewReader(reqByte)
	if _, err := d.newRequest(http.MethodPost, d.webhook, payload); err != nil {
		return discordMsg, err
	}
	return discordMsg, nil
}

func (d *discordClient) newRequest(method string, url string, payload io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	var res *http.Response
	if res, err = client.Do(req); err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return resBody, nil
}
