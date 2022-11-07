package seniority

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
	db *gorm.DB
}

func New(db *gorm.DB) IStore {
	return &store{
		db: db,
	}
}

// One get all Seniorities
func (s *store) All() ([]*model.Seniority, error) {
	var seniories []*model.Seniority
	return seniories, s.db.Find(&seniories).Error
}
