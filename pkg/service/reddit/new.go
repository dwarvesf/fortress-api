package reddit

import (
	"fmt"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

type service struct {
	client *reddit.Client
}

func New(cfg *config.Config, l logger.Logger) (IService, error) {
	clientID := cfg.Reddit.ClientID
	if clientID == "" {
		l.Warn("reddit client id is empty")
	}

	clientSecret := cfg.Reddit.ClientSecret
	if clientSecret == "" {
		l.Warn("reddit client secret is empty")
	}

	username := cfg.Reddit.Username
	if username == "" {
		l.Warn("reddit username is empty")
	}

	password := cfg.Reddit.Password
	if password == "" {
		l.Warn("reddit password is empty")
	}

	auth := reddit.Credentials{
		ID:       clientID,
		Secret:   clientSecret,
		Username: username,
		Password: password,
	}

	client, err := reddit.NewClient(auth, reddit.WithUserAgent("fortress-bot"))
	if err != nil {
		return nil, fmt.Errorf("create reddit client failed: %w", err)
	}

	return &service{
		client: client,
	}, nil
}
