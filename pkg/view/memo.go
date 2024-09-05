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
	Category    []string        `json:"category"`
} // @name MemoLog

// MemoLogAuthor is the author of the memo log
type MemoLogAuthor struct {
	EmployeeID      string `json:"employeeID"`
	GithubUsername  string `json:"githubUsername"`
	DiscordID       string `json:"discordID"`
	DiscordUsername string `json:"discordUsername"`
	PersonalEmail   string `json:"personalEmail"`
	MemoUsername    string `json:"memoUsername"`
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
			if author.Employee != nil {
				employeeID = author.Employee.ID.String()
			}

			authors = append(authors, MemoLogAuthor{
				EmployeeID:      employeeID,
				GithubUsername:  author.GithubUsername,
				DiscordID:       author.DiscordID,
				PersonalEmail:   author.PersonalEmail,
				DiscordUsername: author.DiscordUsername,
				MemoUsername:    author.MemoUsername,
			})
		}

		rs = append(rs, MemoLog{
			ID:          memoLog.ID.String(),
			Title:       memoLog.Title,
			URL:         memoLog.URL,
			Authors:     authors,
			Description: memoLog.Description,
			PublishedAt: memoLog.PublishedAt,
			Reward:      memoLog.Reward,
			Category:    memoLog.Category,
		})
	}

	return rs
}

// MemoLogByDiscordIDResponse response for memo logs
type MemoLogByDiscordIDResponse struct {
	Data MemoLogsByDiscordID `json:"data"`
} // @name MemoLogByDiscordIDResponse

type MemoLogsByDiscordID struct {
	MemoLogs []MemoLog     `json:"memoLogs"`
	Rank     AuthorRanking `json:"rank"`
} // @name MemoLogsByDiscordID

// ToMemoLogByDiscordID ...
func ToMemoLogByDiscordID(memoLogs []model.MemoLog, discordMemoRank *model.DiscordAccountMemoRank) MemoLogsByDiscordID {
	rs := make([]MemoLog, 0)
	for _, memoLog := range memoLogs {
		authors := make([]MemoLogAuthor, 0)
		for _, author := range memoLog.Authors {
			var employeeID string
			if author.Employee != nil {
				employeeID = author.Employee.ID.String()
			}

			rank := &AuthorRanking{}
			if discordMemoRank != nil {
				rank.DiscordID = discordMemoRank.DiscordID
				rank.TotalMemos = discordMemoRank.TotalMemos
				rank.Rank = discordMemoRank.Rank
			}

			authors = append(authors, MemoLogAuthor{
				EmployeeID:      employeeID,
				GithubUsername:  author.GithubUsername,
				DiscordID:       author.DiscordID,
				PersonalEmail:   author.PersonalEmail,
				DiscordUsername: author.DiscordUsername,
				MemoUsername:    author.MemoUsername,
			})
		}

		rs = append(rs, MemoLog{
			ID:          memoLog.ID.String(),
			Title:       memoLog.Title,
			URL:         memoLog.URL,
			Authors:     authors,
			Description: memoLog.Description,
			PublishedAt: memoLog.PublishedAt,
			Reward:      memoLog.Reward,
			Category:    memoLog.Category,
		})
	}

	authorRank := AuthorRanking{}
	if discordMemoRank != nil {
		authorRank = AuthorRanking{
			DiscordID:  discordMemoRank.DiscordID,
			TotalMemos: discordMemoRank.TotalMemos,
			Rank:       discordMemoRank.Rank,
		}
	}

	return MemoLogsByDiscordID{
		MemoLogs: rs,
		Rank:     authorRank,
	}
}

type AuthorRanking struct {
	DiscordID       string `json:"discordID"`
	DiscordUsername string `json:"discordUsername"`
	MemoUsername    string `json:"memoUsername"`
	TotalMemos      int    `json:"totalMemos"`
	Rank            int    `json:"rank"`
} // @name AuthorRanking

// MemoTopAuthorsResponse response for memo top authors
type MemoTopAuthorsResponse struct {
	Data []AuthorRanking `json:"data"`
} // @name MemoTopAuthorsResponse

// ToMemoTopAuthors ...
func ToMemoTopAuthors(discordMemoRank []model.DiscordAccountMemoRank) []AuthorRanking {
	rs := make([]AuthorRanking, 0)
	for _, rank := range discordMemoRank {
		rs = append(rs, AuthorRanking{
			DiscordID:       rank.DiscordID,
			DiscordUsername: rank.DiscordUsername,
			MemoUsername:    rank.MemoUsername,
			TotalMemos:      rank.TotalMemos,
			Rank:            rank.Rank,
		})
	}

	return rs
}
