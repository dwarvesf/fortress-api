package request

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/dwarvesf/fortress-api/pkg/handler/brainerylogs/errs"
)

type CreateBraineryLogRequest struct {
	Title       string          `json:"title" binding:"required"`
	URL         string          `json:"url" binding:"required"`
	GithubID    string          `json:"githubID"`
	DiscordID   string          `json:"discordID" binding:"required"`
	Tags        []string        `json:"tags" binding:"required"`
	PublishedAt string          `json:"publishedAt" binding:"required"`
	Reward      decimal.Decimal `json:"reward" binding:"required"`
}

func (r CreateBraineryLogRequest) Validate() error {
	if _, err := time.Parse(time.RFC3339Nano, r.PublishedAt); err != nil {
		return errs.ErrInvalidPublishedAt
	}
	return nil
}

type SyncBraineryLogs struct {
	StartMessageID string `json:"startMessageID"`
	EndMessageID   string `json:"endMessageID"`
}
