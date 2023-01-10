package service

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/google"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/service/wise"
)

type Service struct {
	Google google.GoogleService
	Notion notion.NotionService
	Wise   wise.IWiseService
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

	return &Service{
		Google: googleSvc,
		Notion: notion.New(
			cfg.Notion.Secret,
		),
		Wise: wise.New(cfg, logger.L),
	}
}
