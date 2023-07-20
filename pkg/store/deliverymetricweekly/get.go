package deliverymetricweekly

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

func (s *store) GetLast(db *gorm.DB, num int) ([]model.WeeklyDeliveryMetric, error) {
	var rs []model.WeeklyDeliveryMetric
	return rs, db.Table("vw_weekly_delivery_metrics").Order("date desc").Limit(num).Find(&rs).Error
}

func (s *store) Avg(db *gorm.DB) (model.AvgWeeklyDeliveryMetric, error) {
	var rs model.AvgWeeklyDeliveryMetric
	return rs, db.Raw(`SELECT AVG(sum_weight) AS weight, 
						AVG(sum_effort) AS effort
					FROM vw_weekly_delivery_metrics;`).Scan(&rs).Error
}

func (s *store) AvgWithoutLatestWeek(db *gorm.DB) (model.AvgWeeklyDeliveryMetric, error) {
	var rs model.AvgWeeklyDeliveryMetric
	return rs, db.Raw(`SELECT AVG(sum_weight) AS weight, 
											AVG(sum_effort) AS effort
										FROM vw_weekly_delivery_metrics
										WHERE date != (SELECT MAX(date) FROM delivery_metrics);`).Scan(&rs).Error
}

func (s *store) AvgByMonth(db *gorm.DB, month time.Time) ([]model.AvgMonthWeeklyDeliveryMetric, error) {
	var rs []model.AvgMonthWeeklyDeliveryMetric
	return rs, db.Raw(`SELECT 
											DATE_TRUNC('month', date) AS "month",
											AVG(sum_weight) AS "weight",
											AVG(sum_effort) AS "effort"
										FROM vw_weekly_delivery_metrics
										GROUP BY DATE_TRUNC('month', date)
										HAVING DATE_TRUNC('month', date) = ?
										ORDER BY "month" DESC;`, month).Scan(&rs).Error
}

// Create creates a new delivery metric
func (s *store) Create(db *gorm.DB, e *model.DeliveryMetric) (rs *model.DeliveryMetric, err error) {
	return e, db.Create(e).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.DeliveryMetric, updatedFields ...string) (*model.DeliveryMetric, error) {
	rs := model.DeliveryMetric{}
	return &rs, db.Model(&rs).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}
