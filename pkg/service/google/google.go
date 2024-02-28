package google

import (
	"context"
	"encoding/json"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"strings"
)

const (
	state                              = "state-token"
	getGoogleUserInfoAPIEndpointLegacy = "https://www.googleapis.com/plus/v1/people/me"
	getGoogleUserInfoAPIEndpoint       = "https://people.googleapis.com/v1/people/me"
)

type googleService struct {
	config *oauth2.Config
}

// New function return Google service
func New(config *oauth2.Config) (IService, error) {
	return &googleService{
		config: config,
	}, nil
}

// GetLoginURL return url for user loggin to google account
func (g *googleService) GetLoginURL() string {
	authURL := g.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return authURL
}

// GetAccessToken return google access token
func (g *googleService) GetAccessToken(code string, redirectURL string) (string, error) {
	g.config.RedirectURL = redirectURL
	token, err := g.config.Exchange(context.Background(), code)
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

// GetGoogleEmailLegacy return google user info
func (g *googleService) GetGoogleEmailLegacy(accessToken string) (email string, err error) {
	var gu struct {
		DisplayName string `json:"displayName"`
		ID          string `json:"id"`
		Emails      []struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"emails"`
	}

	response, err := http.Get(getGoogleUserInfoAPIEndpointLegacy + "?access_token=" + accessToken)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	if err = json.Unmarshal(body, &gu); err != nil {
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

// GetGoogleEmail return google user info
func (g *googleService) GetGoogleEmail(accessToken string) (email string, err error) {
	var gu struct {
		DisplayName string `json:"displayName"`
		ID          string `json:"id"`
		Emails      []struct {
			Metadata struct {
				Primary       bool `json:"primary"`
				Verified      bool `json:"verified"`
				SourcePrimary bool `json:"sourcePrimary"`
			} `json:"metadata"`
			Value string `json:"value"`
		} `json:"emailAddresses"`
	}

	response, err := http.Get(getGoogleUserInfoAPIEndpoint + "?&personFields=emailAddresses&access_token=" + accessToken)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	if err = json.Unmarshal(body, &gu); err != nil {
		return "", err
	}

	var primaryEmail string
	for i := range gu.Emails {
		if gu.Emails[i].Metadata.SourcePrimary {
			primaryEmail = gu.Emails[i].Value
			break
		}
	}

	return primaryEmail, nil
}
