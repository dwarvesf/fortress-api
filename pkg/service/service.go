package service

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/discord"
	"github.com/dwarvesf/fortress-api/pkg/service/google"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/service/wise"
	"github.com/patrickmn/go-cache"
)

type Service struct {
	Google  google.GoogleService
	Notion  notion.NotionService
	Wise    wise.IWiseService
	Cache   *cache.Cache
	Discord discord.DiscordService
}

func New(cfg *config.Config) *Service {
	googleSvc, err := google.New(
		cfg.Google.ClientID,
		cfg.Google.ClientSecret,
		cfg.Google.AppName,
		[]string{"email", "profile"},
		cfg.Google.GCSBucketName,
		cfg.Google.GCSProjectID,
		cfg.Google.GCSCredentials,
	)
	if err != nil {
		logger.L.Error(err, "failed to init google service")
	}
	cch := cache.New(5*time.Minute, 10*time.Minute)

	return &Service{
		Google: googleSvc,
		Notion: notion.New(
			cfg.Notion.Secret,
		),
		Wise:    wise.New(cfg, logger.L),
		Cache:   cch,
		Discord: discord.New(cfg.Discord.Webhooks.Campfire),
	}
}
