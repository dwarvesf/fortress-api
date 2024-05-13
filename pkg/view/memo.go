package view

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/shopspring/decimal"
)

type MemoLog struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	URL         string          `json:"url"`
	Authors     []MemoLogAuthor `json:"authors"`
	Description string          `json:"description"`
	PublishedAt *time.Time      `json:"publishedAt"`
	Reward      decimal.Decimal `json:"reward"`
} // @name MemoLog

// MemoLogAuthor is the author of the memo log
type MemoLogAuthor struct {
	EmployeeID string `json:"employeeID"`
	GithubID   string `json:"githubID"`
	DiscordID  string `json:"discordID"`
}

// MemoLogsResponse response for memo logs
type MemoLogsResponse struct {
	Data []MemoLog `json:"data"`
} // @name MemoLogsResponse

func ToMemoLog(memoLogs []model.MemoLog) []MemoLog {
	rs := make([]MemoLog, 0)
	for _, memoLog := range memoLogs {
		authors := make([]MemoLogAuthor, 0)
		for _, author := range memoLog.Authors {
			var employeeID string
			if author.EmployeeID != nil {
				employeeID = author.EmployeeID.String()
			}

			authors = append(authors, MemoLogAuthor{
				EmployeeID: employeeID,
				GithubID:   author.GithubUsername,
				DiscordID:  author.DiscordID,
			})
		}

		// TODO: change response following the new model
		rs = append(rs, MemoLog{
			ID:          memoLog.ID.String(),
			Title:       memoLog.Title,
			URL:         memoLog.URL,
			Authors:     authors,
			Description: memoLog.Description,
			PublishedAt: memoLog.PublishedAt,
			Reward:      memoLog.Reward,
		})
	}

	return rs
}
