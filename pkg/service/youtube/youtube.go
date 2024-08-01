package youtube

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
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
	return yt.insertBroadcast(e, t)
}

func (yt *youtubeService) insertBroadcast(e *model.Event, startTime time.Time) error {
	if err := yt.ensureToken(yt.appConfig.Youtube.RefreshToken); err != nil {
		return err
	}

	if err := yt.prepareService(); err != nil {
		return err
	}

	liveBroadcast := &youtube.LiveBroadcast{
		Snippet: &youtube.LiveBroadcastSnippet{
			Title:              e.Name,
			Description:        e.Description,
			ScheduledStartTime: startTime.Format(time.RFC3339),
		},
		Status: &youtube.LiveBroadcastStatus{
			PrivacyStatus: "unlisted",
		},
	}

	lbc, err := yt.service.LiveBroadcasts.Insert([]string{"snippet", "status"}, liveBroadcast).Do()
	if err != nil {
		return err
	}

	if e.Image == "" {
		return nil
	}

	// download by url and open the image file
	imgPath := fmt.Sprintf("/tmp/%v.png", e.Image)
	err = yt.downloadImage(fmt.Sprintf("https://cdn.discordapp.com/guild-events/%v/%v.png?size=4096", e.DiscordEventID, e.Image), imgPath)
	if err != nil {
		return err
	}

	// Upload a thumbnail
	file, err := os.Open(imgPath)
	if err != nil {
		return err
	}
	defer file.Close()

	thumbnailCall := yt.service.Thumbnails.Set(lbc.Id)
	_, err = thumbnailCall.Media(file).Do()
	if err != nil {
		return err
	}

	return nil
}

func (yt *youtubeService) downloadImage(url, filepath string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error downloading image: %w", err)
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("error saving image: %w", err)
	}

	return nil
}

// get the latest broadcast
func (yt *youtubeService) GetLatestBroadcast() (*youtube.LiveBroadcast, error) {
	if err := yt.ensureToken(yt.appConfig.Youtube.RefreshToken); err != nil {
		return nil, err
	}

	if err := yt.prepareService(); err != nil {
		return nil, err
	}

	broadcasts, err := yt.listBroadcasts("completed")
	if err != nil {
		return nil, err
	}

	if len(broadcasts) == 0 {
		return nil, nil
	}

	return broadcasts[0], nil
}

// listBroadcasts function list all broadcasts on youtube
// status: all, active, completed, upcoming
func (yt *youtubeService) listBroadcasts(status string) ([]*youtube.LiveBroadcast, error) {
	if err := yt.ensureToken(yt.appConfig.Youtube.RefreshToken); err != nil {
		return nil, err
	}

	if err := yt.prepareService(); err != nil {
		return nil, err
	}

	broadcasts, err := yt.service.LiveBroadcasts.List([]string{"id", "snippet", "status"}).BroadcastStatus(status).Do()
	if err != nil {
		return nil, err
	}

	return broadcasts.Items, nil
}
