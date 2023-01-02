package dashboard

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) GetResourceUtilizationByYear(db *gorm.DB, year int) ([]*model.ResourceUtilization, error) {
	var ru []*model.ResourceUtilization

	query := `
	SELECT d.d AS "date",
			COUNT(*) FILTER (WHERE pm.deployment_type = 'official' AND pm.status = 'active'
							AND pm.joined_date <= d.d
							AND COALESCE(pm.left_date,now()) >= d.d
							AND d.d <= now()) AS official, 
			COUNT(*) FILTER (WHERE pm.deployment_type = 'shadow' AND pm.status = 'active'
							AND pm.joined_date <= d.d
							AND COALESCE(pm.left_date,now()) >= d.d
							AND d.d <= now()) AS shadow,
			COUNT(*) FILTER (WHERE (pm.deployment_type IS NULL
							OR pm.joined_date > d.d
							OR pm.left_date < d.d)
							AND d.d <= now()) AS available
	FROM employees e 
		JOIN project_members pm ON pm.employee_id = e.id, 
		generate_series(
			TO_DATE(?::TEXT, 'YYYY'), 
			TO_DATE(?::TEXT, 'YYYY') + INTERVAL '11 month', 
			'1 month'
		) d
	WHERE e."working_status" = 'full-time'
	group by d.d
	order by d.d
	`

	return ru, db.Debug().Raw(query, year, year).Scan(&ru).Error
}
