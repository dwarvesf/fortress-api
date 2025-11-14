package accountingtaskref

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) Create(db *gorm.DB, ref *model.AccountingTaskRef) error {
	return db.Create(ref).Error
}

func (s *store) FindByProjectMonthYear(db *gorm.DB, projectID string, month, year int, group string) ([]*model.AccountingTaskRef, error) {
	if db == nil {
		return nil, gorm.ErrInvalidDB
	}
	query := db.Where("month = ? AND year = ?", month, year)
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	if group != "" {
		query = query.Where("group_name = ?", group)
	}
	var refs []*model.AccountingTaskRef
	if err := query.Order("created_at DESC").Find(&refs).Error; err != nil {
		return nil, err
	}
	return refs, nil
}
