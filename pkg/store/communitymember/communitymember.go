package communitymember

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type store struct {
}

// New creates a new store
func New() IStore {
	return &store{}
}

// ListByUsernames gets a list of community members by given usernames
func (s *store) ListByUsernames(db *gorm.DB, usernames []string) ([]model.CommunityMember, error) {
	var cms []model.CommunityMember
	err := db.Where("discord_username IN (?)", usernames).Find(&cms).Error
	return cms, err
}

// Insert inserts a community member on conflict discord_id or discord_username do nothing
func (s *store) Insert(db *gorm.DB, cm *model.CommunityMember) error {
	return db.Table("community_members").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "discord_id"}, {Name: "discord_username"}},
		DoNothing: true,
	}).Create(cm).Error
}
