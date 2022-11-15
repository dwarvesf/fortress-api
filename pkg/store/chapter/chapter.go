package chapter

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
}

func New() IStore {
	return &store{}
}

// All get all chapters
func (s *store) All(db *gorm.DB) ([]*model.Chapter, error) {
	var chapters []*model.Chapter
	return chapters, db.Find(&chapters).Error
}
