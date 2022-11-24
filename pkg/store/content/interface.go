package content

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type IStore interface {
	Create(tx *gorm.DB, content model.Content) (*model.Content, error)
	GetByPath(tx *gorm.DB, path string) (*model.Content, error)
}
