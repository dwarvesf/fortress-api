package brainerylog

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, b *model.BraineryLog) (*model.BraineryLog, error)
	GetByTimeRange(db *gorm.DB, start, end *time.Time) ([]*model.BraineryLog, error)
	GetNewContributorDiscordIDs(db *gorm.DB, start, end *time.Time) ([]string, error)
}
