package dashboard

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) GetResourceUtilizationByYear(db *gorm.DB) ([]*model.ResourceUtilization, error) {
	var ru []*model.ResourceUtilization

	query := `
	SELECT "date" ,
		COUNT(DISTINCT(employee_id)) FILTER (
			WHERE deployment_type = 'official' 
				AND status = 'active'
				AND "project_type" != 'dwarves'
				AND joined_date <= "date"
				AND COALESCE(left_date,NOW()) >= "date"
				AND "date" <= NOW()
		) AS official,

		COUNT(DISTINCT(employee_id)) FILTER (
			WHERE status = 'active' 
				AND joined_date <= "date"
				AND COALESCE(left_date,NOW()) >= "date"
				AND "date" <= NOW()
				AND employee_id NOT IN (
					SELECT ru2.employee_id
					FROM resource_utilization ru2
					WHERE ru2.deployment_type = 'official' 
						AND ru2.status = 'active'
						AND ru2."project_type" != 'dwarves'
						AND ru2.joined_date <= ru."date"
						AND COALESCE(ru2.left_date,NOW()) >= ru."date"
						AND ru2."date" <= NOW())
		) AS shadow,
		
		COUNT(DISTINCT(employee_id)) FILTER (
			WHERE project_id IS NULL
				OR employee_id NOT IN (
					SELECT ru3.employee_id
					FROM resource_utilization ru3
					WHERE ru3.status = 'active'
						AND ru3.joined_date <= ru."date" 
						AND COALESCE(ru3.left_date,NOW()) >= ru."date")
		) AS available
	FROM resource_utilization ru 
	GROUP BY "date" 
	`

	return ru, db.Raw(query).Scan(&ru).Error
}
