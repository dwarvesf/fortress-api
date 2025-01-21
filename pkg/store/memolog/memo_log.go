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
	return logs, db.Where("published_at BETWEEN ? AND ?", start, end).Limit(limit).Order("published_at DESC").Find(&logs).Error
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
	query := db.Order("published_at DESC")
	if filter.From != nil {
		query = query.Where("published_at >= ?", *filter.From)
	}
	if filter.To != nil {
		query = query.Where("published_at <= ?", *filter.To)
	}

	if filter.DiscordID != "" {
		query = query.Where("? = ANY(discord_account_ids)", filter.DiscordID)
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
		WHERE 
			discord_account_ids IS NULL OR 
			jsonb_array_length(discord_account_ids) = 0
	`

	return logs, db.Raw(query).Scan(&logs).Error
}

func (s *store) GetRankByDiscordID(db *gorm.DB, discordID string) (*model.DiscordAccountMemoRank, error) {
	query := `
		WITH discord_account AS (
			SELECT id 
			FROM public.discord_accounts 
			WHERE discord_id = ?
		),
		memo_count AS (
			SELECT
				da.discord_id,
				COUNT(DISTINCT ml.id) AS total_memos
			FROM
				public.memo_logs ml,
				discord_account da_id,
				public.discord_accounts da,
				jsonb_array_elements_text(ml.discord_account_ids) AS account_id
			WHERE
				ml.deleted_at IS NULL AND
				da.id = da_id.id AND
				da.id::text = account_id
			GROUP BY
				da.discord_id
		),
		ranked_memos AS (
			SELECT
				discord_id,
				total_memos,
				DENSE_RANK() OVER (ORDER BY total_memos DESC) AS rank
			FROM
				memo_count
		)
		SELECT
			rm.discord_id,
			rm.total_memos,
			rm.rank
		FROM
			ranked_memos rm
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
	return fmt.Errorf("memo_authors table no longer exists")
}

// GetTopAuthors gets the top authors by memo count
func (s *store) GetTopAuthors(db *gorm.DB, limit int) ([]model.DiscordAccountMemoRank, error) {
	query := `
		WITH memo_count AS (
			SELECT
				da.discord_id,
				da.discord_username,
				da.memo_username,
				COUNT(DISTINCT ml.id) AS total_memos
			FROM
				public.memo_logs ml,
				public.discord_accounts da,
				jsonb_array_elements_text(ml.discord_account_ids) AS account_id
			WHERE
				ml.deleted_at IS NULL AND
				da.id::text = account_id
			GROUP BY
				da.discord_id, 
				da.discord_username,
				da.memo_username
		)
		SELECT
			discord_id,
			discord_username,
			memo_username,
			total_memos,
			DENSE_RANK() OVER (ORDER BY total_memos DESC) AS rank
		FROM
			memo_count
		ORDER BY
			total_memos DESC
		LIMIT ?;
	`
	var topAuthors []model.DiscordAccountMemoRank
	return topAuthors, db.Raw(query, limit).Scan(&topAuthors).Error
}
