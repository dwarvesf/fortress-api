package discordevent

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// One get client by id
func (s *store) One(db *gorm.DB, query *Query) (*model.Event, error) {
	var e *model.Event
	if query.ID != "" {
		db = db.Where("id = ?", query.ID)
	}
	if query.DiscordEventID != "" {
		db = db.Where("discord_event_id = ?", query.DiscordEventID)
	}
	return e, db.Preload("EventSpeakers").First(&e).Error
}

// All get all client
func (s *store) All(db *gorm.DB, q *Query, preload bool) ([]*model.Event, error) {
	var e []*model.Event

	if q.Limit == 0 {
		q.Limit = 10
	}

	query := db.Order("date desc").Limit(q.Limit)

	if q.After != nil {
		query = query.Where("date > ?", q.After)
	}

	if len(q.DiscordEventIDs) > 0 {
		query = query.Where("discord_event_id IN (?)", q.DiscordEventIDs)
	}

	if !preload {
		return e, query.Find(&e).Error
	}

	return e, query.Preload("EventSpeakers").Find(&e).Order("date desc").Error
}

// Create creates a new e
func (s *store) Create(db *gorm.DB, e *model.Event) (*model.Event, error) {
	return e, db.Create(e).Error
}

func (s *store) SetSpeakers(db *gorm.DB, e *model.Event) error {
	for _, es := range e.EventSpeakers {
		if es.DiscordAccountID.String() == "" {
			continue
		}
		// upsert speaker
		if err := db.Clauses(
			clause.OnConflict{
				Columns: []clause.Column{
					{Name: "discord_account_id"},
					{Name: "event_id"},
				},
				DoUpdates: clause.Assignments(
					map[string]interface{}{
						"topic": es.Topic,
					},
				),
			},
		).Create(es).Error; err != nil {
			return err
		}
	}
	return nil
}
