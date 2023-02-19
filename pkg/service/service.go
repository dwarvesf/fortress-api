package service

import (
	"time"

	"github.com/patrickmn/go-cache"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/discord"
	googleauth "github.com/dwarvesf/fortress-api/pkg/service/google"
	"github.com/dwarvesf/fortress-api/pkg/service/googledrive"
	"github.com/dwarvesf/fortress-api/pkg/service/googlemail"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/service/wise"
)

type Service struct {
	Google      googleauth.Service
	GoogleDrive googledrive.Service
	GoogleMail  googlemail.Service
	Notion      notion.NotionService
	Wise        wise.IWiseService
	Cache       *cache.Cache
	Discord     discord.DiscordService
}

func New(cfg *config.Config) *Service {
	cch := cache.New(5*time.Minute, 10*time.Minute)

	authServiceCfg := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"email", "profile"},
	}

	googleSvc, err := googleauth.New(
		authServiceCfg,
		cfg.Google.GCSBucketName,
		cfg.Google.GCSProjectID,
		cfg.Google.GCSCredentials,
	)
	if err != nil {
		logger.L.Error(err, "failed to init google service")
	}

	driveConfig := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{googledrive.FullDriveAccessScope},
	}

	googleDriveSvc := googledrive.New(driveConfig, cfg)

	mailConfig := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{gmail.MailGoogleComScope},
	}

	googleMailService := googlemail.New(
		cfg.Google.APIKey,
		mailConfig,
		cfg,
	)

	return &Service{
		Google:      googleSvc,
		GoogleDrive: googleDriveSvc,
		GoogleMail:  googleMailService,
		Notion: notion.New(
			cfg.Notion.Secret,
		),
		Wise:    wise.New(cfg, logger.L),
		Cache:   cch,
		Discord: discord.New(cfg),
	}
}
