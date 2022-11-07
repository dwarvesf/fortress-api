package google

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	state                        = "state-token"
	getGoogleUserInfoAPIEndpoint = "https://www.googleapis.com/plus/v1/people/me"
)

type Google struct {
	Config *oauth2.Config
}

// New function return Google service
func New(ClientID, ClientSecret, AppName string, Scopes []string) *Google {
	Config := &oauth2.Config{
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       Scopes,
	}

	return &Google{
		Config: Config,
	}
}

// GetLoginURL return url for user loggin to google account
func (g *Google) GetLoginURL() string {
	authURL := g.Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return authURL
}

// GetAccessToken return google access token
func (g *Google) GetAccessToken(code string, redirectURL string) (string, error) {
	g.Config.RedirectURL = redirectURL
	token, err := g.Config.Exchange(context.Background(), code)
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

// GetGoogleUserInfo return google user info
func (g *Google) GetGoogleEmail(accessToken string) (email string, err error) {
	var gu struct {
		DisplayName string `json:"displayName"`
		ID          string `json:"id"`
		Emails      []struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"emails"`
	}

	response, err := http.Get(getGoogleUserInfoAPIEndpoint + "?access_token=" + accessToken)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	if err = json.Unmarshal([]byte(body), &gu); err != nil {
		return "", err
	}

	var primaryEmail string
	for i := range gu.Emails {
		if strings.ToLower(gu.Emails[i].Type) == "account" || gu.Emails[i].Type == "primary" {
			primaryEmail = gu.Emails[i].Value
			break
		}
	}
	return primaryEmail, nil
}
