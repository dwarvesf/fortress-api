package service

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/service/google"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
)

type Service struct {
	Google google.GoogleService
	Notion notion.NotionService
}

func New(cfg *config.Config) *Service {
	return &Service{
		Google: google.New(
			cfg.Google.ClientID,
			cfg.Google.ClientSecret,
			cfg.Google.AppName,
			[]string{"email", "profile"},
			cfg.Google.GCSBucketName,
			cfg.Google.GCSProjectID,
		),
		Notion: notion.New(
			cfg.Notion.Secret,
		),
	}
}
