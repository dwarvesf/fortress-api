package youtube

import (
	"context"
	"errors"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type youtubeService struct {
	config    *oauth2.Config
	token     *oauth2.Token
	service   *youtube.Service
	appConfig *config.Config
}

// New function return Google service
func New(config *oauth2.Config, appConfig *config.Config) IService {
	return &youtubeService{
		config:    config,
		appConfig: appConfig,
	}
}

func (yt *youtubeService) prepareService() error {
	client := yt.config.Client(context.Background(), yt.token)
	service, err := youtube.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return errors.New("Get Youtube Service Failed " + err.Error())
	}

	yt.service = service

	return nil
}

func (yt *youtubeService) ensureToken(refreshToken string) error {
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	if !yt.token.Valid() {
		tks := yt.config.TokenSource(context.Background(), token)
		tok, err := tks.Token()
		if err != nil {
			return err
		}

		yt.token = tok
	}

	return nil
}

// CreateBroadcast function create broadcast on youtube
func (yt *youtubeService) CreateBroadcast(e *model.Event) (err error) {
	if err := yt.ensureToken(yt.appConfig.Youtube.RefreshToken); err != nil {
		return err
	}

	if err := yt.prepareService(); err != nil {
		return err
	}

	// Load the Vietnam timezone location
	location, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		return err
	}
	// Get the current time in the Vietnam timezone
	t := time.Now().In(location)
	// check if the event is before 17h, then set the broadcast to 17h in Vietnam
	if t.Hour() < 17 {
		// set the broadcast to 17h in timezone vietnam
		t = time.Date(t.Year(), t.Month(), t.Day(), 17, 0, 0, 0, location)
	}

	// Insert broadcast
	_, err = yt.insertBroadcast(e.Name, e.Description, t)
	if err != nil {
		return err
	}

	return nil
}

func (yt *youtubeService) insertBroadcast(title string, description string, startTime time.Time) (*youtube.LiveBroadcast, error) {
	liveBroadcast := &youtube.LiveBroadcast{
		Snippet: &youtube.LiveBroadcastSnippet{
			Title:              title,
			Description:        description,
			ScheduledStartTime: startTime.Format(time.RFC3339),
		},
		Status: &youtube.LiveBroadcastStatus{
			PrivacyStatus: "unlisted",
		},
	}
	return yt.service.LiveBroadcasts.Insert([]string{"snippet", "status"}, liveBroadcast).Do()
}
