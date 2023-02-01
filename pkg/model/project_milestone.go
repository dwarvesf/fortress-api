package model

import "time"

type ProjectMilestone struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	StartDate     time.Time           `json:"start_date"`
	EndDate       time.Time           `json:"end_date"`
	SubMilestones []*ProjectMilestone `json:"sub_milestones"`
}

func (o *ProjectMilestone) GetSubMilestones() []*ProjectMilestone {
	return o.SubMilestones
}
