package service

import (
	"time"

	"github.com/patrickmn/go-cache"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/gmail/v1"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/service/currency"
	"github.com/dwarvesf/fortress-api/pkg/service/discord"
	"github.com/dwarvesf/fortress-api/pkg/service/github"
	googleauth "github.com/dwarvesf/fortress-api/pkg/service/google"
	"github.com/dwarvesf/fortress-api/pkg/service/googleadmin"
	"github.com/dwarvesf/fortress-api/pkg/service/googledrive"
	"github.com/dwarvesf/fortress-api/pkg/service/googlemail"
	"github.com/dwarvesf/fortress-api/pkg/service/improvmx"
	"github.com/dwarvesf/fortress-api/pkg/service/mochi"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/service/sendgrid"
	"github.com/dwarvesf/fortress-api/pkg/service/wise"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type Service struct {
	Basecamp    *basecamp.Service
	Cache       *cache.Cache
	Currency    currency.IService
	Discord     discord.IService
	Github      github.IService
	Google      googleauth.IService
	GoogleDrive googledrive.IService
	GoogleMail  googlemail.IService
	GoogleAdmin googleadmin.IService
	ImprovMX    improvmx.IService
	Mochi       mochi.IService
	Notion      notion.IService
	Sendgrid    sendgrid.IService
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

	googleAdminConfig := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes: []string{admin.AdminDirectoryUserScope,
			admin.AdminDirectoryGroupScope,
		},
	}
	googleAdminSvc := googleadmin.New(googleAdminConfig, cfg)

	mailConfig := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{gmail.MailGoogleComScope},
	}

	googleMailSvc := googlemail.New(
		mailConfig,
		cfg,
	)

	bc := model.Basecamp{
		ClientID:     cfg.Basecamp.ClientID,
		ClientSecret: cfg.Basecamp.ClientSecret,
	}
	Currency := currency.New(cfg)

	return &Service{
		Basecamp:    basecamp.New(store, repo, cfg, &bc, logger.L),
		Cache:       cch,
		Currency:    Currency,
		Discord:     discord.New(cfg),
		Github:      github.New(cfg, logger.L),
		Google:      googleSvc,
		GoogleAdmin: googleAdminSvc,
		GoogleDrive: googleDriveSvc,
		GoogleMail:  googleMailSvc,
		ImprovMX:    improvmx.New(cfg.ImprovMX.Token),
		Mochi:       mochi.New(cfg, logger.L),
		Notion:      notion.New(cfg.Notion.Secret, cfg.Notion.Databases.Project, logger.L),
		Sendgrid:    sendgrid.New(cfg.Sendgrid.APIKey, cfg, logger.L),
		Wise:        wise.New(cfg, logger.L),
	}
}
