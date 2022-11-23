package seniority

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

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

// IsExist check existence of a seniority
func (s *store) IsExist(db *gorm.DB, id string) (bool, error) {
	type res struct {
		Result bool
	}

	result := res{}
	query := db.Raw("SELECT EXISTS (SELECT * FROM seniorities WHERE id = ?) as result", id)

	return result.Result, query.Scan(&result).Error
}
