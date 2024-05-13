package communitymember

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type IStore interface {
	// ListByUsernames gets a list of community members by given usernames
	ListByUsernames(db *gorm.DB, usernames []string) ([]model.CommunityMember, error)
	// Insert inserts a community member on conflict discord_id or discord_username do nothing
	Insert(db *gorm.DB, cm *model.CommunityMember) error
}
