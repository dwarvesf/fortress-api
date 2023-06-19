package brainerylog

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, b *model.BraineryLog) (braineryLog *model.BraineryLog, err error)
}
