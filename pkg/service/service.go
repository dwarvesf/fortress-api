package service

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
)

type Service struct {
}

func New(cfg *config.Config) *Service {
	return &Service{}
}
