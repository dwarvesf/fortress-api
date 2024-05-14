package memolog

import (
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

// List gets all memo logs
func (s *store) List(db *gorm.DB) ([]model.MemoLog, error) {
	var logs []model.MemoLog
	return logs, db.Preload("Authors").Preload("Authors.Employee").Order("published_at DESC").Find(&logs).Error
}
