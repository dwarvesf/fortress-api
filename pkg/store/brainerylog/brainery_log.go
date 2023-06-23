package brainerylog

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Create creates a brainery log record in the database
func (s *store) Create(db *gorm.DB, b []model.BraineryLog) (braineryLog []model.BraineryLog, err error) {
	return b, db.Create(b).Error
}

// GetLimitByTimeRange gets brainery logs in a specific time range, with limit
func (s *store) GetLimitByTimeRange(db *gorm.DB, start, end *time.Time, limit int) ([]*model.BraineryLog, error) {
	var logs []*model.BraineryLog
	return logs, db.Where("published_at BETWEEN ? AND ?", start, end).Limit(limit).Order("published_at DESC").Find(&logs).Error
}

// GetNewContributorDiscordIDs gets list of discord IDs from new contributors in a specific time range
func (s *store) GetNewContributorDiscordIDs(db *gorm.DB, start, end *time.Time) ([]string, error) {
	var discordIDs []string
	subQuery := db.Select("DISTINCT(discord_id)").Where("published_at <= ?", start).Table("brainery_logs")
	return discordIDs, db.Table("brainery_logs").
		Where("published_at BETWEEN ? AND ? AND discord_id NOT IN (?)", start, end, subQuery).
		Distinct().
		Pluck("discord_id", &discordIDs).Error
}
