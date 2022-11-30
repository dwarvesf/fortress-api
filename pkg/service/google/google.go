package google

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

const (
	state                        = "state-token"
	getGoogleUserInfoAPIEndpoint = "https://www.googleapis.com/plus/v1/people/me"
)

type ClientUploader struct {
	cl         *storage.Client
	projectID  string
	bucketName string
}
type Google struct {
	Config   *oauth2.Config
	Uploader *ClientUploader
}

// New function return Google service
func New(ClientID, ClientSecret, AppName string, Scopes []string, BucketName string, GCSProjectID string, GCSCredentials string) (*Google, error) {
	Config := &oauth2.Config{
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       Scopes,
	}

	decoded, err := base64.StdEncoding.DecodeString(GCSCredentials)
	if err != nil {
		return nil, fmt.Errorf("failed to decode gcs credentials: %w", err)
	}

	client, err := storage.NewClient(context.Background(), option.WithCredentialsJSON(decoded))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &Google{
		Config: Config,
		Uploader: &ClientUploader{
			cl:         client,
			projectID:  GCSProjectID,
			bucketName: BucketName,
		},
	}, nil
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
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
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

func (g *Google) UploadContentGCS(file multipart.File, filePath string) error {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer.
	wc := g.Uploader.cl.Bucket(g.Uploader.bucketName).Object(filePath).NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %w", err)
	}

	return nil
}
