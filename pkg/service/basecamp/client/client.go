package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
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

type client struct {
	client   *http.Client
	basecamp config.Basecamp
}

func NewClient(bc config.Basecamp) (ClientService, error) {
	newToken, err := accesstoken(bc)
	if err != nil {
		return nil, err
	}
	return &client{
		client:   basecampOauthConfig.Client(context.Background(), newToken),
		basecamp: bc,
	}, nil
}

func accesstoken(bc config.Basecamp) (*oauth2.Token, error) {
	basecampOauthConfig.ClientID = bc.ClientID
	basecampOauthConfig.ClientSecret = bc.ClientSecret

	refreshToken := bc.RefreshToken
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

func (c *client) Get(url string) (resp *http.Response, err error) {
	return c.client.Get(url)
}

func (c *client) Do(req *http.Request) (resp *http.Response, err error) {
	return c.client.Do(req)
}

// GetAccessToken return Basecamp AccessToken
func (c *client) GetAccessToken(code, redirectURI string) (string, error) {
	c.basecamp.RedirectURI = redirectURI
	url := fmt.Sprintf(model.BasecampEndpoint+"/token?type=web_server&client_id=%v&redirect_uri=%v&client_secret=%v&code=%v", c.basecamp.ClientID, redirectURI, c.basecamp.ClientSecret, code)

	r, err := http.Post(url, "application/json", nil)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(r.Body)
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
