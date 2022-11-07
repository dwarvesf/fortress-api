package service

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/service/google"
)

type Service struct {
	Google google.GoogleService
}

func New(cfg *config.Config) *Service {
	return &Service{
		Google: google.New(
			cfg.Google.ClientID,
			cfg.Google.ClientSecret,
			cfg.Google.AppName,
			[]string{"email", "profile"},
		),
	}
}
