package discordaccount

import (
	"errors"

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
						"discord_username": da.DiscordUsername,
						"roles":            da.Roles,
						"github_username":  da.GithubUsername,
						"personal_email":   da.PersonalEmail,
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

// OneByUsername gets a discord account by discord username
// Returns nil, nil if not found (graceful handling)
// Returns nil, error if database error occurs
func (r *store) OneByUsername(db *gorm.DB, username string) (*model.DiscordAccount, error) {
	var res model.DiscordAccount
	err := db.Where("discord_username = ?", username).First(&res).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not found is not an error
		}
		return nil, err
	}
	return &res, nil
}

func (r *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.DiscordAccount, updatedFields ...string) (a *model.DiscordAccount, err error) {
	discordAccount := model.DiscordAccount{}
	return &discordAccount, db.Model(&discordAccount).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}

// ListByMemoUsername gets a list of discord accounts by memo usernames, discord username, or github username (fallback)
func (r *store) ListByMemoUsername(db *gorm.DB, usernames []string) ([]model.DiscordAccount, error) {
	var cms []model.DiscordAccount
	err := db.Where("memo_username IN (?) OR discord_username IN (?) OR github_username IN (?)", usernames, usernames, usernames).Find(&cms).Error
	return cms, err
}
