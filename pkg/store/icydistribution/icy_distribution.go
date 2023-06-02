package icydistribution

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

// New implements IStore.
func New() IStore {
	return &store{}
}

// GetWeekly implements IStore.
func (*store) GetWeekly(db *gorm.DB) ([]model.IcyDistribution, error) {
	res := []model.IcyDistribution{}
	return res, db.Raw(`SELECT v.period, v.team, v.amount
	FROM vw_icy_earning_by_team_weekly v
	JOIN (
	  SELECT team, MAX(period) AS max_period
	  FROM vw_icy_earning_by_team_weekly
	  GROUP BY team
	) t ON v.team = t.team AND v.period = t.max_period
	ORDER BY v.period DESC;`).Find(&res).Error
}
