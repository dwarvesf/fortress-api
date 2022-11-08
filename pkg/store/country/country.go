package country

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

// All get all countries
func (s *store) All() ([]*model.Country, error) {
	var countries []*model.Country
	return countries, s.db.Find(&countries).Error
}

// One get 1 country by id
func (s *store) One(id string) (*model.Country, error) {
	var country *model.Country
	return country, s.db.Where("id = ?", id).First(&country).Error
}
