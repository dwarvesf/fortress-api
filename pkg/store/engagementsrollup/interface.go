package engagementsrollup

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type IStore interface {
	Upsert(db *gorm.DB, record *model.EngagementsRollup) (*model.EngagementsRollup, error)
}
