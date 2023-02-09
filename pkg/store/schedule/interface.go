package schedule

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, schedule *model.Schedule) (*model.Schedule, error)
	CreateDiscord(db *gorm.DB, schedule *model.ScheduleDiscordEvent) (*model.ScheduleDiscordEvent, error)
	GetOneByGcalID(db *gorm.DB, gcalID string) (*model.Schedule, error)

	Update(db *gorm.DB, schedule *model.Schedule) (*model.Schedule, error)
}
