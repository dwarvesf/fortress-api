package memolog

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, b []model.MemoLog) ([]model.MemoLog, error)
	GetLimitByTimeRange(db *gorm.DB, start, end *time.Time, limit int) ([]model.MemoLog, error)
	List(db *gorm.DB, filter ListFilter) ([]model.MemoLog, error)
	GetRankByDiscordID(db *gorm.DB, discordID string) (*model.DiscordAccountMemoRank, error)
	ListNonAuthor(db *gorm.DB) ([]model.MemoLog, error)
	CreateMemoAuthor(db *gorm.DB, memoAuthor *model.MemoAuthor) error
}
