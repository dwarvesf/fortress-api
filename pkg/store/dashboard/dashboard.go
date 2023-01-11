package dashboard

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) GetProjectSizes(db *gorm.DB) ([]*model.ProjectSize, error) {
	var ru []*model.ProjectSize

	query := `
		SELECT projects.id, projects.name, projects.code, count(*) AS size
			FROM (projects join project_members pm ON projects.id = pm.project_id)
			WHERE projects.function = 'development' AND pm.status = 'active' 
			GROUP BY projects.id
	`

	return ru, db.Debug().Raw(query).Scan(&ru).Error
}
