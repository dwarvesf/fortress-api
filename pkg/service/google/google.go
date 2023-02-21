package google

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

const (
	state                        = "state-token"
	getGoogleUserInfoAPIEndpoint = "https://www.googleapis.com/plus/v1/people/me"
)

type Google struct {
	config *oauth2.Config
	gcs    *CloudStorage
	token  *oauth2.Token
}

// New function return Google service
func New(config *oauth2.Config, BucketName string, GCSProjectID string, GCSCredentials string) (*Google, error) {

	decoded, err := base64.StdEncoding.DecodeString(GCSCredentials)
	if err != nil {
		return nil, fmt.Errorf("failed to decode gcs credentials: %v", err)
	}

	client, err := storage.NewClient(context.Background(), option.WithCredentialsJSON(decoded))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	return &Google{
		config: config,
		gcs: &CloudStorage{
			client:     client,
			projectID:  GCSProjectID,
			bucketName: BucketName,
		},
	}, nil
}

// GetLoginURL return url for user loggin to google account
func (g *Google) GetLoginURL() string {
	authURL := g.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return authURL
}

// GetAccessToken return google access token
func (g *Google) GetAccessToken(code string, redirectURL string) (string, error) {
	g.config.RedirectURL = redirectURL
	token, err := g.config.Exchange(context.Background(), code)
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

// GetGoogleEmail return google user info
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
