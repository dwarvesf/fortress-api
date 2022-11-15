package country

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
}

func New() IStore {
	return &store{}
}

// All get all countries
func (s *store) All(db *gorm.DB) ([]*model.Country, error) {
	var countries []*model.Country
	return countries, db.Find(&countries).Error
}

// One get 1 country by id
func (s *store) One(db *gorm.DB, id string) (*model.Country, error) {
	var country *model.Country
	return country, db.Where("id = ?", id).First(&country).Error
}
