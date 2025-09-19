package service

import (
	"time"

	"github.com/patrickmn/go-cache"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/sheets/v4"
	"google.golang.org/api/youtube/v3"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/service/communitynft"
	"github.com/dwarvesf/fortress-api/pkg/service/currency"
	"github.com/dwarvesf/fortress-api/pkg/service/discord"
	"github.com/dwarvesf/fortress-api/pkg/service/duckdb"
	"github.com/dwarvesf/fortress-api/pkg/service/evm"
	"github.com/dwarvesf/fortress-api/pkg/service/github"
	googleauth "github.com/dwarvesf/fortress-api/pkg/service/google"
	"github.com/dwarvesf/fortress-api/pkg/service/googleadmin"
	"github.com/dwarvesf/fortress-api/pkg/service/googledrive"
	"github.com/dwarvesf/fortress-api/pkg/service/googlemail"
	"github.com/dwarvesf/fortress-api/pkg/service/googlesheet"
	"github.com/dwarvesf/fortress-api/pkg/service/googlestorage"
	"github.com/dwarvesf/fortress-api/pkg/service/icyswap"
	"github.com/dwarvesf/fortress-api/pkg/service/improvmx"
	"github.com/dwarvesf/fortress-api/pkg/service/landingzone"
	"github.com/dwarvesf/fortress-api/pkg/service/lobsters"
	"github.com/dwarvesf/fortress-api/pkg/service/mochi"
	"github.com/dwarvesf/fortress-api/pkg/service/mochipay"
	"github.com/dwarvesf/fortress-api/pkg/service/mochiprofile"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/service/ogifmemosummarizer"
	"github.com/dwarvesf/fortress-api/pkg/service/parquet"
	"github.com/dwarvesf/fortress-api/pkg/service/reddit"
	"github.com/dwarvesf/fortress-api/pkg/service/sendgrid"
	"github.com/dwarvesf/fortress-api/pkg/service/tono"
	"github.com/dwarvesf/fortress-api/pkg/service/wise"
	yt "github.com/dwarvesf/fortress-api/pkg/service/youtube"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type Service struct {
	Basecamp      *basecamp.Service
	Cache         *cache.Cache
	Currency      currency.IService
	Discord       discord.IService
	DuckDB        duckdb.IService
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
	ParquetSync   parquet.ISyncService
	Sendgrid      sendgrid.IService
	Wise          wise.IService
	BaseClient    evm.IService
	IcySwap       icyswap.IService
	CommunityNft  communitynft.IService
	Tono          tono.IService
	Reddit        reddit.IService
	Lobsters      lobsters.IService
	Youtube       yt.IService
	Dify          ogifmemosummarizer.IService
	LandingZone   landingzone.IService
}

func New(cfg *config.Config, store *store.Store, repo store.DBRepo) (*Service, error) {
	cch := cache.New(5*time.Minute, 10*time.Minute)

	googleAuthSvc, err := googleauth.New(
		&oauth2.Config{
			ClientID:     cfg.Google.ClientID,
			ClientSecret: cfg.Google.ClientSecret,
			Endpoint:     google.Endpoint,
			Scopes:       []string{"email", "profile"},
		},
	)
	if err != nil {
		return nil, err
	}

	gcsSvc, err := googlestorage.New(
		cfg.Google.GCSBucketName,
		cfg.Google.GCSProjectID,
		cfg.Google.GCSCredentials,
	)
	if err != nil {
		return nil, err
	}

	landingZoneSvc, err := landingzone.New(
		cfg.Google.GCSLandingZoneCredentials,
	)
	if err != nil {
		return nil, err
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

	youtubeSvc := yt.New(&oauth2.Config{
		ClientID:     cfg.Youtube.ClientID,
		ClientSecret: cfg.Youtube.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{youtube.YoutubeForceSslScope},
	}, cfg)

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

	baseClient, err := evm.New(evm.DefaultBASEClient, cfg, logger.L)
	if err != nil {
		return nil, err
	}
	icySwap, err := icyswap.New(baseClient, cfg, logger.L)
	if err != nil {
		return nil, err
	}

	communityNft, err := communitynft.New(baseClient, cfg, logger.L)
	if err != nil {
		return nil, err
	}

	reddit, err := reddit.New(cfg, logger.L)
	if err != nil {
		return nil, err
	}

	difySvc := ogifmemosummarizer.New(cfg)

	duckDBSvc, err := duckdb.New(logger.L)
	if err != nil {
		return nil, err
	}

	parquetSvc := parquet.NewSyncService(cfg.Parquet, logger.L)

	return &Service{
		Basecamp:      basecamp.New(store, repo, cfg, &bc, logger.L),
		Cache:         cch,
		Currency:      Currency,
		Discord:       discord.New(cfg),
		DuckDB:        duckDBSvc,
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
		Notion:        notion.New(cfg.Notion.Secret, cfg.Notion.Databases.Project, logger.L, repo.DB()),
		ParquetSync:   parquetSvc,
		Sendgrid:      sendgrid.New(cfg.Sendgrid.APIKey, cfg, logger.L),
		Wise:          wise.New(cfg, logger.L),
		BaseClient:    baseClient,
		IcySwap:       icySwap,
		CommunityNft:  communityNft,
		Tono:          tono.New(cfg, logger.L),
		Reddit:        reddit,
		Lobsters:      lobsters.New(),
		Youtube:       youtubeSvc,
		Dify:          difySvc,
		LandingZone:   landingZoneSvc,
	}, nil
}

// NewForTest returns a Service with all external services set to nil or mocked for testing purposes.
func NewForTest() *Service {
	return &Service{
		Cache: cache.New(5*time.Minute, 10*time.Minute),
	}
}
