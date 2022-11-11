package chapter

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
	db *gorm.DB
}

func New(db *gorm.DB) IStore {
	return &store{
		db: db,
	}
}

// All get all chapters
func (s *store) All() ([]*model.Chapter, error) {
	var chapters []*model.Chapter
	return chapters, s.db.Find(&chapters).Error
}
