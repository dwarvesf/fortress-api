package seniority

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
}

func New() IStore {
	return &store{}
}

// All get all Seniorities
func (s *store) All(db *gorm.DB) ([]*model.Seniority, error) {
	var seniories []*model.Seniority
	return seniories, db.Find(&seniories).Error
}

// One get 1 one by id
func (s *store) One(db *gorm.DB, id model.UUID) (*model.Seniority, error) {
	var sen *model.Seniority
	return sen, db.Where("id = ?", id).First(&sen).Error
}
