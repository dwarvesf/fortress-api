package deliverymetric

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// One get delivery metrict by id
func (s *store) One(db *gorm.DB, id string) (*model.DeliveryMetric, error) {
	var rs *model.DeliveryMetric
	return rs, db.Where("id = ?", id).First(&rs).Error
}

// Create creates a new delivery metric
func (s *store) Create(db *gorm.DB, e []model.DeliveryMetric) (rs []model.DeliveryMetric, err error) {
	return e, db.Create(&e).Error
}

func (s *store) GetLatestWeek(db *gorm.DB) (*time.Time, error) {
	var rs *time.Time
	return rs, db.Model(&model.DeliveryMetric{}).Select("date").Order("date DESC").Limit(1).First(&rs).Error
}

func (s *store) GetTopWeighMetrics(db *gorm.DB, w *time.Time, limit int) ([]model.DeliveryMetric, error) {
	var rs []model.DeliveryMetric
	return rs, db.Where("date = ?", w).Order("weight DESC, effectiveness DESC").Limit(limit).Find(&rs).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.DeliveryMetric, updatedFields ...string) (*model.DeliveryMetric, error) {
	rs := model.DeliveryMetric{}
	return &rs, db.Model(&rs).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}

// GetLatest get latest delivery metric by id
func (s *store) GetLatest(db *gorm.DB) (*model.DeliveryMetric, error) {
	var rs *model.DeliveryMetric
	return rs, db.Order("ref DESC").Limit(1).First(&rs).Error
}
