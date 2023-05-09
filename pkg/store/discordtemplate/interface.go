package discordtemplate

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	GetTemplateByType(db *gorm.DB, templateType string) (*model.DiscordLogTemplate, error)
}
