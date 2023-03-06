package google

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

const (
	state                        = "state-token"
	getGoogleUserInfoAPIEndpoint = "https://www.googleapis.com/plus/v1/people/me"
)

type googleService struct {
	config *oauth2.Config
	gcs    *CloudStorage
	token  *oauth2.Token
}

// New function return Google service
func New(config *oauth2.Config, BucketName string, GCSProjectID string, GCSCredentials string) (IService, error) {
	decoded, err := base64.StdEncoding.DecodeString(GCSCredentials)
	if err != nil {
		return nil, fmt.Errorf("failed to decode gcs credentials: %v", err)
	}

	client, err := storage.NewClient(context.Background(), option.WithCredentialsJSON(decoded))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	return &googleService{
		config: config,
		gcs: &CloudStorage{
			client:     client,
			projectID:  GCSProjectID,
			bucketName: BucketName,
		},
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

// GetGoogleEmail return google user info
func (g *googleService) GetGoogleEmail(accessToken string) (email string, err error) {
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

	body, err := io.ReadAll(response.Body)
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

func (g *googleService) UploadContentGCS(file io.Reader, filePath string) error {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer.
	wc := g.gcs.client.Bucket(g.gcs.bucketName).Object(filePath).NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
}
