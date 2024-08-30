package eventspeaker

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Create creates a new e
func (s *store) Create(db *gorm.DB, e *model.EventSpeaker) (*model.EventSpeaker, error) {
	return e, db.Create(e).Error
}

// DeleteAllByEventID deletes all event speakers by event id
func (s *store) DeleteAllByEventID(db *gorm.DB, eventID string) error {
	return db.Table("event_speakers").Where("event_id = ?", eventID).Delete(&model.EventSpeaker{}).Error
}

// List returns a lit of event speaker with loaded event info
func (s *store) List(db *gorm.DB, discordID string, after *time.Time, topic string) ([]model.EventSpeaker, error) {
	var eventSpeakers []model.EventSpeaker
	query := db.Table("event_speakers").
		Joins("JOIN discord_accounts ON event_speakers.discord_account_id = discord_accounts.id").
		Joins("JOIN events ON events.id = event_speakers.event_id").
		Order("events.date DESC")

	if after != nil {
		query = query.Where("events.date > ?", after)
	}

	if topic != "" {
		query = query.Where("LOWER(event_speakers.topic) LIKE LOWER(?)", "%"+topic+"%")
	}

	if discordID != "" {
		query = query.Where("discord_accounts.discord_id = ?", discordID)
	}

	err := query.Preload("Event").
		Find(&eventSpeakers).Error
	if err != nil {
		return nil, err
	}
	return eventSpeakers, nil
}

// Count returns the total count of event speakers with the same filtering criteria as List
func (s *store) Count(db *gorm.DB, discordID string, after *time.Time, topic string) (int64, error) {
	var count int64
	query := db.Table("event_speakers").
		Joins("JOIN discord_accounts ON event_speakers.discord_account_id = discord_accounts.id").
		Joins("JOIN events ON events.id = event_speakers.event_id")

	if after != nil {
		query = query.Where("events.date > ?", after)
	}

	if topic != "" {
		query = query.Where("LOWER(event_speakers.topic) LIKE LOWER(?)", "%"+topic+"%")
	}

	if discordID != "" {
		query = query.Where("discord_accounts.discord_id = ?", discordID)
	}

	err := query.Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// SpeakerStats
type SpeakerStats struct {
	TotalSpeakCount int64 `gorm:"column:total_speak_count"`
	SpeakRank       int64 `gorm:"column:speak_rank"`
}

// GetSpeakerStats returns the total speak count and rank for a given discord_id
func (s *store) GetSpeakerStats(db *gorm.DB, discordID string, after *time.Time, topic string) (SpeakerStats, error) {
	var stats SpeakerStats

	subQuery := db.Table("event_speakers").
		Select("discord_accounts.discord_id, COUNT(event_speakers.topic) as total_speak_count").
		Joins("JOIN discord_accounts ON event_speakers.discord_account_id = discord_accounts.id").
		Joins("JOIN events ON events.id = event_speakers.event_id")

	if after != nil {
		subQuery = subQuery.Where("events.date > ?", after)
	}
	if topic != "" {
		subQuery = subQuery.Where("LOWER(event_speakers.topic) LIKE LOWER(?)", "%"+topic+"%")
	}

	subQuery = subQuery.Group("discord_accounts.discord_id")

	err := db.Table("(?) as subquery", subQuery).
		Select("total_speak_count, RANK() OVER (ORDER BY total_speak_count DESC) as speak_rank").
		Where("discord_id = ?", discordID).
		Scan(&stats).Error

	if err != nil {
		return SpeakerStats{}, err
	}

	return stats, nil
}
