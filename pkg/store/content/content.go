package content

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
}

func New() IStore {
	return &store{}
}

// Create a content record
func (s *store) Create(tx *gorm.DB, content model.Content) (*model.Content, error) {
	return &content, tx.Create(&content).Error
}

// GetByName get content by name
func (s *store) GetByPath(tx *gorm.DB, path string) (*model.Content, error) {
	content := model.Content{}
	return &content, tx.
		Where("LOWER(path) = LOWER(?)", path).First(&content).Error
}
