package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"golang.org/x/oauth2"
)

var (
	basecampOauthConfig = &oauth2.Config{
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://launchpad.37signals.com/authorization/new?type=refresh",
			TokenURL:  "https://launchpad.37signals.com/authorization/token?type=refresh",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
)

// Client --
type Client struct {
	client   *http.Client
	basecamp *model.Basecamp
}

// NewClient --
func NewClient(bc *model.Basecamp, cfg *config.Config) (Service, error) {
	newToken, err := accesstoken(cfg)
	if err != nil {
		logger.L.Error(err, "can't init basecamp service")
		return nil, err
	}

	return &Client{
		client:   basecampOauthConfig.Client(context.Background(), newToken),
		basecamp: bc,
	}, nil
}

func accesstoken(cfg *config.Config) (*oauth2.Token, error) {
	basecampOauthConfig.ClientID = cfg.Basecamp.ClientID
	basecampOauthConfig.ClientSecret = cfg.Basecamp.ClientSecret

	refreshToken := cfg.Basecamp.OAuthRefreshToken
	if refreshToken == "" {
		return nil, errors.New("missing basecampapp_oauth_refresh_token env variable")
	}
	token := new(oauth2.Token)
	token.RefreshToken = refreshToken
	newToken, err := basecampOauthConfig.TokenSource(context.Background(), token).Token()
	if err != nil {
		return nil, err
	}

	return newToken, nil
}

// func (c *Client) intervalRefreshToken() {
// 	interval := time.Tick(15 * time.Minute)
// 	for {
// 		<-interval
// 		newToken, _ := accesstoken()
// 		c.client = basecampOauthConfig.Client(oauth2.NoContext, newToken)
// 	}
// }

func (c *Client) Get(url string) (resp *http.Response, err error) {
	return c.client.Get(url)
}

func (c *Client) Do(req *http.Request) (resp *http.Response, err error) {
	return c.client.Do(req)
}

// New --
func New(clientID, clientSecret string) *model.Basecamp {
	return &model.Basecamp{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
}

// GetAccessToken return Basecamp AccessToken
func (c *Client) GetAccessToken(code, redirectURI string) (string, error) {
	c.basecamp.RedirectURI = redirectURI
	url := fmt.Sprintf(model.BasecampEndpoint+"/token?type=web_server&client_id=%v&redirect_uri=%v&client_secret=%v&code=%v", c.basecamp.ClientID, redirectURI, c.basecamp.ClientSecret, code)

	r, err := http.Post(url, "application/json", nil)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()

	var resp model.AuthenticationResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		return "", err
	}

	return resp.AccessToken, nil
}
