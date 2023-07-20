package discordaccount

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB) (discordAccounts []*model.DiscordAccount, err error)
	One(db *gorm.DB, id string) (*model.DiscordAccount, error)
	OneByDiscordID(db *gorm.DB, discordID string) (*model.DiscordAccount, error)

	Upsert(db *gorm.DB, da *model.DiscordAccount) (*model.DiscordAccount, error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, client model.DiscordAccount, updatedFields ...string) (a *model.DiscordAccount, err error)
}
