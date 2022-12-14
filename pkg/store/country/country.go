package country

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

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

// OneByName get 1 country by name
func (s *store) OneByName(db *gorm.DB, name string) (*model.Country, error) {
	var country *model.Country
	return country, db.Where("name = ?", name).First(&country).Error
}

// IsExist check country existence
func (s *store) IsExist(db *gorm.DB, id string) (bool, error) {
	type res struct {
		Result bool
	}

	result := res{}
	query := db.Raw("SELECT EXISTS (SELECT * FROM countries WHERE id = ?) as result", id)

	return result.Result, query.Scan(&result).Error
}
