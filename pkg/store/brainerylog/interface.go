package brainerylog

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, b []model.BraineryLog) ([]model.BraineryLog, error)
	GetLimitByTimeRange(db *gorm.DB, start, end *time.Time, limit int) ([]*model.BraineryLog, error)
	GetNewContributorDiscordIDs(db *gorm.DB, start, end *time.Time) ([]string, error)
}
