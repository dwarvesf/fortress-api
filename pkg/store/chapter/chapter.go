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

// Exist check existence of a chapter
func (s *store) Exists(db *gorm.DB, id string) (bool, error) {
	type res struct {
		Result bool
	}

	result := res{}
	query := db.Raw("SELECT EXISTS (SELECT * FROM chapters WHERE id = ?) as result", id)

	return result.Result, query.Scan(&result).Error
}
