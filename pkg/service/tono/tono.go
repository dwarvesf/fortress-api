package tono

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

type IService interface {
	GetGuildUserProfile(profileId, guildId string) (*GuildProfile, error)
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

func (m *client) GetGuildUserProfile(profileId, guildId string) (*GuildProfile, error) {
	var client = &http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("%s/api/v1/users/profiles?profile_id=%s&guild_id=%s", m.cfg.Tono.BaseURL, profileId, guildId)
	r, err := client.Get(url)
	if err != nil {
		m.l.Errorf(err, "[tono.GetGuildUserProfile] client.Get failed")
		return nil, err
	}
	defer r.Body.Close()

	res := &GetGuildProfileResponse{}
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		m.l.Errorf(err, "[tono.GetGuildUserProfile] decoder.Decode failed")
		return nil, err
	}

	return &res.Data, handleErrorStatusCode("Get Guild User Profile", r.StatusCode)
}

func handleErrorStatusCode(method string, statusCode int) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}

	return fmt.Errorf("%s status code: %d", method, statusCode)
}
