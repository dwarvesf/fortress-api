package discordtemplate

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
}

func New() IStore {
	return &store{}
}

func (s *store) GetTemplateByType(db *gorm.DB, templateType string) (*model.DiscordLogTemplate, error) {

	var template *model.DiscordLogTemplate
	return template, db.Where("type = ?", templateType).First(&template).Error
}
