package mochiprofile

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

type IService interface {
	GetProfile(id string) (*MochiProfile, error)
	GetListProfiles(req ListProfilesRequest) (*GetMochiProfilesResponse, error)
	GetProfileByDiscordID(discordID string) (*MochiProfile, error)
	GetProfileByEvmAddress(address string) (*MochiProfile, error)
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
		m.l.Errorf(err, "[mochiprofile.GetProfile] decoder.Decode failed")
		return nil, err
	}

	if len(res.Data) == 0 {
		return nil, nil
	}

	if len(res.Data) > 1 {
		m.l.Errorf(nil, "[mochiprofile.GetProfile] more than 1 profile")
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

func (m *client) GetProfileByEvmAddress(address string) (*MochiProfile, error) {
	var client = &http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("%s/api/v1/profiles/get-by-evm/%s", m.cfg.MochiProfile.BaseURL, address)
	r, err := client.Get(url)
	if err != nil {
		m.l.Errorf(err, "[mochiprofile.GetProfileByEvmAddress] client.Get failed")
		return nil, err
	}
	defer r.Body.Close()

	res := &MochiProfile{}
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		m.l.Errorf(err, "[mochiprofile.GetProfileByEvmAddress] decoder.Decode failed")
		return nil, err
	}

	return res, handleErrorStatusCode("Get Mochi profile by evm address", r.StatusCode)
}

func (m *client) GetListProfiles(req ListProfilesRequest) (*GetMochiProfilesResponse, error) {
	var client = &http.Client{
		Timeout: 10 * time.Second,
	}

	queryParams := url.Values{}
	var pageSize int64 = 10
	if req.Size != 0 {
		pageSize = req.Size
	}
	queryParams.Add("size", fmt.Sprintf("%v", pageSize))
	queryParams.Add("page", fmt.Sprintf("%v", req.Page))

	if len(req.IDs) > 0 {
		ids := getListParams(req.IDs)
		queryParams.Add("ids", ids)
	}

	if len(req.Types) > 0 {
		types := getListParams(req.Types)
		queryParams.Add("types", types)
	}

	url := fmt.Sprintf("%s/api/v1/profiles?", m.cfg.MochiProfile.BaseURL) + queryParams.Encode()
	r, err := client.Get(url)
	if err != nil {
		m.l.Errorf(err, "[mochiprofile.GetListProfile] client.Get failed")
		return nil, err
	}
	defer r.Body.Close()

	res := &GetMochiProfilesResponse{}
	if err := json.NewDecoder(r.Body).Decode(res); err != nil {
		m.l.Errorf(err, "[mochiprofile.GetListProfiles] decoder.Decode failed")
		return nil, err
	}

	return res, nil
}

func handleErrorStatusCode(method string, statusCode int) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}

	return fmt.Errorf("%s status code: %d", method, statusCode)
}

func getListParams[T fmt.Stringer](data []T) string {
	param := ""
	for i, a := range data {
		if i == 0 {
			param += fmt.Sprintf("%s", a)
			continue
		}
		param += fmt.Sprintf("|%s", a)
	}

	return param
}
