package discordaccount

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
}

func New() IStore {
	return &store{}
}

func (r *store) Upsert(db *gorm.DB, da *model.DiscordAccount) (*model.DiscordAccount, error) {
	return da, db.
		Table("discord_accounts").
		Clauses(
			clause.OnConflict{
				Columns: []clause.Column{
					{Name: "discord_id"},
				},
				DoUpdates: clause.Assignments(
					map[string]interface{}{
						"username": da.Username,
					},
				),
			},
		).
		Create(da).
		Error
}

func (r *store) All(db *gorm.DB) ([]*model.DiscordAccount, error) {
	var res []*model.DiscordAccount
	return res, db.Find(&res).Error
}

func (r *store) One(db *gorm.DB, id string) (*model.DiscordAccount, error) {
	res := model.DiscordAccount{}
	return &res, db.Where("id = ?", id).First(&res).Error
}

func (r *store) OneByDiscordID(db *gorm.DB, discordID string) (*model.DiscordAccount, error) {
	res := model.DiscordAccount{}
	return &res, db.Where("discord_id = ?", discordID).First(&res).Error
}

func (r *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.DiscordAccount, updatedFields ...string) (a *model.DiscordAccount, err error) {
	discordAccount := model.DiscordAccount{}
	return &discordAccount, db.Model(&discordAccount).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}
