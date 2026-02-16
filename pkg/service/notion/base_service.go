package notion

import (
	"errors"

	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

// baseService provides shared fields and constructor logic for Notion services.
type baseService struct {
	client *nt.Client
	cfg    *config.Config
	logger logger.Logger
}

// newBaseService creates a baseService with a configured Notion client.
// Returns nil if the Notion secret is not configured.
func newBaseService(cfg *config.Config, l logger.Logger) *baseService {
	if cfg.Notion.Secret == "" {
		l.Error(errors.New("notion secret not configured"), "notion secret is empty")
		return nil
	}
	return &baseService{
		client: nt.NewClient(cfg.Notion.Secret),
		cfg:    cfg,
		logger: l,
	}
}
