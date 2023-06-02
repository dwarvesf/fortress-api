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
	return res, db.Raw(`SELECT
	Team,
	SUM(Amount) AS Total_Amount
FROM
	vw_icy_earning_by_team_weekly
WHERE
	Period >= to_char(date_trunc('week', CURRENT_DATE), 'yyyy-mm-dd')
	AND Period <= to_char(date_trunc('week', CURRENT_DATE) + '6 days'::interval, 'yyyy-mm-dd')
GROUP BY
	Team
ORDER BY
	Total_Amount DESC;`).Find(&res).Error
}
