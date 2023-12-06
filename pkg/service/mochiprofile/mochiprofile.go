package mochiprofile

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

type IService interface {
	GetProfile(id string) (*MochiProfile, error)
	GetProfileByDiscordID(discordID string) (*MochiProfile, error)
}

type client struct {
	cfg *config.Config
	l   logger.Logger
}

func New(cfg *config.Config, l logger.Logger) IService {
	return &client{
		cfg: cfg,
		l:   l,
	}
}

func (m *client) GetProfile(id string) (*MochiProfile, error) {
	var client = &http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("%s/api/v1/profiles?ids=%s", m.cfg.MochiProfile.BaseURL, id)
	r, err := client.Get(url)
	if err != nil {
		m.l.Errorf(err, "[mochipay.GetListTransaction] client.Get failed")
		return nil, err
	}
	defer r.Body.Close()

	res := &GetMochiProfilesResponse{}
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		m.l.Errorf(err, "[mochipay.GetListTransaction] decoder.Decode failed")
		return nil, err
	}

	if len(res.Data) == 0 {
		return nil, nil
	}

	if len(res.Data) > 1 {
		m.l.Errorf(nil, "[mochipay.GetListTransaction] more than 1 profile")
		return nil, fmt.Errorf("more than 1 profile")
	}

	return &res.Data[0], handleErrorStatusCode("Get Mochi profile", r.StatusCode)
}

func (m *client) GetProfileByDiscordID(discordID string) (*MochiProfile, error) {
	var client = &http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("%s/api/v1/profiles/get-by-discord/%s", m.cfg.MochiProfile.BaseURL, discordID)
	r, err := client.Get(url)
	if err != nil {
		m.l.Errorf(err, "[mochipay.GetListTransaction] client.Get failed")
		return nil, err
	}
	defer r.Body.Close()

	res := &MochiProfile{}
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		m.l.Errorf(err, "[mochipay.GetListTransaction] decoder.Decode failed")
		return nil, err
	}

	return res, handleErrorStatusCode("Get Mochi profile by discord ID", r.StatusCode)
}

func handleErrorStatusCode(method string, statusCode int) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}

	return fmt.Errorf("%s status code: %d", method, statusCode)
}
