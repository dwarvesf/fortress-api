package service

import (
	"time"

	cache "github.com/patrickmn/go-cache"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gmail "google.golang.org/api/gmail/v1"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/service/currency"
	"github.com/dwarvesf/fortress-api/pkg/service/discord"
	googleauth "github.com/dwarvesf/fortress-api/pkg/service/google"
	"github.com/dwarvesf/fortress-api/pkg/service/googledrive"
	"github.com/dwarvesf/fortress-api/pkg/service/googlemail"
	"github.com/dwarvesf/fortress-api/pkg/service/improvmx"
	"github.com/dwarvesf/fortress-api/pkg/service/mochi"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/service/redis"
	"github.com/dwarvesf/fortress-api/pkg/service/sendgrid"
	"github.com/dwarvesf/fortress-api/pkg/service/wise"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type Service struct {
	Basecamp    *basecamp.Service
	Cache       *cache.Cache
	Currency    currency.IService
	Discord     discord.IService
	Google      googleauth.IService
	GoogleDrive googledrive.Service
	GoogleMail  googlemail.IService
	ImprovMX    improvmx.IService
	Mochi       mochi.IService
	Notion      notion.IService
	Redis       redis.RedisService
	Sendgrid    sendgrid.Service
	Wise        wise.IService
}

func New(cfg *config.Config, store *store.Store, repo store.DBRepo) *Service {
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
		mailConfig,
		cfg,
	)

	bc := model.Basecamp{
		ClientID:     cfg.Basecamp.ClientID,
		ClientSecret: cfg.Basecamp.ClientSecret,
	}
	Currency := currency.New(cfg)

	cch := cache.New(5*time.Minute, 10*time.Minute)
	redisSvc, err := redis.New(cfg)
	if err != nil {
		logger.L.Error(err, "failed to init redis service")
	}

	return &Service{
		Basecamp:    basecamp.NewService(store, repo, cfg, &bc, logger.L),
		Cache:       cch,
		Currency:    Currency,
		Discord:     discord.New(cfg),
		Google:      googleSvc,
		GoogleDrive: googleDriveSvc,
		GoogleMail:  googleMailService,
		ImprovMX:    improvmx.New(cfg.ImprovMX.Token),
		Mochi:       mochi.New(cfg, logger.L),
		Notion:      notion.New(cfg.Notion.Secret, cfg.Notion.Databases.Project, logger.L),
		Redis:       redisSvc,
		Sendgrid:    sendgrid.New(cfg.Sendgrid.APIKey, cfg, logger.L),
		Wise:        wise.New(cfg, logger.L),
	}
}
