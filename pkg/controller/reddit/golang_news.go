package reddit

import (
	"context"
)

const (
	// defaultGolangChannelID is the default Discord channel ID for Golang news.
	defaultGolangChannelID = "1257546039997501613"
)

// SyncGolangNews fetches new Golang news from Reddit and sends it to Discord.
func (c *controller) SyncGolangNews(ctx context.Context) error {
	logger := c.logger.Field("func", "GolangNews")

	popular, emerging, err := c.service.Reddit.FetchGolangNews(ctx)
	if err != nil {
		logger.Error(err, "failed to fetch Golang news")
		return err
	}

	if len(popular) == 0 && len(emerging) == 0 {
		logger.Info("no new Golang news")
		return nil
	}

	logger.Infof("new Golang news: %d popular, %d emerging", len(popular), len(emerging))

	golangChannelID := c.config.Discord.IDs.GolangChannel
	if golangChannelID == "" {
		golangChannelID = defaultGolangChannelID
	}

	if err := c.service.Discord.SendGolangNewsMessage(golangChannelID, emerging, popular); err != nil {
		logger.Error(err, "failed to send Golang news message to discord")
		return err
	}

	return nil
}
