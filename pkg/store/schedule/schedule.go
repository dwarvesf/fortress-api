package schedule

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) GetOneByGcalID(db *gorm.DB, gcalID string) (*model.Schedule, error) {
	var sch *model.Schedule
	return sch, db.
		Table("schedules").
		Joins("JOIN schedule_google_calendars sgc ON schedules.id = sgc.schedule_id").Where("sgc.google_calendar_id = ?", gcalID).First(&sch).Error
}

func (s *store) Create(db *gorm.DB, schedule *model.Schedule) (*model.Schedule, error) {
	return schedule, db.Create(schedule).Error
}

func (s *store) CreateDiscord(db *gorm.DB, schedule *model.ScheduleDiscordEvent) (*model.ScheduleDiscordEvent, error) {
	return schedule, db.Create(schedule).Error
}

func (s *store) Upsert(db *gorm.DB, schedule *model.Schedule) (*model.Schedule, error) {
	return schedule, db.Save(schedule).Error
}

func (s *store) Update(db *gorm.DB, schedule *model.Schedule) (*model.Schedule, error) {
	return schedule, db.Updates(schedule).Error
}
