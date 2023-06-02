package engagementsrollup

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) Upsert(db *gorm.DB, record *model.EngagementsRollup) (*model.EngagementsRollup, error) {
	return record, db.
		Table("engagements_rollup").
		Clauses(
			clause.OnConflict{
				Columns: []clause.Column{
					{Name: "discord_user_id"},
					{Name: "channel_id"},
				},
				DoUpdates: clause.Assignments(
					map[string]interface{}{
						// COALESCE is needed since anything can be null
						"message_count":   gorm.Expr("COALESCE(engagements_rollup.message_count, 0) + COALESCE(excluded.message_count, 0)"),
						"reaction_count":  gorm.Expr("COALESCE(engagements_rollup.reaction_count, 0) + COALESCE(excluded.reaction_count, 0)"),
						"last_message_id": gorm.Expr("MAX(engagements_rollup.last_message_id, excluded.last_message_id)"),
					},
				),
			},
		).
		Create(record).
		Error
}

func (s *store) GetLastMessageID(db *gorm.DB, channelID string) (string, error) {
	lastMessageID := ""
	err := db.
		Raw(
			"SELECT COALESCE(MAX(last_message_id), 0) FROM engagements_rollup WHERE channel_id = ?",
			channelID,
		).
		Scan(&lastMessageID).
		Error
	return lastMessageID, err
}
