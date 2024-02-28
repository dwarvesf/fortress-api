package service

import (
	"github.com/dwarvesf/fortress-api/pkg/service/googlestorage"
	"time"

	"github.com/patrickmn/go-cache"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/sheets/v4"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/service/currency"
	"github.com/dwarvesf/fortress-api/pkg/service/discord"
	"github.com/dwarvesf/fortress-api/pkg/service/evm"
	"github.com/dwarvesf/fortress-api/pkg/service/github"
	googleauth "github.com/dwarvesf/fortress-api/pkg/service/google"
	"github.com/dwarvesf/fortress-api/pkg/service/googleadmin"
	"github.com/dwarvesf/fortress-api/pkg/service/googledrive"
	"github.com/dwarvesf/fortress-api/pkg/service/googlemail"
	"github.com/dwarvesf/fortress-api/pkg/service/googlesheet"
	"github.com/dwarvesf/fortress-api/pkg/service/icyswap"
	"github.com/dwarvesf/fortress-api/pkg/service/improvmx"
	"github.com/dwarvesf/fortress-api/pkg/service/mochi"
	"github.com/dwarvesf/fortress-api/pkg/service/mochipay"
	"github.com/dwarvesf/fortress-api/pkg/service/mochiprofile"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/service/sendgrid"
	"github.com/dwarvesf/fortress-api/pkg/service/wise"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type Service struct {
	Basecamp      *basecamp.Service
	Cache         *cache.Cache
	Currency      currency.IService
	Discord       discord.IService
	Github        github.IService
	Google        googleauth.IService
	GoogleStorage googlestorage.IService
	GoogleAdmin   googleadmin.IService
	GoogleDrive   googledrive.IService
	GoogleMail    googlemail.IService
	GoogleSheet   googlesheet.IService
	ImprovMX      improvmx.IService
	Mochi         mochi.IService
	MochiPay      mochipay.IService
	MochiProfile  mochiprofile.IService
	Notion        notion.IService
	Sendgrid      sendgrid.IService
	Wise          wise.IService
	PolygonClient evm.IService
	IcySwap       icyswap.IService
}

func New(cfg *config.Config, store *store.Store, repo store.DBRepo) *Service {
	cch := cache.New(5*time.Minute, 10*time.Minute)

	authServiceCfg := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"email", "profile"},
	}

	gcsSvc, err := googlestorage.New(
		cfg.Google.GCSBucketName,
		cfg.Google.GCSProjectID,
		cfg.Google.GCSCredentials,
	)
	if err != nil {
		logger.L.Error(err, "failed to init gcs")
	}

	googleAuthSvc, err := googleauth.New(
		authServiceCfg,
	)
	if err != nil {
		logger.L.Error(err, "failed to init google auth")
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

	gSheetConfig := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes: []string{
			sheets.DriveScope,
			sheets.DriveFileScope,
			sheets.DriveReadonlyScope,
			sheets.SpreadsheetsScope,
			sheets.SpreadsheetsReadonlyScope,
		},
	}

	gSheetSvc := googlesheet.New(
		gSheetConfig,
		cfg,
	)

	bc := model.Basecamp{
		ClientID:     cfg.Basecamp.ClientID,
		ClientSecret: cfg.Basecamp.ClientSecret,
	}

	Currency := currency.New(cfg)

	polygonClient, err := evm.New(evm.DefaultPolygonClient, cfg, logger.L)
	if err != nil {
		logger.L.Error(err, "failed to init polygon client service")
	}
	icySwap, err := icyswap.New(polygonClient, cfg, logger.L)
	if err != nil {
		logger.L.Error(err, "failed to init icyswap service")
	}

	return &Service{
		Basecamp:      basecamp.New(store, repo, cfg, &bc, logger.L),
		Cache:         cch,
		Currency:      Currency,
		Discord:       discord.New(cfg),
		Github:        github.New(cfg, logger.L),
		Google:        googleAuthSvc,
		GoogleStorage: gcsSvc,
		GoogleAdmin:   googleAdminSvc,
		GoogleDrive:   googleDriveSvc,
		GoogleMail:    googleMailSvc,
		GoogleSheet:   gSheetSvc,
		ImprovMX:      improvmx.New(cfg.ImprovMX.Token),
		Mochi:         mochi.New(cfg, logger.L),
		MochiPay:      mochipay.New(cfg, logger.L),
		MochiProfile:  mochiprofile.New(cfg, logger.L),
		Notion:        notion.New(cfg.Notion.Secret, cfg.Notion.Databases.Project, logger.L),
		Sendgrid:      sendgrid.New(cfg.Sendgrid.APIKey, cfg, logger.L),
		Wise:          wise.New(cfg, logger.L),
		PolygonClient: polygonClient,
		IcySwap:       icySwap,
	}
}
