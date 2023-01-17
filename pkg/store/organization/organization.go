package organization

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type store struct{}

func New() IStore {
	return &store{}
}
func (s *store) One(db *gorm.DB, id string) (*model.Organization, error) {
	var organization *model.Organization
	return organization, db.Where("id = ?", id).First(&organization).Error
}

func (s *store) All(db *gorm.DB) ([]*model.Organization, error) {
	var organizaitons []*model.Organization
	return organizaitons, db.Find(&organizaitons).Error
}
