package memolog

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Create creates a memo log record in the database
func (s *store) Create(db *gorm.DB, b []model.MemoLog) ([]model.MemoLog, error) {
	return b, db.Table("memo_logs").Create(b).Error
}

// GetLimitByTimeRange gets memo logs in a specific time range, with limit
func (s *store) GetLimitByTimeRange(db *gorm.DB, start, end *time.Time, limit int) ([]model.MemoLog, error) {
	var logs []model.MemoLog
	return logs, db.Preload("Authors").Preload("Authors.Employee").Where("published_at BETWEEN ? AND ?", start, end).Limit(limit).Order("published_at DESC").Find(&logs).Error
}

// ListFilter is a filter for List function
type ListFilter struct {
	From      *time.Time
	To        *time.Time
	DiscordID string
}

// List gets all memo logs
func (s *store) List(db *gorm.DB, filter ListFilter) ([]model.MemoLog, error) {
	var logs []model.MemoLog
	query := db.Preload("Authors").Preload("Authors.Employee").Order("published_at DESC")
	if filter.From != nil {
		query = query.Where("published_at >= ?", *filter.From)
	}
	if filter.To != nil {
		query = query.Where("published_at <= ?", *filter.To)
	}

	if filter.DiscordID != "" {
		query = query.Joins("JOIN memo_authors ma ON ma.memo_log_id = memo_logs.id").
			Joins("JOIN discord_accounts da ON da.id = ma.discord_account_id AND da.discord_id = ?", filter.DiscordID)
	}

	return logs, query.Find(&logs).Error
}

// ListNonAuthor gets all memo logs that have no author info
func (s *store) ListNonAuthor(db *gorm.DB) ([]model.MemoLog, error) {
	var logs []model.MemoLog
	query := `
		SELECT 
			memo_logs.*
		FROM 
			memo_logs
		LEFT JOIN 
			memo_authors ON memo_authors.memo_log_id = memo_logs.id
		GROUP BY 
			memo_logs.id
		HAVING 
			STRING_AGG(memo_authors.discord_account_id::text, ', ') IS NULL OR 
			STRING_AGG(memo_authors.discord_account_id::text, ', ') = ''
	`

	return logs, db.Raw(query).Scan(&logs).Error
}

func (s *store) GetRankByDiscordID(db *gorm.DB, discordID string) (*model.DiscordAccountMemoRank, error) {
	query := `
		WITH memo_count AS (
			SELECT
				da.discord_id,
				COUNT(ml.id) AS total_memos
			FROM
				public.memo_authors ma
			JOIN
				public.memo_logs ml ON ma.memo_log_id = ml.id
			JOIN
				public.discord_accounts da ON ma.discord_account_id = da.id
			WHERE
				ml.deleted_at IS NULL
			GROUP BY
				da.discord_id
		),
		ranked_memos AS (
			SELECT
				discord_id,
				total_memos,
				RANK() OVER (ORDER BY total_memos DESC) AS rank
			FROM
				memo_count
		)
		SELECT
			rm.discord_id,
			rm.total_memos,
			rm.rank
		FROM
			ranked_memos rm
		WHERE
			rm.discord_id = ?
	`
	var memoRank model.DiscordAccountMemoRank
	result := db.Raw(query, discordID).Scan(&memoRank)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("no records found for discord_id: %s", discordID)
	}

	return &memoRank, nil
}

// CreateMemoAuthor creates a memo author record in the database
func (s *store) CreateMemoAuthor(db *gorm.DB, memoAuthor *model.MemoAuthor) error {
	return db.Create(memoAuthor).Error
}

// GetTopAuthors gets the top authors by memo count
func (s *store) GetTopAuthors(db *gorm.DB, limit int) ([]model.DiscordAccountMemoRank, error) {
	query := `
		WITH memo_count AS (
    SELECT
        da.discord_id,
        COUNT(ml.id) AS total_memos
    FROM
        public.memo_authors ma
    JOIN
        public.memo_logs ml ON ma.memo_log_id = ml.id
    JOIN
        public.discord_accounts da ON ma.discord_account_id = da.id
    WHERE
        ml.deleted_at IS NULL -- Exclude deleted memos if necessary
    GROUP BY
        da.discord_id, da.discord_username
		)
		SELECT
			discord_id,
			total_memos,
			RANK() OVER (ORDER BY total_memos DESC) AS rank
		FROM
			memo_count
		ORDER BY
			total_memos DESC
		LIMIT ?;
	`
	var topAuthors []model.DiscordAccountMemoRank
	return topAuthors, db.Raw(query, limit).Scan(&topAuthors).Error
}
