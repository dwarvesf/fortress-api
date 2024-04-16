package request

import (
	"github.com/shopspring/decimal"
)

type CreateMemoLogsRequest []MemoLogItem // @name CreateMemoLogsRequest

type MemoLogItem struct {
	Title       string          `json:"title" binding:"required"`
	URL         string          `json:"url" binding:"required"`
	Authors     []string        `json:"authors"`
	Tags        []string        `json:"tags"`
	Description string          `json:"description"`
	PublishedAt string          `json:"publishedAt"`
	Reward      decimal.Decimal `json:"reward"`
}
