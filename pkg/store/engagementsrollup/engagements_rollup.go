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
						"message_count":  gorm.Expr("engagements_rollup.message_count + excluded.message_count"),
						"reaction_count": gorm.Expr("engagements_rollup.reaction_count + excluded.reaction_count"),
					},
				),
			},
		).
		Create(record).
		Error
}
