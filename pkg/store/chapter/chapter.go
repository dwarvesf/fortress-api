package chapter

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// All get all chapters
func (s *store) All(db *gorm.DB) ([]*model.Chapter, error) {
	var chapters []*model.Chapter
	return chapters, db.Find(&chapters).Error
}

// IsExist check existence of a chapter
func (s *store) IsExist(db *gorm.DB, id string) (bool, error) {
	type res struct {
		Result bool
	}

	result := res{}
	query := db.Raw("SELECT EXISTS (SELECT * FROM chapters WHERE id = ?) as result", id)

	return result.Result, query.Scan(&result).Error
}

func (s *store) UpdateChapterLead(db *gorm.DB, id string, lead string) error {
	return db.Model(&model.Chapter{}).Where("id = ?", id).Update("lead_id", lead).Error
}

// GetAllByLeadID get all chapters by lead_id
func (s *store) GetAllByLeadID(db *gorm.DB, leadID string) ([]*model.Chapter, error) {
	var chapters []*model.Chapter
	return chapters, db.Where("lead_id = ?", leadID).Find(&chapters).Error
}
