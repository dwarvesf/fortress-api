package service

import (
	"golang.org/x/oauth2"
	oauth2Google "golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp"
	"github.com/dwarvesf/fortress-api/pkg/service/google"
	"github.com/dwarvesf/fortress-api/pkg/service/googlemail"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/service/wise"
)

type Service struct {
	Google     google.GoogleService
	GoogleMail googlemail.GoogleMailService
	Notion     notion.NotionService
	Wise       wise.WiseService
	Bc3        basecamp.BasecampService
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
	bc3 := basecamp.NewTestService()
	if cfg.Env != "test" {
		bc3, err = basecamp.New(cfg.Basecamp)
		if err != nil {
			logger.L.Error(err, "failed to init basecamp service")
		}
	}

	return &Service{
		Google: googleSvc,
		GoogleMail: googlemail.New(
			cfg.Env,
			cfg.Google.MailApiKey,
			&oauth2.Config{
				ClientID:     cfg.Google.ClientID,
				ClientSecret: cfg.Google.ClientSecret,
				Endpoint:     oauth2Google.Endpoint,
				Scopes:       []string{gmail.MailGoogleComScope},
			},
			cfg.Google.TemplatePath,
			cfg.Google.TeamEmailToken,
			cfg.Google.TeamEmailID,
			cfg.Google.AccountingEmailToken,
			cfg.Google.AccountingEmailID,
		),
		Notion: notion.New(
			cfg.Notion.Secret,
		),
		Wise: wise.New(
			cfg.Wise.ApiKey,
			cfg.Wise.Profile,
			cfg.Env,
		),
		Bc3: *bc3,
	}
}
