package question

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// AllByCategory get all by category and subcategory
func (s *store) AllByCategory(db *gorm.DB, category model.EventType, subcategory model.EventSubtype) ([]*model.Question, error) {
	var questions []*model.Question
	return questions, db.Where("category = ? AND subcategory = ?", category, subcategory).Order("\"order\"").Find(&questions).Error
}
